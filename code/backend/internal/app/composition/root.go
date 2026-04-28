package composition

import (
	"context"
	"sync"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/platform/events"
)

type Root struct {
	Events events.Bus

	jobsMu sync.Mutex
	jobs   []BackgroundJob

	cfg   *config.Config
	log   *zap.Logger
	db    *gorm.DB
	cache *redislib.Client
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
		func(context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			if started {
				return nil
			}
			started = true

			runCtx, runCancel := context.WithCancel(context.Background())
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
	return &Root{
		Events: events.NewBus(),
		cfg:    cfg,
		log:    log,
		db:     db,
		cache:  cache,
	}, nil
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
