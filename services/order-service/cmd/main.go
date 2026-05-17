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

	grpchandler "order-service/internal/handler/grpc"
	natspkg "order-service/internal/nats"
	"order-service/internal/repository"
	"order-service/internal/usecase"
	pb "order-service/pkg/pb"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting order-service...")

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

	orderRepo := repository.NewOrderRepository(pool)
	redisCache := repository.NewRedisCache(rdb)

	var publisher *natspkg.Publisher
	if nc != nil {
		publisher = natspkg.NewPublisher(nc)
	}

	orderUsecase := usecase.NewOrderUsecase(orderRepo, redisCache, publisher)
	orderHandler := grpchandler.NewOrderHandler(orderUsecase)

	if nc != nil {
		subscriber := natspkg.NewSubscriber(nc, orderRepo)
		if err := subscriber.Subscribe(); err != nil {
			log.Printf("WARNING: NATS subscribe failed: %v", err)
		}
	}

	grpcPort := getEnv("GRPC_PORT", "50052")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterOrderServiceServer(server, orderHandler)
	reflection.Register(server)

	go func() {
		log.Printf("order-service gRPC listening on :%s", grpcPort)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down order-service...")
	server.GracefulStop()
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("Running migrations...")
	queries := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			total_price NUMERIC(12,2) DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL DEFAULT 1,
			price NUMERIC(12,2) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id)`,
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
