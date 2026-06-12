package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const streamFanoutMessageField = "event"

var streamFanoutPublishScript = redislib.NewScript(`
local marker = redis.call("GET", KEYS[2])
if marker then
	return marker
end
local messageID = redis.call("XADD", KEYS[1], "*", ARGV[1], ARGV[2])
redis.call("SET", KEYS[2], messageID)
return messageID
`)

type StreamFanoutOptions struct {
	StreamKey        string
	CursorKeyPrefix  string
	PublishKeyPrefix string
}

type StreamFanout struct {
	cache  *redislib.Client
	opts   StreamFanoutOptions
	logger *zap.Logger
}

func NewStreamFanout(cache *redislib.Client, opts StreamFanoutOptions, logger *zap.Logger) *StreamFanout {
	if logger == nil {
		logger = zap.NewNop()
	}
	if opts.StreamKey == "" {
		opts.StreamKey = "platform:events:stream"
	}
	if opts.CursorKeyPrefix == "" {
		opts.CursorKeyPrefix = "platform:events:cursor:"
	}
	if opts.PublishKeyPrefix == "" {
		opts.PublishKeyPrefix = "platform:events:publish:"
	}
	return &StreamFanout{cache: cache, opts: opts, logger: logger}
}

func (s *StreamFanout) Publish(ctx context.Context, event OutboxEvent) (string, error) {
	if s == nil || s.cache == nil {
		return "", fmt.Errorf("stream fanout requires redis client")
	}
	payload, err := json.Marshal(streamFanoutEnvelopeFromEvent(event))
	if err != nil {
		return "", fmt.Errorf("encode stream fanout event: %w", err)
	}
	if event.DedupeKey == "" {
		return s.cache.XAdd(ctx, &redislib.XAddArgs{
			Stream: s.opts.StreamKey,
			Values: map[string]any{streamFanoutMessageField: string(payload)},
		}).Result()
	}
	return streamFanoutPublishScript.Run(
		ctx,
		s.cache,
		[]string{s.opts.StreamKey, s.publishMarkerKey(event.DedupeKey)},
		streamFanoutMessageField,
		string(payload),
	).Text()
}

func (s *StreamFanout) ConsumeOnce(ctx context.Context, consumer string, handler func(context.Context, OutboxEvent) error) error {
	if s == nil || s.cache == nil {
		return fmt.Errorf("stream fanout requires redis client")
	}
	if handler == nil {
		return fmt.Errorf("stream fanout handler is required")
	}
	if consumer == "" {
		consumer = "platform-event-consumer"
	}
	streams, err := s.readFromCursor(ctx, consumer)
	if err != nil {
		return err
	}
	for _, stream := range streams {
		for _, message := range stream.Messages {
			event, err := decodeStreamFanoutMessage(message)
			if err != nil {
				return err
			}
			if err := handler(ctx, event); err != nil {
				return err
			}
			if err := s.cache.Set(ctx, s.cursorKey(consumer), message.ID, 0).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *StreamFanout) readFromCursor(ctx context.Context, consumer string) ([]redislib.XStream, error) {
	cursor, err := s.cache.Get(ctx, s.cursorKey(consumer)).Result()
	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return s.initializeCursor(ctx, consumer)
		}
		return nil, err
	}
	return s.readFromID(ctx, cursor)
}

func (s *StreamFanout) initializeCursor(ctx context.Context, consumer string) ([]redislib.XStream, error) {
	latest, err := s.cache.XRevRangeN(ctx, s.opts.StreamKey, "+", "-", 1).Result()
	if err != nil {
		return nil, err
	}
	if len(latest) == 0 {
		if err := s.cache.Set(ctx, s.cursorKey(consumer), "0-0", 0).Err(); err != nil {
			return nil, err
		}
		return s.readFromID(ctx, "0-0")
	}
	if err := s.cache.Set(ctx, s.cursorKey(consumer), latest[0].ID, 0).Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *StreamFanout) readFromID(ctx context.Context, cursor string) ([]redislib.XStream, error) {
	streams, err := s.cache.XRead(ctx, &redislib.XReadArgs{
		Streams: []string{s.opts.StreamKey, cursor},
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

func (s *StreamFanout) cursorKey(consumer string) string {
	return s.opts.CursorKeyPrefix + consumer
}

func (s *StreamFanout) publishMarkerKey(dedupeKey string) string {
	return s.opts.PublishKeyPrefix + dedupeKey
}

type streamFanoutEnvelope struct {
	Name           string    `json:"name"`
	PayloadVersion int       `json:"payload_version"`
	Payload        string    `json:"payload"`
	Route          string    `json:"route"`
	DedupeKey      string    `json:"dedupe_key,omitempty"`
	OccurredAt     time.Time `json:"occurred_at"`
}

func streamFanoutEnvelopeFromEvent(event OutboxEvent) streamFanoutEnvelope {
	return streamFanoutEnvelope{
		Name:           event.Name,
		PayloadVersion: event.PayloadVersion,
		Payload:        string(event.Payload),
		Route:          event.Route,
		DedupeKey:      event.DedupeKey,
		OccurredAt:     normalizeOutboxTime(event.OccurredAt),
	}
}

func (e streamFanoutEnvelope) toEvent() OutboxEvent {
	return OutboxEvent{
		Name:           e.Name,
		PayloadVersion: e.PayloadVersion,
		Payload:        []byte(e.Payload),
		Route:          e.Route,
		DedupeKey:      e.DedupeKey,
		OccurredAt:     normalizeOutboxTime(e.OccurredAt),
	}
}

func decodeStreamFanoutMessage(message redislib.XMessage) (OutboxEvent, error) {
	raw, ok := message.Values[streamFanoutMessageField]
	if !ok {
		return OutboxEvent{}, fmt.Errorf("stream fanout message missing %s", streamFanoutMessageField)
	}
	text, ok := raw.(string)
	if !ok {
		return OutboxEvent{}, fmt.Errorf("unexpected stream fanout payload type: %T", raw)
	}
	var envelope streamFanoutEnvelope
	if err := json.Unmarshal([]byte(text), &envelope); err != nil {
		return OutboxEvent{}, fmt.Errorf("decode stream fanout payload: %w", err)
	}
	if envelope.Payload == "" {
		if legacyPayload, ok := message.Values["payload"]; ok {
			envelope.Payload = fmt.Sprint(legacyPayload)
		}
	}
	if envelope.PayloadVersion == 0 {
		if rawVersion, ok := message.Values["payload_version"]; ok {
			version, _ := strconv.Atoi(fmt.Sprint(rawVersion))
			envelope.PayloadVersion = version
		}
	}
	return envelope.toEvent(), nil
}

var _ StreamPublisher = (*StreamFanout)(nil)
