package composition

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimemodule "ctf-platform/internal/module/runtime/runtime"
	"go.uber.org/zap"
)

type RuntimeModule struct {
	Handler *runtimehttp.Handler

	PracticeInstanceRepository interface {
		FindByID(ctx context.Context, id int64) (*model.Instance, error)
		UpdateRuntime(ctx context.Context, instance *model.Instance) error
		FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
		RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
		UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
		FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error)
		ListPendingInstances(ctx context.Context, limit int) ([]*model.Instance, error)
		TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error)
		CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error)
	}
	PracticeRuntimeService  practiceports.RuntimeInstanceService
	ChallengeImageRuntime   challengeports.ImageRuntime
	ChallengeRuntimeProbe   challengeports.ChallengeRuntimeProbe
	OpsRuntimeQuery         opsports.RuntimeQuery
	OpsRuntimeStatsProvider opsports.RuntimeStatsProvider
	ContestContainerFiles   contestports.AWDContainerFileWriter

	runtime *runtimemodule.Module
}

func BuildRuntimeModule(root *Root) *RuntimeModule {
	cfg := root.Config()
	log := root.Logger()
	engine := buildRuntimeEngine(root)
	module := runtimemodule.Build(runtimemodule.Deps{
		Config: cfg,
		Logger: log,
		DB:     root.DB(),
		Cache:  root.Cache(),
		Engine: engine,
	})

	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewBackgroundJob(job.Name, job.Start, job.Stop))
	}
	if cfg.Container.DefenseSSHEnabled && module.ProxyTicketService != nil && module.ProxyTicketReader != nil && module.SSHExecutor != nil {
		gateway := NewAWDDefenseSSHGateway(
			module.ProxyTicketService,
			module.ProxyTicketReader,
			module.SSHExecutor,
			cfg.Container.DefenseSSHPort,
			log.Named("awd_defense_ssh_gateway"),
		)
		root.RegisterBackgroundJob(NewBackgroundJob(
			"awd_defense_ssh_gateway",
			gateway.Start,
			gateway.Stop,
		))
	}

	return &RuntimeModule{
		PracticeInstanceRepository: module.PracticeInstanceRepository,
		PracticeRuntimeService:     module.PracticeRuntimeService,
		ChallengeImageRuntime:      module.ChallengeImageRuntime,
		ChallengeRuntimeProbe:      module.ChallengeRuntimeProbe,
		OpsRuntimeQuery:            module.OpsRuntimeQuery,
		OpsRuntimeStatsProvider:    module.OpsRuntimeStatsProvider,
		ContestContainerFiles:      module.ContestContainerFiles,
		runtime:                    module,
	}
}

func (m *RuntimeModule) BuildHandler(root *Root, ops *OpsModule) {
	if m == nil || m.runtime == nil {
		return
	}

	cfg := root.Config()
	m.runtime.BuildHandler(ops.AuditService, runtimehttp.CookieConfig{
		Secure:   cfg.Auth.SessionCookieSecure,
		SameSite: cfg.Auth.CookieSameSite(),
	})
	m.Handler = m.runtime.Handler
}

func buildRuntimeEngine(root *Root) runtimemodule.Engine {
	if root == nil {
		return nil
	}

	cfg := root.Config()
	log := root.Logger()
	if cfg == nil {
		return nil
	}
	if cfg.App.Env == "test" {
		if log != nil {
			log.Info("runtime_engine_enabled_with_test_adapter_for_router")
		}
		return newTestRuntimeEngine(log.Named("runtime_test_engine"))
	}

	engine, err := runtimeinfra.NewEngine(&cfg.Container)
	if err != nil {
		if log != nil {
			log.Warn("runtime_engine_init_failed_for_router", zap.Error(err))
		}
		return nil
	}
	return engine
}
