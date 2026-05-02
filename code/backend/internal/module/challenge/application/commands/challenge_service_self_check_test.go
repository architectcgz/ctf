package commands

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
	flagcrypto "ctf-platform/pkg/crypto"
)

type fakeChallengeRuntimeProbe struct {
	createContainerCalled bool
	createTopologyCalled  bool
	cleanupCalled         bool

	containerResultAccessURL string
	containerResultDetails   model.InstanceRuntimeDetails
	containerResultErr       error

	topologyResult *challengeports.RuntimeTopologyCreateResult
	topologyErr    error

	cleanupErr error
}

func (f *fakeChallengeRuntimeProbe) CreateTopology(_ context.Context, _ *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	f.createTopologyCalled = true
	if f.topologyErr != nil {
		return nil, f.topologyErr
	}
	return f.topologyResult, nil
}

func (f *fakeChallengeRuntimeProbe) CreateContainer(_ context.Context, _ string, _ map[string]string) (string, model.InstanceRuntimeDetails, error) {
	f.createContainerCalled = true
	if f.containerResultErr != nil {
		return "", model.InstanceRuntimeDetails{}, f.containerResultErr
	}
	return f.containerResultAccessURL, f.containerResultDetails, nil
}

func (f *fakeChallengeRuntimeProbe) CleanupRuntimeDetails(_ context.Context, _ model.InstanceRuntimeDetails) error {
	f.cleanupCalled = true
	return f.cleanupErr
}

func TestChallengeSelfCheckSkipsRuntimeWhenPrecheckFails(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:      "no-image",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		FlagType:   model.FlagTypeStatic,
		FlagSalt:   salt,
		FlagHash:   flagcrypto.HashStaticFlag("flag{test}", salt),
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	probe := &fakeChallengeRuntimeProbe{}
	service := NewChallengeService(nil, repo, imageRepo, repo, repo, probe, SelfCheckConfig{}, zap.NewNop())

	resp, err := service.SelfCheckChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("SelfCheckChallenge() error = %v", err)
	}
	if resp.Precheck.Passed {
		t.Fatalf("expected precheck failed, got %+v", resp.Precheck)
	}
	if resp.Runtime.Passed {
		t.Fatalf("expected runtime failed when precheck fails, got %+v", resp.Runtime)
	}
	if probe.createContainerCalled || probe.createTopologyCalled {
		t.Fatalf("runtime should not start when precheck fails")
	}
}

func TestChallengeSelfCheckAttachmentOnlyChallengeSkipsRuntimeStartup(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:         "attachment-only",
		Category:      model.DimensionWeb,
		Difficulty:    model.ChallengeDifficultyEasy,
		Points:        100,
		AttachmentURL: "/api/v1/challenges/attachments/imports/web-source-audit-double-wrap-01/source.html",
		FlagType:      model.FlagTypeStatic,
		FlagSalt:      salt,
		FlagHash:      flagcrypto.HashStaticFlag("flag{test}", salt),
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	probe := &fakeChallengeRuntimeProbe{}
	service := NewChallengeService(nil, repo, imageRepo, repo, repo, probe, SelfCheckConfig{}, zap.NewNop())

	resp, err := service.SelfCheckChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("SelfCheckChallenge() error = %v", err)
	}
	if !resp.Precheck.Passed {
		t.Fatalf("expected attachment-only challenge precheck passed, got %+v", resp.Precheck)
	}
	if !resp.Runtime.Passed {
		t.Fatalf("expected attachment-only challenge runtime passed, got %+v", resp.Runtime)
	}
	if probe.createContainerCalled || probe.createTopologyCalled {
		t.Fatalf("attachment-only challenge should skip runtime startup")
	}
	if len(resp.Runtime.Steps) == 0 || resp.Runtime.Steps[0].Message == "" {
		t.Fatalf("expected runtime skip message, got %+v", resp.Runtime.Steps)
	}
}

