package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	"ctf-platform/internal/module/ops/infrastructure/cachekeys"
	ctfws "ctf-platform/internal/websocket"
)

type recordingContestRealtimeBroadcaster struct {
	channels []string
	users    []int64
	messages []ctfws.Envelope
}

func (b *recordingContestRealtimeBroadcaster) SendToChannel(channel string, message ctfws.Envelope) int {
	b.channels = append(b.channels, channel)
	b.messages = append(b.messages, message)
	return 1
}

func (b *recordingContestRealtimeBroadcaster) SendToUser(userID int64, message ctfws.Envelope) int {
	b.users = append(b.users, userID)
	b.messages = append(b.messages, message)
	return 1
}

func TestContestRealtimeStreamPublishAndConsume(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	cache := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = cache.Close() })

	broadcaster := &recordingContestRealtimeBroadcaster{}
	stream := opsinfra.NewContestRealtimeStream(cache, broadcaster, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := cache.Set(ctx, cachekeys.ContestRealtimeCursorKey+"test-consumer", "0-0", 0).Err(); err != nil {
		t.Fatalf("seed cursor error = %v", err)
	}

	if _, err := stream.Publish(ctx, contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventScoreboardUpdated,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.ScoreboardChannel(42),
		MessageType: "scoreboard.updated",
		Payload:     contestcontracts.ScoreboardUpdatedRelayPayload{ContestID: 42},
		Timestamp:   time.Now().UTC(),
	}, ""); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if err := stream.ConsumeOnce(ctx, "test-consumer"); err != nil {
		t.Fatalf("ConsumeOnce() error = %v", err)
	}

	if len(broadcaster.channels) != 1 || broadcaster.channels[0] != contestcontracts.ScoreboardChannel(42) {
		t.Fatalf("unexpected channel fanout: %+v", broadcaster.channels)
	}
	if len(broadcaster.messages) != 1 || broadcaster.messages[0].Type != "scoreboard.updated" {
		t.Fatalf("unexpected message fanout: %+v", broadcaster.messages)
	}
}

func TestContestRealtimeStreamConsumeOnceInitializesCursorFromTailWithoutReplay(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	cache := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = cache.Close() })

	broadcaster := &recordingContestRealtimeBroadcaster{}
	stream := opsinfra.NewContestRealtimeStream(cache, broadcaster, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := stream.Publish(ctx, contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventAnnouncementDeleted,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.AnnouncementChannel(77),
		MessageType: "contest.announcement.deleted",
		Payload: contestcontracts.AnnouncementDeletedRelayPayload{
			ContestID:      77,
			AnnouncementID: 501,
		},
		Timestamp: time.Now().UTC(),
	}, "contest:77:announcement:501:deleted"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if err := stream.ConsumeOnce(ctx, "instance-a"); err != nil {
		t.Fatalf("ConsumeOnce() error = %v", err)
	}
	if len(broadcaster.messages) != 0 {
		t.Fatalf("expected first consumer bootstrap to skip history, got %+v", broadcaster.messages)
	}
	cursor, err := cache.Get(ctx, cachekeys.ContestRealtimeCursorKey+"instance-a").Result()
	if err != nil || cursor == "" {
		t.Fatalf("expected bootstrap cursor persisted, got cursor=%q err=%v", cursor, err)
	}

	if _, err := stream.Publish(ctx, contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventAnnouncementDeleted,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.AnnouncementChannel(77),
		MessageType: "contest.announcement.deleted",
		Payload: contestcontracts.AnnouncementDeletedRelayPayload{
			ContestID:      77,
			AnnouncementID: 502,
		},
		Timestamp: time.Now().UTC(),
	}, "contest:77:announcement:502:deleted"); err != nil {
		t.Fatalf("Publish(second) error = %v", err)
	}

	if err := stream.ConsumeOnce(ctx, "instance-a"); err != nil {
		t.Fatalf("ConsumeOnce(second) error = %v", err)
	}

	if len(broadcaster.channels) != 1 || broadcaster.channels[0] != contestcontracts.AnnouncementChannel(77) {
		t.Fatalf("unexpected channel fanout after bootstrap: %+v", broadcaster.channels)
	}
	if len(broadcaster.messages) != 1 || broadcaster.messages[0].Type != "contest.announcement.deleted" {
		t.Fatalf("unexpected message fanout after bootstrap: %+v", broadcaster.messages)
	}
}

func TestContestRealtimeStreamConsumeOnceRecoversPendingMessageForSameConsumer(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	cache := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = cache.Close() })

	broadcaster := &recordingContestRealtimeBroadcaster{}
	stream := opsinfra.NewContestRealtimeStream(cache, broadcaster, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := cache.Set(ctx, cachekeys.ContestRealtimeCursorKey+"instance-a", "0-0", 0).Err(); err != nil {
		t.Fatalf("seed cursor error = %v", err)
	}
	if _, err := stream.Publish(ctx, contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventAnnouncementDeleted,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.AnnouncementChannel(77),
		MessageType: "contest.announcement.deleted",
		Payload: contestcontracts.AnnouncementDeletedRelayPayload{
			ContestID:      77,
			AnnouncementID: 501,
		},
		Timestamp: time.Now().UTC(),
	}, "contest:77:announcement:501:deleted"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if err := stream.ConsumeOnce(ctx, "instance-a"); err != nil {
		t.Fatalf("ConsumeOnce() error = %v", err)
	}

	if len(broadcaster.channels) != 1 || broadcaster.channels[0] != contestcontracts.AnnouncementChannel(77) {
		t.Fatalf("unexpected recovered channel fanout: %+v", broadcaster.channels)
	}
	if len(broadcaster.messages) != 1 || broadcaster.messages[0].Type != "contest.announcement.deleted" {
		t.Fatalf("unexpected recovered message fanout: %+v", broadcaster.messages)
	}
}

func TestContestRealtimeStreamPublishIsIdempotentByDedupeKey(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	cache := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = cache.Close() })

	stream := opsinfra.NewContestRealtimeStream(cache, &recordingContestRealtimeBroadcaster{}, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	relay := contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventScoreboardUpdated,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.ScoreboardChannel(42),
		MessageType: "scoreboard.updated",
		Payload:     contestcontracts.ScoreboardUpdatedRelayPayload{ContestID: 42},
		Timestamp:   time.Now().UTC(),
	}

	firstID, err := stream.Publish(ctx, relay, "contest:42:scoreboard:updated")
	if err != nil {
		t.Fatalf("Publish(first) error = %v", err)
	}
	secondID, err := stream.Publish(ctx, relay, "contest:42:scoreboard:updated")
	if err != nil {
		t.Fatalf("Publish(second) error = %v", err)
	}

	if firstID == "" || secondID == "" || firstID != secondID {
		t.Fatalf("expected same message id for idempotent publish, got first=%q second=%q", firstID, secondID)
	}

	messages, err := cache.XRange(ctx, cachekeys.ContestRealtimeStreamKey, "-", "+").Result()
	if err != nil {
		t.Fatalf("XRange() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected single stream message after duplicate publish, got %d", len(messages))
	}
}
