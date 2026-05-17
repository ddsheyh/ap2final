package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"user-service/internal/domain"
)

// UserRepository handles all PostgreSQL operations for users.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and returns the created user.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash, name string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, name, is_banned, created_at, updated_at`,
		email, passwordHash, name,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, is_banned, created_at, updated_at FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, is_banned, created_at, updated_at FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

// List returns paginated users.
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, email, password_hash, name, is_banned, created_at, updated_at FROM users ORDER BY id LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.IsBanned, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, total, nil
}

// Update modifies a user's name and email.
func (r *UserRepository) Update(ctx context.Context, id int64, name, email string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(ctx,
		`UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3
		 RETURNING id, email, password_hash, name, is_banned, created_at, updated_at`,
		name, email, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdatePassword changes the user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, newHash string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, newHash, id,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// SetBanned updates the banned status of a user.
func (r *UserRepository) SetBanned(ctx context.Context, id int64, banned bool) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET is_banned = $1, updated_at = NOW() WHERE id = $2`, banned, id,
	)
	if err != nil {
		return fmt.Errorf("set banned: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
