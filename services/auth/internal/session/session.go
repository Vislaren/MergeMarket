// Package session stores refresh sessions in Redis with encrypted payloads.
package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Vislaren/MergeMarket/services/auth/internal/secure"
)

// ErrNotFound indicates that a refresh token has no active Redis session.
var ErrNotFound = errors.New("session: not found")

// Data is the encrypted JSON stored in Redis.
type Data struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store persists refresh sessions.
type Store interface {
	Create(ctx context.Context, data Data) (string, error)
	Consume(ctx context.Context, refreshToken string) (Data, error)
}

// Redis stores sessions as AES-256-GCM encrypted JSON under session:{hash}.
type Redis struct {
	client          *redis.Client
	sessionTTL      time.Duration
	refreshTokenTTL time.Duration
	key             []byte
	now             func() time.Time
}

// NewRedis creates a Redis-backed encrypted session store.
func NewRedis(client *redis.Client, sessionTTL, refreshTokenTTL time.Duration, key []byte) (*Redis, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("session: encryption key must be 32 bytes")
	}
	return &Redis{client: client, sessionTTL: sessionTTL, refreshTokenTTL: refreshTokenTTL, key: key, now: time.Now}, nil
}

// Create stores a new encrypted session and returns its opaque refresh token.
func (r *Redis) Create(ctx context.Context, data Data) (string, error) {
	refresh, err := randomToken()
	if err != nil {
		return "", err
	}
	now := r.now().UTC()
	data.CreatedAt = now
	data.ExpiresAt = now.Add(r.refreshTokenTTL)
	raw, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("session: marshal: %w", err)
	}
	encrypted, err := secure.EncryptString(r.key, string(raw))
	if err != nil {
		return "", err
	}
	if err := r.client.Set(ctx, redisKey(refresh), encrypted, r.sessionTTL).Err(); err != nil {
		return "", fmt.Errorf("session: redis set: %w", err)
	}
	return refresh, nil
}

// Consume atomically loads and removes a refresh session.
func (r *Redis) Consume(ctx context.Context, refreshToken string) (Data, error) {
	encrypted, err := r.client.GetDel(ctx, redisKey(refreshToken)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Data{}, ErrNotFound
		}
		return Data{}, fmt.Errorf("session: redis getdel: %w", err)
	}
	plain, err := secure.DecryptString(r.key, encrypted)
	if err != nil {
		return Data{}, err
	}
	var data Data
	if err := json.Unmarshal([]byte(plain), &data); err != nil {
		return Data{}, fmt.Errorf("session: unmarshal: %w", err)
	}
	if r.now().After(data.ExpiresAt) {
		return Data{}, ErrNotFound
	}
	return data, nil
}

func redisKey(refreshToken string) string {
	return "session:" + secure.SHA256Base64(refreshToken)
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", fmt.Errorf("session: random refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
