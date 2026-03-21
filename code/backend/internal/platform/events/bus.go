package events

import (
	"context"
	"errors"
	"sync"
)

type Event struct {
	Name    string
	Payload any
}

type Handler func(ctx context.Context, evt Event) error

type Bus interface {
	Subscribe(name string, fn Handler)
	Publish(ctx context.Context, evt Event) error
}

type inMemoryBus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
}

func NewBus() Bus {
	return &inMemoryBus{
		subscribers: make(map[string][]Handler),
	}
}

func (b *inMemoryBus) Subscribe(name string, fn Handler) {
	if fn == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[name] = append(b.subscribers[name], fn)
}

func (b *inMemoryBus) Publish(ctx context.Context, evt Event) error {
	if ctx == nil {
		ctx = context.Background()
	}

	b.mu.RLock()
	handlers := append([]Handler(nil), b.subscribers[evt.Name]...)
	b.mu.RUnlock()

	var errs []error
	for _, handler := range handlers {
		if err := handler(ctx, evt); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
