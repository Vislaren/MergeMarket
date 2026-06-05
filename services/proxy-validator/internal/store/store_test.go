package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// These tests exercise the real RedisStore against a live Redis. They are
// skipped unless REDIS_TEST_ADDR is set (e.g. "localhost:6379"), so the unit
// suite stays hermetic while CI (A-10), which has Redis available, runs them.
func testStore(t *testing.T) (*RedisStore, *redis.Client) {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR not set; skipping live Redis store test")
	}
	key := "proxy_pool:test"
	s := NewRedis(addr, os.Getenv("REDIS_TEST_PASSWORD"), 0, key)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Ping(ctx); err != nil {
		t.Skipf("Redis not reachable at %s: %v", addr, err)
	}
	raw := redis.NewClient(&redis.Options{Addr: addr, Password: os.Getenv("REDIS_TEST_PASSWORD")})
	t.Cleanup(func() {
		_ = raw.Del(context.Background(), key, key+":staging").Err()
		_ = s.Close()
		_ = raw.Close()
	})
	return s, raw
}

func TestRedisReplacePopulatesPoolWithTTL(t *testing.T) {
	s, raw := testStore(t)
	ctx := context.Background()

	members := []string{"1.2.3.4:8080", "5.6.7.8:3128"}
	if err := s.Replace(ctx, members, 5*time.Minute); err != nil {
		t.Fatalf("Replace error = %v", err)
	}

	card, err := raw.SCard(ctx, s.key).Result()
	if err != nil || card != 2 {
		t.Fatalf("SCard = %d (err %v), want 2", card, err)
	}
	ttl, err := raw.TTL(ctx, s.key).Result()
	if err != nil || ttl <= 0 || ttl > 5*time.Minute {
		t.Fatalf("TTL = %s (err %v), want (0,5m]", ttl, err)
	}
}

func TestRedisReplaceIsAtomicSwap(t *testing.T) {
	s, raw := testStore(t)
	ctx := context.Background()

	if err := s.Replace(ctx, []string{"1.1.1.1:80"}, time.Minute); err != nil {
		t.Fatalf("first Replace error = %v", err)
	}
	if err := s.Replace(ctx, []string{"2.2.2.2:80", "3.3.3.3:80"}, time.Minute); err != nil {
		t.Fatalf("second Replace error = %v", err)
	}
	members, err := raw.SMembers(ctx, s.key).Result()
	if err != nil {
		t.Fatalf("SMembers error = %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("after swap got %d members, want 2: %#v", len(members), members)
	}
	// The staging key must not linger after the rename.
	if exists, _ := raw.Exists(ctx, s.key+":staging").Result(); exists != 0 {
		t.Error("staging key should not exist after Replace")
	}
}

func TestRedisReplaceEmptyClearsPool(t *testing.T) {
	s, raw := testStore(t)
	ctx := context.Background()

	if err := s.Replace(ctx, []string{"1.1.1.1:80"}, time.Minute); err != nil {
		t.Fatalf("seed Replace error = %v", err)
	}
	if err := s.Replace(ctx, nil, time.Minute); err != nil {
		t.Fatalf("empty Replace error = %v", err)
	}
	if exists, _ := raw.Exists(ctx, s.key).Result(); exists != 0 {
		t.Error("pool key should be deleted when no working proxies remain")
	}
}
