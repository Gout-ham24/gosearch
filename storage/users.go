package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents an account in the database.
type User struct {
	ID           int
	Email        string
	PasswordHash string
}

// CreateUser inserts a new user and returns their ID.
// Returns an error if the email already exists.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, email, passwordHash string) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash,
	).Scan(&id)
	return id, err
}

// GetUserByEmail looks up a user by email. Returns nil if not found.
func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*User, error) {
	var u User
	err := pool.QueryRow(ctx,
		`SELECT id, email, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByID looks up a user by ID. Returns nil if not found.
func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id int) (*User, error) {
	var u User
	err := pool.QueryRow(ctx,
		`SELECT id, email, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// StoreRefreshToken saves a new refresh token for a user, valid for 30 days.
func StoreRefreshToken(ctx context.Context, pool *pgxpool.Pool, userID int, token string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, time.Now().Add(30*24*time.Hour),
	)
	return err
}

// ValidateRefreshToken checks if a refresh token is valid and not expired, returning the user_id if so.
func ValidateRefreshToken(ctx context.Context, pool *pgxpool.Pool, token string) (int, error) {
	var userID int
	err := pool.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens WHERE token = $1 AND expires_at > now()`,
		token,
	).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return userID, err
}

// DeleteRefreshToken removes a single refresh token (used for normal logout).
func DeleteRefreshToken(ctx context.Context, pool *pgxpool.Pool, token string) error {
	_, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}

// DeleteAllUserRefreshTokens removes ALL refresh tokens for a user (used for "logout everywhere").
func DeleteAllUserRefreshTokens(ctx context.Context, pool *pgxpool.Pool, userID int) error {
	_, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

// DeleteUser permanently deletes a user account (and cascades to their tokens/history).
func DeleteUser(ctx context.Context, pool *pgxpool.Pool, userID int) error {
	_, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return err
}

// UpdatePassword changes a user's password hash.
func UpdatePassword(ctx context.Context, pool *pgxpool.Pool, userID int, newHash string) error {
	_, err := pool.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, newHash, userID)
	return err
}
