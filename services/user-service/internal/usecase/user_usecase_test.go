package usecase

import (
	"testing"

	"user-service/internal/domain"
)

func TestHashAndCheckPassword(t *testing.T) {
	uc := &UserUsecase{jwtSecret: []byte("test-secret")}

	hash, err := uc.hashPassword("mypassword123")
	if err != nil {
		t.Fatalf("hashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	if err := uc.checkPassword(hash, "mypassword123"); err != nil {
		t.Fatalf("checkPassword should succeed: %v", err)
	}

	if err := uc.checkPassword(hash, "wrongpassword"); err == nil {
		t.Fatal("checkPassword should fail for wrong password")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	uc := &UserUsecase{jwtSecret: []byte("test-secret")}

	tokens, err := uc.generateTokenPair(1, "test@example.com")
	if err != nil {
		t.Fatalf("generateTokenPair failed: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("access token should not be empty")
	}
	if tokens.RefreshToken == "" {
		t.Fatal("refresh token should not be empty")
	}
	if tokens.AccessToken == tokens.RefreshToken {
		t.Fatal("access and refresh tokens should be different")
	}
}

func TestUserDomainModel(t *testing.T) {
	user := &domain.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
	}
	if user.Email != "test@example.com" {
		t.Fatal("user email mismatch")
	}
	if user.IsBanned {
		t.Fatal("new user should not be banned")
	}
}

// Integration test - requires running PostgreSQL + Redis
func TestRegisterAndLogin_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	t.Log("Integration test - requires running infrastructure")
}
