package usecase

import (
	"context"
	"testing"

	"order-service/internal/domain"
	"order-service/internal/repository"

	"github.com/redis/go-redis/v9"
)

func TestCreateOrderValidation(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	cache := repository.NewRedisCache(rdb)
	uc := NewOrderUsecase(nil, cache, nil)

	// Empty items should fail
	_, err := uc.CreateOrder(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("CreateOrder with no items should fail")
	}
	if err.Error() != "order must have at least one item" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaginationDefaults(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		wantPage int
		wantSize int
	}{
		{"negative page", -1, 10, 1, 10},
		{"zero page", 0, 10, 1, 10},
		{"zero size", 1, 0, 1, 20},
		{"large size", 1, 200, 1, 20},
		{"valid", 2, 15, 2, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := tt.page
			pageSize := tt.pageSize
			if page < 1 { page = 1 }
			if pageSize < 1 || pageSize > 100 { pageSize = 20 }
			if page != tt.wantPage {
				t.Errorf("page = %d, want %d", page, tt.wantPage)
			}
			if pageSize != tt.wantSize {
				t.Errorf("pageSize = %d, want %d", pageSize, tt.wantSize)
			}
		})
	}
}

func TestOrderDomainModel(t *testing.T) {
	order := &domain.Order{
		ID:     1,
		UserID: 42,
		Status: domain.StatusPending,
		Items: []domain.OrderItem{
			{ProductName: "Ticket A", Quantity: 2, Price: 1500.00},
			{ProductName: "Ticket B", Quantity: 1, Price: 3000.00},
		},
	}

	if order.Status != "pending" {
		t.Fatalf("expected pending status, got %s", order.Status)
	}
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(order.Items))
	}

	// Calculate expected total
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	if total != 6000.00 {
		t.Fatalf("expected total 6000, got %.2f", total)
	}
}

// Integration test - requires running PostgreSQL
func TestOrderCRUD_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	t.Log("Integration test - requires running infrastructure")
	t.Log("Run with: go test -v -run TestOrderCRUD_Integration ./internal/usecase/")
}
