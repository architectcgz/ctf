package composition

import (
	"context"
	"errors"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/platform/clustersecret"
	"ctf-platform/internal/platform/events"
)

type Root struct {
	Events                          events.Bus
	outboxHandlers                  *events.OutboxHandlerRegistry
	eventStream                     *events.StreamFanout
	practiceOutboxHandlerRegistrars []func(*events.OutboxHandlerRegistry)

	jobsMu sync.Mutex
	jobs   []BackgroundJob

	appCtx    context.Context
	appCancel context.CancelFunc
	cfg       *config.Config
	log       *zap.Logger
	db        *gorm.DB
	cache     *redislib.Client
}

type BackgroundJob struct {
	name  string
	start func(context.Context) error
	stop  func(context.Context) error
}

func NewBackgroundJob(name string, start func(context.Context) error, stop func(context.Context) error) BackgroundJob {
	return BackgroundJob{name: name, start: start, stop: stop}
}

func NewLoopBackgroundJobWithLogger(name string, logger *zap.Logger, run func(context.Context)) BackgroundJob {
	var (
		mu      sync.Mutex
		cancel  context.CancelFunc
		started bool
		wg      sync.WaitGroup
	)

	return NewBackgroundJob(
		name,
		func(ctx context.Context) error {
			if ctx == nil {
				return errors.New("background job start requires context")
			}
			mu.Lock()
			defer mu.Unlock()
			if started {
				return nil
			}
			started = true

			runCtx, runCancel := context.WithCancel(ctx)
			cancel = runCancel
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					if recovered := recover(); recovered != nil {
						if logger != nil {
							logger.Error("background_job_panicked",
								zap.Any("panic", recovered),
								zap.String("task_name", name),
								zap.ByteString("stack", debug.Stack()),
							)
						}
						panic(recovered)
					}
				}()
				run(runCtx)
			}()
			return nil
		},
		func(ctx context.Context) error {
			mu.Lock()
			if !started {
				mu.Unlock()
				return nil
			}
			stopFn := cancel
			mu.Unlock()

			if stopFn != nil {
				stopFn()
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	)
}

func BuildRoot(cfg *config.Config, log *zap.Logger, db *gorm.DB, cache *redislib.Client) (*Root, error) {
	if log == nil {
		log = zap.NewNop()
	}
	appCtx, appCancel := context.WithCancel(context.Background())
	if err := registerContainerFlagSecret(appCtx, cfg, db); err != nil {
		appCancel()
		return nil, err
	}
	root := &Root{
		Events:         events.NewBus(),
		outboxHandlers: events.NewOutboxHandlerRegistry(),
		eventStream:    events.NewStreamFanout(cache, events.StreamFanoutOptions{}, log.Named("platform_event_stream")),
		appCtx:         appCtx,
		appCancel:      appCancel,
		cfg:            cfg,
		log:            log,
		db:             db,
		cache:          cache,
	}
	root.registerPlatformEventJobs()
	return root, nil
}

func registerContainerFlagSecret(ctx context.Context, cfg *config.Config, db *gorm.DB) error {
	if cfg == nil || db == nil || cfg.Container.FlagGlobalSecret == "" {
		return nil
	}
	keyID := cfg.Container.ResolvedFlagSecretKeyID
	if keyID == "" {
		keyID = cfg.Container.FlagGlobalSecretKeyID
	}
	if keyID == "" {
		keyID = "default"
	}
	secrets := cfg.Container.ResolvedFlagSecrets
	if secrets == nil {
		secrets = map[string]string{keyID: cfg.Container.FlagGlobalSecret}
	}
	requiredKeyIDs, err := clustersecret.RequiredContainerFlagSecretKeyIDs(ctx, db)
	if err != nil {
		return err
	}
	return clustersecret.RegisterContainerFlagSecretKeyring(ctx, db, clustersecret.ContainerFlagSecretKeyring{
		ActiveKeyID:    keyID,
		ActiveSecret:   cfg.Container.FlagGlobalSecret,
		Secrets:        secrets,
		RequiredKeyIDs: requiredKeyIDs,
		AllowRotation:  cfg.Container.FlagGlobalSecretAllowRotation,
	})
}

