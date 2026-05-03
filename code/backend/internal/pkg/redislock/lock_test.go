package redislock

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
)

func TestLockRefreshExtendsOwnedLock(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	lock, acquired, err := Acquire(context.Background(), client, "lock:refresh:owned", 60*time.Millisecond)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if !acquired {
		t.Fatal("expected lock acquisition")
	}

	refreshed, err := lock.Refresh(context.Background(), 60*time.Millisecond)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if !refreshed {
		t.Fatal("expected refresh to succeed")
	}

	mini.FastForward(50 * time.Millisecond)
	if !mini.Exists("lock:refresh:owned") {
		t.Fatal("expected refreshed lock to still exist after original ttl window")
	}
}

func TestLockRefreshDoesNotExtendForeignLock(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	client := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	lock, acquired, err := Acquire(context.Background(), client, "lock:refresh:foreign", 60*time.Millisecond)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if !acquired {
		t.Fatal("expected lock acquisition")
	}

	if err := mini.Set("lock:refresh:foreign", "another-token"); err != nil {
		t.Fatalf("seed foreign token: %v", err)
	}
	mini.SetTTL("lock:refresh:foreign", 10*time.Millisecond)

	refreshed, err := lock.Refresh(context.Background(), 60*time.Millisecond)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed {
		t.Fatal("expected refresh to fail for foreign token")
	}

	mini.FastForward(20 * time.Millisecond)
	if mini.Exists("lock:refresh:foreign") {
		t.Fatal("expected foreign lock to expire without refresh")
	}
}
