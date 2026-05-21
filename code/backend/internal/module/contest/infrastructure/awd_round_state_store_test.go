package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
)

func TestAWDRoundStateStoreReplaceAWDRoundFlagIfMatchReplacesWhenCurrentMatches(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	store := contestinfra.NewAWDRoundStateStore(client)
	ctx := context.Background()
	roundKey := rediskeys.AWDRoundFlagsKey(12, 121)
	field := rediskeys.AWDRoundFlagServiceField(1212, 12001)
	if err := client.HSet(ctx, roundKey, map[string]any{
		field: "awd{old-flag}",
	}).Err(); err != nil {
		t.Fatalf("seed round flag: %v", err)
	}

	replaced, err := store.ReplaceAWDRoundFlagIfMatch(ctx, 12, 121, 1212, 12001, "awd{old-flag}", "awd{new-flag}", 30*time.Second)
	if err != nil {
		t.Fatalf("ReplaceAWDRoundFlagIfMatch() error = %v", err)
	}
	if !replaced {
		t.Fatal("expected flag replacement to succeed")
	}

	got, err := client.HGet(ctx, roundKey, field).Result()
	if err != nil {
		t.Fatalf("load replaced flag: %v", err)
	}
	if got != "awd{new-flag}" {
		t.Fatalf("unexpected replaced flag: got %q want %q", got, "awd{new-flag}")
	}
}

func TestAWDRoundStateStoreReplaceAWDRoundFlagIfMatchRejectsStaleCurrentValue(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	store := contestinfra.NewAWDRoundStateStore(client)
	ctx := context.Background()
	roundKey := rediskeys.AWDRoundFlagsKey(13, 131)
	field := rediskeys.AWDRoundFlagServiceField(1312, 13001)
	if err := client.HSet(ctx, roundKey, map[string]any{
		field: "awd{claimed-flag}",
	}).Err(); err != nil {
		t.Fatalf("seed round flag: %v", err)
	}

	replaced, err := store.ReplaceAWDRoundFlagIfMatch(ctx, 13, 131, 1312, 13001, "awd{stale-flag}", "awd{new-flag}", 30*time.Second)
	if err != nil {
		t.Fatalf("ReplaceAWDRoundFlagIfMatch() error = %v", err)
	}
	if replaced {
		t.Fatal("expected stale replacement to be rejected")
	}

	got, err := client.HGet(ctx, roundKey, field).Result()
	if err != nil {
		t.Fatalf("load preserved flag: %v", err)
	}
	if got != "awd{claimed-flag}" {
		t.Fatalf("unexpected preserved flag: got %q want %q", got, "awd{claimed-flag}")
	}
}