func (r *Root) Context() context.Context {
	if r == nil {
		return nil
	}
	return r.appCtx
}

func (r *Root) Cancel() {
	if r == nil || r.appCancel == nil {
		return
	}
	r.appCancel()
}

func (r *Root) Config() *config.Config {
	return r.cfg
}

func (r *Root) Logger() *zap.Logger {
	return r.log
}

func (r *Root) DB() *gorm.DB {
	return r.db
}

func (r *Root) Cache() *redislib.Client {
	return r.cache
}

func (r *Root) OutboxHandlerRegistry() *events.OutboxHandlerRegistry {
	if r == nil {
		return nil
	}
	return r.outboxHandlers
}

func (r *Root) RegisterOutboxHandler(name string, handler events.OutboxHandler) {
	if r == nil || r.outboxHandlers == nil {
		return
	}
	r.outboxHandlers.Register(name, handler)
}

func (r *Root) addPracticeOutboxHandlerRegistrar(registrar func(*events.OutboxHandlerRegistry)) {
	if r == nil || registrar == nil {
		return
	}
	r.practiceOutboxHandlerRegistrars = append(r.practiceOutboxHandlerRegistrars, registrar)
}

func (r *Root) registerPracticeOutboxHandlerRegistrars() {
	if r == nil || r.outboxHandlers == nil {
		return
	}
	for _, registrar := range r.practiceOutboxHandlerRegistrars {
		if registrar != nil {
			registrar(r.outboxHandlers)
		}
	}
}

func (r *Root) RegisterBackgroundJob(job BackgroundJob) {
	if r == nil || job.name == "" {
		return
	}
	r.jobsMu.Lock()
	defer r.jobsMu.Unlock()
	r.jobs = append(r.jobs, job)
}

func (r *Root) BackgroundJobs() []BackgroundJob {
	if r == nil {
		return nil
	}
	r.jobsMu.Lock()
	defer r.jobsMu.Unlock()
	jobs := make([]BackgroundJob, len(r.jobs))
	copy(jobs, r.jobs)
	return jobs
}

func (j BackgroundJob) Name() string {
	return j.name
}

func (j BackgroundJob) Start(ctx context.Context) error {
	if j.start == nil {
		return nil
	}
	return j.start(ctx)
}

func (j BackgroundJob) Stop(ctx context.Context) error {
	if j.stop == nil {
		return nil
	}
	return j.stop(ctx)
}

func (r *Root) registerPlatformEventJobs() {
	if r == nil || r.db == nil || r.cache == nil {
		return
	}
	dispatcher := events.NewOutboxDispatcher(
		events.NewOutboxRepository(r.db),
		r.eventStream,
		r.outboxHandlers,
		r.log.Named("platform_event_outbox_dispatcher"),
	)
	r.RegisterBackgroundJob(NewLoopBackgroundJobWithLogger("platform_event_outbox_dispatcher", r.log.Named("platform_event_outbox_dispatcher"), func(ctx context.Context) {
		dispatcher.Run(ctx, platformEventWorkerID("dispatcher"))
	}))
	r.RegisterBackgroundJob(NewLoopBackgroundJobWithLogger("platform_event_stream_consumer", r.log.Named("platform_event_stream_consumer"), func(ctx context.Context) {
		runPlatformEventStreamConsumer(ctx, r.eventStream, r.outboxHandlers, platformEventWorkerID("consumer"), r.log.Named("platform_event_stream_consumer"))
	}))
}

func runPlatformEventStreamConsumer(ctx context.Context, stream *events.StreamFanout, handlers *events.OutboxHandlerRegistry, consumerID string, logger *zap.Logger) {
	if stream == nil || handlers == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := stream.ConsumeOnce(ctx, consumerID, handlers.Handle); err != nil {
			if logger != nil {
				logger.Warn("consume platform event stream failed", zap.Error(err))
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

func platformEventWorkerID(kind string) string {
	hostname, err := os.Hostname()
	if err != nil {
		return "platform-event-" + kind
	}
	name := strings.TrimSpace(hostname)
	if name == "" {
		return "platform-event-" + kind
	}
	return "platform-event-" + kind + "-" + name
}
