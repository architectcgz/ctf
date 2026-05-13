package commands

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

func TestChallengeServiceCreateChallengeTreatsModuleImageNotFoundAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		nil,
		&challengeCommandContextRepoStub{},
		&challengeCommandImageRepoStub{
			findByIDFn: func(context.Context, int64) (*model.Image, error) {
				return nil, challengeports.ErrChallengeImageNotFound
			},
		},
		&challengeCommandTopologyRepoStub{},
		nil,
		nil,
		SelfCheckConfig{},
		zap.NewNop(),
	)

	_, err := service.CreateChallenge(context.Background(), 1001, CreateChallengeInput{ImageID: 9})
	if err == nil || err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected image not found error, got %v", err)
	}
}

func TestChallengeServiceUpdateChallengeTreatsModuleChallengeNotFoundAsErrChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		nil,
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
				return nil, challengeports.ErrChallengeCommandChallengeNotFound
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{},
		nil,
		nil,
		SelfCheckConfig{},
		zap.NewNop(),
	)

	err := service.UpdateChallenge(context.Background(), 9, UpdateChallengeInput{Title: "updated"})
	if err == nil || err.Error() != errcode.ErrChallengeNotFound.Error() {
		t.Fatalf("expected challenge not found error, got %v", err)
	}
}

func TestChallengeServiceUpdateChallengeTreatsTopologySentinelAsMissingTopology(t *testing.T) {
	t.Parallel()

	updated := false
	service := NewChallengeService(
		nil,
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:              9,
					Title:           "shared",
					FlagType:        model.FlagTypeStatic,
					InstanceSharing: model.InstanceSharingPerUser,
				}, nil
			},
			updateWithHintsFn: func(context.Context, *model.Challenge, []*model.ChallengeHint, bool) error {
				updated = true
				return nil
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{
			findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*model.ChallengeTopology, error) {
				return nil, challengeports.ErrChallengeTopologyNotFound
			},
		},
		nil,
		nil,
		SelfCheckConfig{},
		zap.NewNop(),
	)

	err := service.UpdateChallenge(context.Background(), 9, UpdateChallengeInput{InstanceSharing: model.InstanceSharingShared})
	if err != nil {
		t.Fatalf("expected missing topology sentinel to be tolerated, got %v", err)
	}
	if !updated {
		t.Fatal("expected challenge update to proceed when topology is missing")
	}
}

func TestChallengeServiceRequestPublishCheckTreatsMissingActiveJobSentinelAsNoActiveJob(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		nil,
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
				return &model.Challenge{ID: 9, Status: model.ChallengeStatusDraft}, nil
			},
			findActivePublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
				return nil, challengeports.ErrChallengePublishCheckJobNotFound
			},
			createPublishCheckJobFn: func(_ context.Context, job *model.ChallengePublishCheckJob) error {
				job.ID = 101
				job.CreatedAt = time.Now()
				job.UpdatedAt = job.CreatedAt
				return nil
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{},
		nil,
		nil,
		SelfCheckConfig{},
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

func TestChallengeServiceGetLatestPublishCheckTreatsMissingJobSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		nil,
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
				return &model.Challenge{ID: 9, UpdatedAt: time.Now()}, nil
			},
			findLatestPublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
				return nil, challengeports.ErrChallengePublishCheckJobNotFound
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{},
		nil,
		nil,
		SelfCheckConfig{},
		zap.NewNop(),
	)

	_, err := service.GetLatestPublishCheck(context.Background(), 9)
	if err == nil || err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected publish check not found error, got %v", err)
	}
}

func TestChallengeServiceSelfCheckChallengeTreatsModuleChallengeNotFoundAsErrChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		nil,
		&challengeCommandContextRepoStub{
			findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
				return nil, challengeports.ErrChallengeCommandChallengeNotFound
			},
		},
		&challengeCommandImageRepoStub{},
		&challengeCommandTopologyRepoStub{},
		nil,
		nil,
		SelfCheckConfig{},
		zap.NewNop(),
	)

	_, err := service.SelfCheckChallenge(context.Background(), 9)
	if err == nil || err.Error() != errcode.ErrChallengeNotFound.Error() {
		t.Fatalf("expected challenge not found error, got %v", err)
	}
}
