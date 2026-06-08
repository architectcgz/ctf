package composition

import (
	"context"
	"errors"
	"sync"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/platform/clustersecret"
	"ctf-platform/internal/platform/events"
)

type Root struct {
	Events events.Bus

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

func NewLoopBackgroundJob(name string, run func(context.Context)) BackgroundJob {
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
	appCtx, appCancel := context.WithCancel(context.Background())
	if err := registerContainerFlagSecret(appCtx, cfg, db); err != nil {
		appCancel()
		return nil, err
	}
	return &Root{
		Events:    events.NewBus(),
		appCtx:    appCtx,
		appCancel: appCancel,
		cfg:       cfg,
		log:       log,
		db:        db,
		cache:     cache,
	}, nil
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
