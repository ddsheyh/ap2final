package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"user-service/internal/domain"
	"user-service/internal/email"
	natspub "user-service/internal/nats"
	"user-service/internal/repository"
)

type UserUsecase struct {
	repo      *repository.UserRepository
	cache     *repository.RedisCache
	publisher *natspub.Publisher
	mailer    *email.SMTPSender
	jwtSecret []byte
}

func NewUserUsecase(
	repo *repository.UserRepository,
	cache *repository.RedisCache,
	publisher *natspub.Publisher,
	mailer *email.SMTPSender,
	jwtSecret string,
) *UserUsecase {
	return &UserUsecase{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
		mailer:    mailer,
		jwtSecret: []byte(jwtSecret),
	}
}

func (uc *UserUsecase) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (uc *UserUsecase) checkPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (uc *UserUsecase) generateTokenPair(userID int64, email string) (*domain.TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"type":    "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(uc.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"type":    "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(uc.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &domain.TokenPair{AccessToken: accessStr, RefreshToken: refreshStr}, nil
}

func (uc *UserUsecase) Register(ctx context.Context, emailAddr, password, name string) (*domain.User, *domain.TokenPair, error) {
	hash, err := uc.hashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	user, err := uc.repo.Create(ctx, emailAddr, hash, name)
	if err != nil {
		return nil, nil, fmt.Errorf("register: %w", err)
	}

	tokens, err := uc.generateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	_ = uc.cache.StoreRefreshToken(ctx, tokens.RefreshToken, user.ID, 7*24*time.Hour)

	_ = uc.cache.SetUser(ctx, user)

	if uc.mailer != nil {
		go func() {
			if err := uc.mailer.SendWelcomeEmail(emailAddr, name); err != nil {
				log.Printf("Failed to send welcome email: %v", err)
			}
		}()
	}

	if uc.publisher != nil {
		_ = uc.publisher.PublishUserRegistered(user.ID, user.Email, user.Name)
	}

	return user, tokens, nil
}

func (uc *UserUsecase) Login(ctx context.Context, emailAddr, password string) (*domain.User, *domain.TokenPair, error) {
	user, err := uc.repo.GetByEmail(ctx, emailAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	if user.IsBanned {
		return nil, nil, fmt.Errorf("account is banned")
	}

	if err := uc.checkPassword(user.PasswordHash, password); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	tokens, err := uc.generateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	_ = uc.cache.StoreRefreshToken(ctx, tokens.RefreshToken, user.ID, 7*24*time.Hour)

	return user, tokens, nil
}

func (uc *UserUsecase) Logout(ctx context.Context, accessToken string) error {
	return uc.cache.BlacklistToken(ctx, accessToken, 15*time.Minute)
}

func (uc *UserUsecase) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	userID, err := uc.cache.GetRefreshTokenUserID(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	_ = uc.cache.DeleteRefreshToken(ctx, refreshToken)

	tokens, err := uc.generateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.StoreRefreshToken(ctx, tokens.RefreshToken, user.ID, 7*24*time.Hour)

	return tokens, nil
}

func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	if cached, err := uc.cache.GetUser(ctx, id); err == nil {
		return cached, nil
	}

	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.SetUser(ctx, user)
	return user, nil
}

func (uc *UserUsecase) GetUserByEmail(ctx context.Context, emailAddr string) (*domain.User, error) {
	return uc.repo.GetByEmail(ctx, emailAddr)
}

func (uc *UserUsecase) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return uc.repo.List(ctx, offset, pageSize)
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, id int64, name, emailAddr string) (*domain.User, error) {
	user, err := uc.repo.Update(ctx, id, name, emailAddr)
	if err != nil {
		return nil, err
	}
	_ = uc.cache.DeleteUser(ctx, id)
	return user, nil
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, id int64) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = uc.cache.DeleteUser(ctx, id)

	if uc.publisher != nil {
		_ = uc.publisher.PublishUserDeleted(id)
	}
	return nil
}

func (uc *UserUsecase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := uc.checkPassword(user.PasswordHash, oldPassword); err != nil {
		return fmt.Errorf("incorrect old password")
	}

	newHash, err := uc.hashPassword(newPassword)
	if err != nil {
		return err
	}

	return uc.repo.UpdatePassword(ctx, userID, newHash)
}

func (uc *UserUsecase) BanUser(ctx context.Context, userID int64) error {
	if err := uc.repo.SetBanned(ctx, userID, true); err != nil {
		return err
	}
	_ = uc.cache.DeleteUser(ctx, userID)
	return nil
}

func (uc *UserUsecase) UnbanUser(ctx context.Context, userID int64) error {
	if err := uc.repo.SetBanned(ctx, userID, false); err != nil {
		return err
	}
	_ = uc.cache.DeleteUser(ctx, userID)
	return nil
}
