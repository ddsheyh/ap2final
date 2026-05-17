package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"payment-service/internal/email"
	grpchandler "payment-service/internal/handler/grpc"
	natspkg "payment-service/internal/nats"
	"payment-service/internal/repository"
	"payment-service/internal/usecase"
	pb "payment-service/pkg/pb"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting payment-service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USER", "goticket"),
		getEnv("DB_PASSWORD", "goticket123"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "goticket"),
	)

	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 30; i++ {
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			err = pool.Ping(ctx)
		}
		if err == nil {
			break
		}
		log.Printf("Waiting for database... attempt %d/30: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("FATAL: cannot connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	runMigrations(ctx, pool)

	rdb := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Redis not available: %v", err)
	} else {
		log.Println("Connected to Redis")
	}
	defer rdb.Close()

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	var nc *nats.Conn
	for i := 0; i < 30; i++ {
		nc, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		log.Printf("Waiting for NATS... attempt %d/30: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Printf("WARNING: NATS not available: %v", err)
	} else {
		log.Println("Connected to NATS")
		defer nc.Close()
	}

	paymentRepo := repository.NewPaymentRepository(pool)
	redisCache := repository.NewRedisCache(rdb)

	var publisher *natspkg.Publisher
	if nc != nil {
		publisher = natspkg.NewPublisher(nc)
	}

	var mailer *email.SMTPSender
	smtpHost := getEnv("SMTP_HOST", "")
	if smtpHost != "" {
		mailer = email.NewSMTPSender(
			smtpHost,
			getEnv("SMTP_PORT", "587"),
			getEnv("SMTP_USER", ""),
			getEnv("SMTP_PASSWORD", ""),
			getEnv("SMTP_FROM", ""),
		)
	}

	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo, redisCache, publisher, mailer)
	paymentHandler := grpchandler.NewPaymentHandler(paymentUsecase)

	if nc != nil {
		subscriber := natspkg.NewSubscriber(nc, paymentRepo)
		if err := subscriber.Subscribe(); err != nil {
			log.Printf("WARNING: NATS subscribe failed: %v", err)
		}
	}

	grpcPort := getEnv("GRPC_PORT", "50053")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterPaymentServiceServer(server, paymentHandler)
	reflection.Register(server)

	go func() {
		log.Printf("payment-service gRPC listening on :%s", grpcPort)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down payment-service...")
	server.GracefulStop()
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("Running migrations...")
	queries := []string{
		`CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			order_id INT NOT NULL,
			user_id INT NOT NULL,
			amount NUMERIC(12,2) NOT NULL,
			currency VARCHAR(10) DEFAULT 'KZT',
			status VARCHAR(50) DEFAULT 'pending',
			payment_method VARCHAR(50) DEFAULT 'card',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			payment_id INT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			amount NUMERIC(12,2) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			description TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_payment_id ON transactions(payment_id)`,
	}
	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			log.Printf("Migration warning: %v", err)
		}
	}
	log.Println("Migrations completed")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
