package events

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type StreamPublisher interface {
	Publish(ctx context.Context, event OutboxEvent) (string, error)
}

type OutboxHandler func(ctx context.Context, event OutboxEvent) error

type OutboxHandlerRegistry struct {
	handlers map[string][]OutboxHandler
}

func NewOutboxHandlerRegistry() *OutboxHandlerRegistry {
	return &OutboxHandlerRegistry{handlers: make(map[string][]OutboxHandler)}
}

func (r *OutboxHandlerRegistry) Register(name string, handler OutboxHandler) {
	if r == nil || name == "" || handler == nil {
		return
	}
	r.handlers[name] = append(r.handlers[name], handler)
}

func (r *OutboxHandlerRegistry) Handle(ctx context.Context, event OutboxEvent) error {
	if r == nil {
		return fmt.Errorf("outbox handler registry is not configured")
	}
	handlers := r.handlers[event.Name]
	if len(handlers) == 0 {
		return fmt.Errorf("outbox handler not registered for %s", event.Name)
	}
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

type OutboxDispatcher struct {
	store        OutboxStore
	stream       StreamPublisher
	handlers     *OutboxHandlerRegistry
	logger       *zap.Logger
	batchSize    int
	lease        time.Duration
	retryBackoff time.Duration
}

func NewOutboxDispatcher(store OutboxStore, stream StreamPublisher, handlers *OutboxHandlerRegistry, logger *zap.Logger) *OutboxDispatcher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &OutboxDispatcher{
		store:        store,
		stream:       stream,
		handlers:     handlers,
		logger:       logger,
		batchSize:    32,
		lease:        time.Minute,
		retryBackoff: 5 * time.Second,
	}
}

func (d *OutboxDispatcher) Run(ctx context.Context, workerID string) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := d.DispatchOnce(ctx, workerID); err != nil && d.logger != nil {
			d.logger.Warn("dispatch platform event outbox failed", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *OutboxDispatcher) DispatchOnce(ctx context.Context, workerID string) error {
	if d == nil || d.store == nil {
		return nil
	}
	if workerID == "" {
		workerID = "platform-event-dispatcher"
	}
	now := time.Now().UTC()
	items, err := d.store.ClaimDue(ctx, workerID, now, d.lease, d.batchSize)
	if err != nil {
		return err
	}
	for _, item := range items {
		streamMessageID, err := d.dispatch(ctx, item.Event)
		if err != nil {
			if markErr := d.store.MarkFailed(ctx, item.ID, err, now.Add(d.retryBackoff)); markErr != nil {
				return markErr
			}
			continue
		}
		if err := d.store.MarkDispatched(ctx, item.ID, streamMessageID, now); err != nil {
			return err
		}
	}
	return nil
}

func (d *OutboxDispatcher) dispatch(ctx context.Context, event OutboxEvent) (string, error) {
	switch event.Route {
	case OutboxRouteStream:
		if d.stream == nil {
			return "", fmt.Errorf("stream publisher is not configured")
		}
		return d.stream.Publish(ctx, event)
	case OutboxRouteHandler, "":
		if err := d.handlers.Handle(ctx, event); err != nil {
			return "", err
		}
		return "", nil
	default:
		return "", fmt.Errorf("unsupported outbox route %s", event.Route)
	}
}