func TestChallengeSelfCheckSingleContainerSuccess(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	image := &model.Image{
		Name:   "ctf/web-demo",
		Tag:    "latest",
		Status: model.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:      "single-container",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    image.ID,
		FlagType:   model.FlagTypeStatic,
		FlagSalt:   salt,
		FlagHash:   flagcrypto.HashStaticFlag("flag{ok}", salt),
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	probe := &fakeChallengeRuntimeProbe{
		containerResultAccessURL: "http://127.0.0.1:30001",
		containerResultDetails: model.InstanceRuntimeDetails{
			Containers: []model.InstanceRuntimeContainer{{ContainerID: "ctr-1"}},
			Networks:   []model.InstanceRuntimeNetwork{{NetworkID: "net-1"}},
		},
	}
	service := NewChallengeService(nil, repo, imageRepo, repo, repo, probe, SelfCheckConfig{}, zap.NewNop())

	resp, err := service.SelfCheckChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("SelfCheckChallenge() error = %v", err)
	}
	if !resp.Precheck.Passed {
		t.Fatalf("expected precheck passed, got %+v", resp.Precheck)
	}
	if !resp.Runtime.Passed {
		t.Fatalf("expected runtime passed, got %+v", resp.Runtime)
	}
	if !probe.createContainerCalled {
		t.Fatalf("expected single-container startup called")
	}
	if !probe.cleanupCalled {
		t.Fatalf("expected runtime cleanup called")
	}
}

func TestChallengeSelfCheckRuntimeStartupFailure(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	image := &model.Image{
		Name:   "ctf/web-broken",
		Tag:    "latest",
		Status: model.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate salt: %v", err)
	}
	challenge := &model.Challenge{
		Title:      "runtime-fail",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    image.ID,
		FlagType:   model.FlagTypeStatic,
		FlagSalt:   salt,
		FlagHash:   flagcrypto.HashStaticFlag("flag{ok}", salt),
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	probe := &fakeChallengeRuntimeProbe{
		containerResultErr: errors.New("docker start failed"),
	}
	service := NewChallengeService(nil, repo, imageRepo, repo, repo, probe, SelfCheckConfig{}, zap.NewNop())

	resp, err := service.SelfCheckChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("SelfCheckChallenge() error = %v", err)
	}
	if !resp.Precheck.Passed {
		t.Fatalf("expected precheck passed, got %+v", resp.Precheck)
	}
	if resp.Runtime.Passed {
		t.Fatalf("expected runtime failed, got %+v", resp.Runtime)
	}
	if probe.cleanupCalled {
		t.Fatalf("cleanup should not be called when startup failed before creating runtime details")
	}
}

func TestChallengeSelfCheckFailsOnInvalidRegexFlag(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	challenge := &model.Challenge{
		Title:      "regex-invalid",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		FlagType:   model.FlagTypeRegex,
		FlagRegex:  "[",
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := NewChallengeService(nil, repo, imageRepo, repo, repo, &fakeChallengeRuntimeProbe{}, SelfCheckConfig{}, zap.NewNop())

	resp, err := service.SelfCheckChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("SelfCheckChallenge() error = %v", err)
	}
	if resp.Precheck.Passed {
		t.Fatalf("expected precheck failed for invalid regex, got %+v", resp.Precheck)
	}
	if len(resp.Precheck.Steps) == 0 || resp.Precheck.Steps[0].Passed {
		t.Fatalf("expected flag_config step failure, got %+v", resp.Precheck.Steps)
	}
}

func TestChallengeSelfCheckManualReviewSkipsFlagValidationFailure(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	image := &model.Image{
		Name:   "ctf/web-manual",
		Tag:    "latest",
		Status: model.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	challenge := &model.Challenge{
		Title:      "manual-review",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    image.ID,
		FlagType:   model.FlagTypeManualReview,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	probe := &fakeChallengeRuntimeProbe{
		containerResultAccessURL: "http://127.0.0.1:30002",
		containerResultDetails: model.InstanceRuntimeDetails{
			Containers: []model.InstanceRuntimeContainer{{ContainerID: "ctr-manual"}},
			Networks:   []model.InstanceRuntimeNetwork{{NetworkID: "net-manual"}},
		},
	}
	service := NewChallengeService(nil, repo, imageRepo, repo, repo, probe, SelfCheckConfig{}, zap.NewNop())

	resp, err := service.SelfCheckChallenge(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("SelfCheckChallenge() error = %v", err)
	}
	if !resp.Precheck.Passed {
		t.Fatalf("expected manual review challenge precheck passed, got %+v", resp.Precheck)
	}
	if !resp.Runtime.Passed {
		t.Fatalf("expected manual review challenge runtime passed, got %+v", resp.Runtime)
	}
}
