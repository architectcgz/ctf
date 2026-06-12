package runtime

import (
	"context"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
)

func registerContestRealtimeConsumers(deps moduleDeps) []BackgroundJob {
	stream := opsinfra.NewContestRealtimeStream(deps.input.Cache, deps.webSocketManager, deps.input.Logger.Named("contest_realtime_stream"))
	relayService := opscmd.NewContestRealtimeService(stream)
	relayService.RegisterContestEventConsumers(deps.input.Events)

	dispatcher := opscmd.NewContestRealtimeOutboxDispatcher(deps.contestRealtimeOutbox, stream, deps.input.Logger.Named("contest_realtime_outbox_dispatcher"))
	consumerID := contestRealtimeConsumerID()
	return []BackgroundJob{
		{Name: "contest_realtime_outbox_dispatcher", Run: dispatcher.Run},
		{
			Name: "contest_realtime_stream_consumer",
			Run: func(ctx context.Context) {
				runContestRealtimeConsumer(ctx, stream, consumerID, deps.input.Logger.Named("contest_realtime_stream_consumer"))
			},
		},
	}
}

func runContestRealtimeConsumer(ctx context.Context, stream interface {
	ConsumeOnce(context.Context, string) error
}, consumerID string, logger *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := stream.ConsumeOnce(ctx, consumerID); err != nil {
			if logger != nil {
				logger.Warn("consume contest realtime stream failed", zap.Error(err))
			}
			timer := time.NewTimer(250 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

func contestRealtimeConsumerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "contest-realtime-consumer"
	}
	name := strings.TrimSpace(hostname)
	if name == "" {
		return "contest-realtime-consumer"
	}
	return name
}
