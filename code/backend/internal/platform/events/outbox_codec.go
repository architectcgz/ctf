package events

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type OutboxCodec struct {
	mu        sync.RWMutex
	factories map[outboxCodecKey]func() any
}

type outboxCodecKey struct {
	name    string
	version int
}

type DecodedOutboxEvent struct {
	Event   OutboxEvent
	Payload any
}

func NewOutboxCodec() *OutboxCodec {
	return &OutboxCodec{factories: make(map[outboxCodecKey]func() any)}
}

func (c *OutboxCodec) Register(name string, version int, factory func() any) {
	if c == nil || name == "" || version <= 0 || factory == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[outboxCodecKey{name: name, version: version}] = factory
}

func (c *OutboxCodec) Encode(name string, version int, payload any, occurredAt time.Time) (OutboxEvent, error) {
	if name == "" {
		return OutboxEvent{}, fmt.Errorf("outbox event name is required")
	}
	if version <= 0 {
		return OutboxEvent{}, fmt.Errorf("outbox payload version is required")
	}
	normalized := normalizeTimeFields(payload)
	data, err := json.Marshal(normalized)
	if err != nil {
		return OutboxEvent{}, fmt.Errorf("encode outbox payload %s v%d: %w", name, version, err)
	}
	return OutboxEvent{
		Name:           name,
		PayloadVersion: version,
		Payload:        data,
		OccurredAt:     normalizeOutboxTime(occurredAt),
	}, nil
}

func (c *OutboxCodec) Decode(event OutboxEvent) (DecodedOutboxEvent, error) {
	if c == nil {
		return DecodedOutboxEvent{}, fmt.Errorf("outbox codec is nil")
	}
	c.mu.RLock()
	factory := c.factories[outboxCodecKey{name: event.Name, version: event.PayloadVersion}]
	c.mu.RUnlock()
	if factory == nil {
		return DecodedOutboxEvent{}, fmt.Errorf("unknown outbox event %s v%d", event.Name, event.PayloadVersion)
	}
	payload := factory()
	if err := json.Unmarshal(event.Payload, payload); err != nil {
		return DecodedOutboxEvent{}, fmt.Errorf("decode outbox payload %s v%d: %w", event.Name, event.PayloadVersion, err)
	}
	normalizeTimeFieldsInPlace(reflect.ValueOf(payload))
	event.OccurredAt = normalizeOutboxTime(event.OccurredAt)
	return DecodedOutboxEvent{Event: event, Payload: payload}, nil
}

func normalizeTimeFields(value any) any {
	if value == nil {
		return nil
	}
	source := reflect.ValueOf(value)
	copied := reflect.New(source.Type()).Elem()
	copied.Set(source)
	normalizeTimeFieldsInPlace(copied)
	return copied.Interface()
}

func normalizeTimeFieldsInPlace(value reflect.Value) {
	if !value.IsValid() {
		return
	}
	if value.Kind() == reflect.Interface && !value.IsNil() {
		normalizeTimeFieldsInPlace(value.Elem())
		return
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		normalizeTimeFieldsInPlace(value.Elem())
		return
	}
	if value.Type() == reflect.TypeOf(time.Time{}) {
		if value.CanSet() {
			value.Set(reflect.ValueOf(normalizeOutboxTime(value.Interface().(time.Time))))
		}
		return
	}
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			if field.CanSet() || field.Kind() == reflect.Pointer || field.Kind() == reflect.Struct {
				normalizeTimeFieldsInPlace(field)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			normalizeTimeFieldsInPlace(value.Index(i))
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			item := value.MapIndex(key)
			if !item.IsValid() {
				continue
			}
			copied := reflect.New(item.Type()).Elem()
			copied.Set(item)
			normalizeTimeFieldsInPlace(copied)
			value.SetMapIndex(key, copied)
		}
	}
}
