package storage

import (
	"fmt"
	"testing"
	"time"
)

// helper to create a redis service for tests
func newTestRedis(t *testing.T) *RedisService {
	t.Helper()
	svc, err := NewRedis()
	if err != nil {
		t.Fatalf("failed to create redis client: %v", err)
	}
	return svc
}

func TestRedisServiceLifecycle(t *testing.T) {
	svc := newTestRedis(t)

	key := fmt.Sprintf("test:redis:lifecycle:%d", time.Now().UnixNano())
	expected := true

	if err := svc.Set(key, expected, time.Minute); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	var got bool
	ok, err := svc.Get(key, &got)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatalf("Get returned ok=false for existing key")
	}
	if got != expected {
		t.Fatalf("Get returned %+v, want %+v", got, expected)
	}

	if err := svc.Delete(key); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	ok, err = svc.Get(key, &got)
	if err != nil {
		t.Fatalf("Get after delete returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false after delete")
	}
}

func TestRedisServiceExpiration(t *testing.T) {
	svc := newTestRedis(t)

	key := fmt.Sprintf("test:redis:ttl:%d", time.Now().UnixNano())
	payload := false

	if err := svc.Set(key, payload, 25*time.Millisecond); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	var got bool
	ok, err := svc.Get(key, &got)
	if err != nil {
		t.Fatalf("Get after expiration returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false after expiration")
	}
}
