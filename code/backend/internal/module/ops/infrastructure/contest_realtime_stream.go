package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"ctf-platform/internal/module/ops/infrastructure/cachekeys"
	ctfws "ctf-platform/internal/websocket"
)

var contestRealtimePublishScript = redislib.NewScript(`
local marker = redis.call("GET", KEYS[2])
if marker then
	return marker
end
local messageID = redis.call("XADD", KEYS[1], "*", ARGV[1], ARGV[2], ARGV[3], ARGV[4])
redis.call("SET", KEYS[2], messageID)
return messageID
`)

type contestRealtimeStreamBroadcaster interface {
	SendToChannel(channel string, message ctfws.Envelope) int
	SendToUser(userID int64, message ctfws.Envelope) int
}

type ContestRealtimeStream struct {
	cache       *redislib.Client
	broadcaster contestRealtimeStreamBroadcaster
	logger      *zap.Logger
}

func NewContestRealtimeStream(cache *redislib.Client, broadcaster contestRealtimeStreamBroadcaster, logger *zap.Logger) *ContestRealtimeStream {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ContestRealtimeStream{
		cache:       cache,
		broadcaster: broadcaster,
		logger:      logger,
	}
}

func (s *ContestRealtimeStream) Publish(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) (string, error) {
	if s == nil || s.cache == nil {
		return "", errors.New("contest realtime stream requires redis client")
	}
	payload, err := contestcontracts.EncodeRealtimeRelay(relay)
	if err != nil {
		return "", fmt.Errorf("marshal realtime relay: %w", err)
	}
	if dedupeKey == "" {
		return s.cache.XAdd(ctx, &redislib.XAddArgs{
			Stream: cachekeys.ContestRealtimeStreamKey,
			Values: map[string]any{
				cachekeys.ContestRealtimeMessageKey: string(payload),
			},
		}).Result()
	}
	messageID, err := contestRealtimePublishScript.Run(
		ctx,
		s.cache,
		[]string{cachekeys.ContestRealtimeStreamKey, contestRealtimePublishMarkerKey(dedupeKey)},
		cachekeys.ContestRealtimeMessageKey,
		string(payload),
		cachekeys.ContestRealtimeDedupeKeyFieldName,
		dedupeKey,
	).Text()
	if err != nil {
		return "", err
	}
	return messageID, nil
}

func (s *ContestRealtimeStream) ConsumeOnce(ctx context.Context, consumer string) error {
	if s == nil || s.cache == nil {
		return errors.New("contest realtime stream requires redis client")
	}
	if s.broadcaster == nil {
		return errors.New("contest realtime stream requires broadcaster")
	}
	if consumer == "" {
		consumer = "contest-realtime-consumer"
	}
	streams, err := s.readFromCursor(ctx, consumer)
	if err != nil {
		return err
	}
	for _, stream := range streams {
		for _, message := range stream.Messages {
			relay, err := decodeStreamRelay(message)
			if err != nil {
				return err
			}
			if err := s.fanout(relay); err != nil {
				return err
			}
			if err := s.cache.Set(ctx, contestRealtimeCursorKey(consumer), message.ID, 0).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ContestRealtimeStream) readFromCursor(ctx context.Context, consumer string) ([]redislib.XStream, error) {
	lastID, err := s.cache.Get(ctx, contestRealtimeCursorKey(consumer)).Result()
	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return s.initializeCursor(ctx, consumer)
		} else {
			return nil, err
		}
	}
	return s.readFromID(ctx, lastID)
}

func (s *ContestRealtimeStream) initializeCursor(ctx context.Context, consumer string) ([]redislib.XStream, error) {
	latest, err := s.cache.XRevRangeN(ctx, cachekeys.ContestRealtimeStreamKey, "+", "-", 1).Result()
	if err != nil {
		return nil, err
	}
	if len(latest) == 0 {
		if err := s.cache.Set(ctx, contestRealtimeCursorKey(consumer), "0-0", 0).Err(); err != nil {
			return nil, err
		}
		return s.readFromID(ctx, "0-0")
	}
	if err := s.cache.Set(ctx, contestRealtimeCursorKey(consumer), latest[0].ID, 0).Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *ContestRealtimeStream) readFromID(ctx context.Context, lastID string) ([]redislib.XStream, error) {
	streams, err := s.cache.XRead(ctx, &redislib.XReadArgs{
		Streams: []string{cachekeys.ContestRealtimeStreamKey, lastID},
		Count:   1,
		Block:   100 * time.Millisecond,
	}).Result()
	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return streams, nil
}

func decodeStreamRelay(message redislib.XMessage) (contestcontracts.RealtimeRelayEvent, error) {
	raw, ok := message.Values[cachekeys.ContestRealtimeMessageKey]
	if !ok {
		return contestcontracts.RealtimeRelayEvent{}, fmt.Errorf("stream message missing %s", cachekeys.ContestRealtimeMessageKey)
	}
	text, ok := raw.(string)
	if !ok {
		return contestcontracts.RealtimeRelayEvent{}, fmt.Errorf("unexpected stream relay type: %T", raw)
	}
	relay, err := contestcontracts.DecodeRealtimeRelay([]byte(text))
	if err != nil {
		return contestcontracts.RealtimeRelayEvent{}, fmt.Errorf("decode stream relay: %w", err)
	}
	return relay, nil
}

func (s *ContestRealtimeStream) fanout(relay contestcontracts.RealtimeRelayEvent) error {
	message := ctfws.Envelope{
		Type:      relay.MessageType,
		Payload:   relay.Payload,
		Timestamp: contestRealtimeRelayTimestamp(relay.Timestamp),
	}
	switch relay.Delivery {
	case contestcontracts.RealtimeDeliveryChannel:
		s.broadcaster.SendToChannel(relay.Channel, message)
		return nil
	case contestcontracts.RealtimeDeliveryUser:
		if relay.RecipientUserID == nil {
			return errors.New("user delivery missing recipient_user_id")
		}
		s.broadcaster.SendToUser(*relay.RecipientUserID, message)
		return nil
	default:
		return fmt.Errorf("unsupported realtime delivery: %s", relay.Delivery)
	}
}

func contestRealtimeRelayTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}

func contestRealtimeCursorKey(consumer string) string {
	return cachekeys.ContestRealtimeCursorKey + consumer
}

func contestRealtimePublishMarkerKey(dedupeKey string) string {
	return cachekeys.ContestRealtimePublishMarkerKey + dedupeKey
}
