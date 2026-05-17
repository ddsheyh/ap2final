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

	"user-service/internal/email"
	grpchandler "user-service/internal/handler/grpc"
	natspub "user-service/internal/nats"
	"user-service/internal/repository"
	"user-service/internal/usecase"
	pb "user-service/pkg/pb"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting user-service...")

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

	userRepo := repository.NewUserRepository(pool)
	redisCache := repository.NewRedisCache(rdb)

	var publisher *natspub.Publisher
	if nc != nil {
		publisher = natspub.NewPublisher(nc)
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

	userUsecase := usecase.NewUserUsecase(userRepo, redisCache, publisher, mailer, getEnv("JWT_SECRET", "secret"))
	userHandler := grpchandler.NewUserHandler(userUsecase)

	grpcPort := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, userHandler)
	reflection.Register(server)

	go func() {
		log.Printf("user-service gRPC listening on :%s", grpcPort)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down user-service...")
	server.GracefulStop()
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("Running migrations...")
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			is_banned BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Migration warning: %v", err)
	} else {
		log.Println("Migrations completed")
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
