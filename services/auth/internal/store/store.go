// Package store persists and looks up auth users in PostgreSQL.
package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ErrEmailExists indicates a duplicate registration attempt.
var ErrEmailExists = errors.New("store: email exists")

// ErrNotFound indicates no matching user exists.
var ErrNotFound = errors.New("store: user not found")

// User is the auth-service view of a user record.
type User struct {
	ID           string
	Email        string
	PasswordHash string
}

// Repository defines the user persistence surface required by auth.
type Repository interface {
	CreateUser(ctx context.Context, email, password string) (User, error)
	UserByEmail(ctx context.Context, email string) (User, error)
}

// Postgres implements Repository using pgxpool.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres opens a database pool and verifies connectivity.
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping postgres: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

// Close closes the PostgreSQL pool.
func (p *Postgres) Close() {
	p.pool.Close()
}

// CreateUser inserts a new user with a bcrypt password hash.
func (p *Postgres) CreateUser(ctx context.Context, email, password string) (User, error) {
	normalized := normalizeEmail(email)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("store: hash password: %w", err)
	}
	var user User
	err = p.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id::text, email, password_hash
	`, normalized, string(hash)).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailExists
		}
		return User{}, fmt.Errorf("store: insert user: %w", err)
	}
	return user, nil
}

// UserByEmail loads a user by normalized email.
func (p *Postgres) UserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := p.pool.QueryRow(ctx, `
		SELECT id::text, email, password_hash
		FROM users
		WHERE email = $1
	`, normalizeEmail(email)).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("store: select user: %w", err)
	}
	return user, nil
}

// CheckPassword verifies a plaintext password against the stored bcrypt hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
