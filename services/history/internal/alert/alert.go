// Package alert defines the price-drop alert event and how it is published. When
// the heartbeat detects a product's price crossing down through an alert
// threshold, it emits an Event onto a Redis list for the Notification Worker
// (ARCHITECTURE §1) to deliver via Firebase/APNs. The Publisher interface keeps
// the runner testable without Redis.
package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Event is one fired price-drop alert.
type Event struct {
	AlertID     string    `json:"alert_id"`
	ProductID   string    `json:"product_id"`
	Title       string    `json:"title"`
	Store       string    `json:"store"`
	Price       float64   `json:"price"`
	Threshold   float64   `json:"threshold_price"`
	Currency    string    `json:"currency"`
	TriggeredAt time.Time `json:"triggered_at"`
}

// Publisher delivers fired alert events.
type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

// RedisPublisher pushes events onto a Redis list with RPUSH.
type RedisPublisher struct {
	client *redis.Client
	key    string
}

// NewRedisPublisher builds a RedisPublisher writing to key on client.
func NewRedisPublisher(client *redis.Client, key string) *RedisPublisher {
	return &RedisPublisher{client: client, key: key}
}

// Publish marshals the event and appends it to the alert queue.
func (p *RedisPublisher) Publish(ctx context.Context, e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("alert: encode event: %w", err)
	}
	if err := p.client.RPush(ctx, p.key, data).Err(); err != nil {
		return fmt.Errorf("alert: rpush %s: %w", p.key, err)
	}
	return nil
}
