package commands

import (
	"context"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

func TestChallengeServiceCreateChallengeTreatsModuleImageNotFoundAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		&challengeCommandContextRepoStub{},
		&challengeCommandImageRepoStub{
			findByIDFn: func(context.Context, int64) (*challengeentity.Image, error) {
				return nil, challengeports.ErrChallengeImageNotFound
			},
		},
		&challengeCommandTopologyRepoStub{},
		zap.NewNop(),
	)

	_, err := service.CreateChallenge(context.Background(), 1001, CreateChallengeInput{ImageID: 9})
	if err == nil || err.Error() != apperror.ErrNotFound.Error() {
		t.Fatalf("expected image not found error, got %v", err)
	}
}

func TestChallengeServiceUpdateChallengeTreatsModuleChallengeNotFoundAsErrChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*challengeentity.Challenge, error) {
				return nil, challengeports.ErrChallengeCommandChallengeNotFound
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{},
		zap.NewNop(),
	)

	err := service.UpdateChallenge(context.Background(), 9, UpdateChallengeInput{Title: "updated"})
	if err == nil || err.Error() != challengecontracts.ErrChallengeNotFound.Error() {
		t.Fatalf("expected challenge not found error, got %v", err)
	}
}

func TestChallengeServiceUpdateChallengeTreatsTopologySentinelAsMissingTopology(t *testing.T) {
	t.Parallel()

	updated := false
	service := NewChallengeService(
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*challengeentity.Challenge, error) {
				return &challengeentity.Challenge{
					ID:              9,
					Title:           "shared",
					FlagType:        challengeentity.FlagTypeStatic,
					InstanceSharing: challengeentity.InstanceSharingPerUser,
				}, nil
			},
			updateWithHintsFn: func(context.Context, *challengeentity.Challenge, []*challengeentity.ChallengeHint, bool) error {
				updated = true
				return nil
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{
			findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*challengeentity.ChallengeTopology, error) {
				return nil, challengeports.ErrChallengeTopologyNotFound
			},
		},
		zap.NewNop(),
	)

	err := service.UpdateChallenge(context.Background(), 9, UpdateChallengeInput{InstanceSharing: string(challengeentity.InstanceSharingShared)})
	if err != nil {
		t.Fatalf("expected missing topology sentinel to be tolerated, got %v", err)
	}
	if !updated {
		t.Fatal("expected challenge update to proceed when topology is missing")
	}
}

func TestChallengePublishCheckServiceRequestPublishCheckTreatsMissingActiveJobSentinelAsNoActiveJob(t *testing.T) {
	t.Parallel()

	repo := &challengeCommandContextRepoStub{
		findByIDWithContextFn: func(context.Context, int64) (*challengeentity.Challenge, error) {
			return &challengeentity.Challenge{ID: 9, Status: challengeentity.ChallengeStatusDraft}, nil
		},
		findActivePublishCheckJobByIDFn: func(context.Context, int64) (*challengeentity.ChallengePublishCheckJob, error) {
			return nil, challengeports.ErrChallengePublishCheckJobNotFound
		},
		createPublishCheckJobFn: func(_ context.Context, job *challengeentity.ChallengePublishCheckJob) error {
			job.ID = 101
			job.CreatedAt = time.Now()
			job.UpdatedAt = job.CreatedAt
			return nil
		},
	}
	service := NewChallengePublishCheckService(
		repo,
		repo,
		nil,
		nil,
		SelfCheckConfig{},
		nil,
		zap.NewNop(),
	)

	resp, err := service.RequestPublishCheck(context.Background(), 1001, 9)
	if err != nil {
		t.Fatalf("RequestPublishCheck() error = %v", err)
	}
	if resp == nil || resp.ID != 101 {
		t.Fatalf("unexpected publish check resp: %+v", resp)
	}
}

func TestChallengePublishCheckServiceGetLatestPublishCheckTreatsMissingJobSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	repo := &challengeCommandContextRepoStub{
		findByIDWithContextFn: func(context.Context, int64) (*challengeentity.Challenge, error) {
			return &challengeentity.Challenge{ID: 9, UpdatedAt: time.Now()}, nil
		},
		findLatestPublishCheckJobByIDFn: func(context.Context, int64) (*challengeentity.ChallengePublishCheckJob, error) {
			return nil, challengeports.ErrChallengePublishCheckJobNotFound
		},
	}
	service := NewChallengePublishCheckService(
		repo,
		repo,
		nil,
		nil,
		SelfCheckConfig{},
		nil,
		zap.NewNop(),
	)

	_, err := service.GetLatestPublishCheck(context.Background(), 9)
	if err == nil || err.Error() != apperror.ErrNotFound.Error() {
		t.Fatalf("expected publish check not found error, got %v", err)
	}
}

func TestChallengeSelfCheckServiceSelfCheckChallengeTreatsModuleChallengeNotFoundAsErrChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeSelfCheckService(
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*challengeentity.Challenge, error) {
				return nil, challengeports.ErrChallengeCommandChallengeNotFound
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{},
		nil,
		SelfCheckConfig{},
		zap.NewNop(),
	)

	_, err := service.SelfCheckChallenge(context.Background(), 9)
	if err == nil || err.Error() != challengecontracts.ErrChallengeNotFound.Error() {
		t.Fatalf("expected challenge not found error, got %v", err)
	}
}
