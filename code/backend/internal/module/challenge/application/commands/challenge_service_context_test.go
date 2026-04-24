package commands

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeCommandContextRepoStub struct {
	createWithHintsFn               func(challenge *model.Challenge, hints []*model.ChallengeHint) error
	createWithHintsWithContextFn    func(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error
	findByIDFn                      func(id int64) (*model.Challenge, error)
	findByIDWithContextFn           func(ctx context.Context, id int64) (*model.Challenge, error)
	updateFn                        func(challenge *model.Challenge) error
	updateWithContextFn             func(ctx context.Context, challenge *model.Challenge) error
	updateWithHintsFn               func(challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	updateWithHintsWithContextFn    func(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error
	deleteFn                        func(id int64) error
	deleteWithContextFn             func(ctx context.Context, id int64) error
	hasRunningInstancesFn           func(challengeID int64) (bool, error)
	hasRunningInstancesWithCtxFn    func(ctx context.Context, challengeID int64) (bool, error)
	createPublishCheckJobFn         func(ctx context.Context, job *model.ChallengePublishCheckJob) error
	findActivePublishCheckJobByIDFn func(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error)
	findLatestPublishCheckJobByIDFn func(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error)
	findPublishCheckJobByIDFn       func(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error)
	updatePublishCheckJobFn         func(ctx context.Context, job *model.ChallengePublishCheckJob) error
}

func (s *challengeCommandContextRepoStub) CreateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint) error {
	if s.createWithHintsFn != nil {
		return s.createWithHintsFn(challenge, hints)
	}
	return nil
}

func (s *challengeCommandContextRepoStub) CreateWithHintsWithContext(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error {
	if s.createWithHintsWithContextFn != nil {
		return s.createWithHintsWithContextFn(ctx, challenge, hints)
	}
	return s.CreateWithHints(challenge, hints)
}

func (s *challengeCommandContextRepoStub) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *challengeCommandContextRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *challengeCommandContextRepoStub) Update(challenge *model.Challenge) error {
	if s.updateFn != nil {
		return s.updateFn(challenge)
	}
	return nil
}

func (s *challengeCommandContextRepoStub) UpdateWithContext(ctx context.Context, challenge *model.Challenge) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, challenge)
	}
	return s.Update(challenge)
}

func (s *challengeCommandContextRepoStub) UpdateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error {
	if s.updateWithHintsFn != nil {
		return s.updateWithHintsFn(challenge, hints, replaceHints)
	}
	return nil
}

func (s *challengeCommandContextRepoStub) UpdateWithHintsWithContext(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error {
	if s.updateWithHintsWithContextFn != nil {
		return s.updateWithHintsWithContextFn(ctx, challenge, hints, replaceHints)
	}
	return s.UpdateWithHints(challenge, hints, replaceHints)
}

func (s *challengeCommandContextRepoStub) Delete(id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

func (s *challengeCommandContextRepoStub) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return s.Delete(id)
}

func (s *challengeCommandContextRepoStub) HasRunningInstances(challengeID int64) (bool, error) {
	if s.hasRunningInstancesFn != nil {
		return s.hasRunningInstancesFn(challengeID)
	}
	return false, nil
}

func (s *challengeCommandContextRepoStub) HasRunningInstancesWithContext(ctx context.Context, challengeID int64) (bool, error) {
	if s.hasRunningInstancesWithCtxFn != nil {
		return s.hasRunningInstancesWithCtxFn(ctx, challengeID)
	}
	return s.HasRunningInstances(challengeID)
}

func (s *challengeCommandContextRepoStub) CreatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error {
	if s.createPublishCheckJobFn != nil {
		return s.createPublishCheckJobFn(ctx, job)
	}
	return nil
}

func (s *challengeCommandContextRepoStub) FindPublishCheckJobByID(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error) {
	if s.findPublishCheckJobByIDFn != nil {
		return s.findPublishCheckJobByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *challengeCommandContextRepoStub) FindActivePublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
	if s.findActivePublishCheckJobByIDFn != nil {
		return s.findActivePublishCheckJobByIDFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *challengeCommandContextRepoStub) FindLatestPublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
	if s.findLatestPublishCheckJobByIDFn != nil {
		return s.findLatestPublishCheckJobByIDFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *challengeCommandContextRepoStub) ListPendingPublishCheckJobs(ctx context.Context, limit int) ([]*model.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (s *challengeCommandContextRepoStub) TryStartPublishCheckJob(ctx context.Context, id int64, startedAt time.Time) (bool, error) {
	return false, nil
}

func (s *challengeCommandContextRepoStub) UpdatePublishCheckJob(ctx context.Context, job *model.ChallengePublishCheckJob) error {
	if s.updatePublishCheckJobFn != nil {
		return s.updatePublishCheckJobFn(ctx, job)
	}
	return nil
}

type challengeCommandImageRepoStub struct {
	findByIDWithContextFn func(ctx context.Context, id int64) (*model.Image, error)
}

func (s *challengeCommandImageRepoStub) CreateWithContext(ctx context.Context, image *model.Image) error {
	return nil
}
func (s *challengeCommandImageRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}
func (s *challengeCommandImageRepoStub) FindByNameTagWithContext(ctx context.Context, name, tag string) (*model.Image, error) {
	return nil, nil
}
func (s *challengeCommandImageRepoStub) ListWithContext(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	return nil, 0, nil
}
func (s *challengeCommandImageRepoStub) UpdateWithContext(ctx context.Context, image *model.Image) error {
	return nil
}
func (s *challengeCommandImageRepoStub) DeleteWithContext(ctx context.Context, id int64) error {
	return nil
}

type challengeCommandTopologyRepoStub struct {
	findByIDFn                              func(id int64) (*model.Challenge, error)
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn    func(challengeID int64) (*model.ChallengeTopology, error)
	findChallengeTopologyByChallengeIDCtxFn func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

func (s *challengeCommandTopologyRepoStub) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}
func (s *challengeCommandTopologyRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}
func (s *challengeCommandTopologyRepoStub) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(challengeID)
	}
	return nil, nil
}
func (s *challengeCommandTopologyRepoStub) FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDCtxFn != nil {
		return s.findChallengeTopologyByChallengeIDCtxFn(ctx, challengeID)
	}
	return s.FindChallengeTopologyByChallengeID(challengeID)
}
func (s *challengeCommandTopologyRepoStub) UpsertChallengeTopology(topology *model.ChallengeTopology) error {
	return nil
}
func (s *challengeCommandTopologyRepoStub) UpsertChallengeTopologyWithContext(ctx context.Context, topology *model.ChallengeTopology) error {
	return nil
}
func (s *challengeCommandTopologyRepoStub) DeleteChallengeTopologyByChallengeID(challengeID int64) error {
	return nil
}
func (s *challengeCommandTopologyRepoStub) DeleteChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) error {
	return nil
}

type challengeCommandContextKey string

func TestChallengeServiceCreateChallengeWithContextPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("create")
	expectedCtxValue := "ctx-create"
	imageCalled := false
	createCalled := false

	repo := &challengeCommandContextRepoStub{
		createWithHintsWithContextFn: func(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint) error {
			createCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			if challenge.CreatedBy == nil || *challenge.CreatedBy != 1001 || challenge.ImageID != 7 {
				t.Fatalf("unexpected challenge payload: %+v", challenge)
			}
			return nil
		},
	}
	imageRepo := &challengeCommandImageRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			imageCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected image find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1"}, nil
		},
	}
	service := NewChallengeService(nil, repo, imageRepo, &challengeCommandTopologyRepoStub{}, nil, SelfCheckConfig{}, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.CreateChallengeWithContext(ctx, 1001, &dto.CreateChallengeReq{
		Title:       "Test Challenge",
		Description: "desc",
		Category:    "web",
		Difficulty:  "easy",
		Points:      100,
		ImageID:     7,
	})
	if err != nil {
		t.Fatalf("CreateChallengeWithContext() error = %v", err)
	}
	if !imageCalled || !createCalled {
		t.Fatalf("expected repository calls, got image=%v create=%v", imageCalled, createCalled)
	}
	if resp == nil || resp.ImageID != 7 {
		t.Fatalf("unexpected challenge resp: %+v", resp)
	}
}

func TestChallengeServiceUpdateChallengeWithContextPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("update")
	expectedCtxValue := "ctx-update"
	findCalled := false
	imageCalled := false
	topologyCalled := false
	updateCalled := false
	rawSpec, err := model.EncodeTopologySpec(model.TopologySpec{Nodes: []model.TopologyNode{{Key: "web", Name: "Web", ServicePort: 8080}}})
	if err != nil {
		t.Fatalf("encode topology spec: %v", err)
	}

	repo := &challengeCommandContextRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Title: "Old", Category: "misc", Difficulty: "easy", Points: 50, FlagType: model.FlagTypeStatic, InstanceSharing: model.InstanceSharingPerUser}, nil
		},
		updateWithHintsWithContextFn: func(ctx context.Context, challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			if challenge.ImageID != 7 || challenge.InstanceSharing != model.InstanceSharingShared {
				t.Fatalf("unexpected updated challenge payload: %+v", challenge)
			}
			return nil
		},
	}
	imageRepo := &challengeCommandImageRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			imageCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected image find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v2"}, nil
		},
	}
	topologyRepo := &challengeCommandTopologyRepoStub{
		findChallengeTopologyByChallengeIDCtxFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
			topologyCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected topology find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengeTopology{ChallengeID: challengeID, EntryNodeKey: "web", Spec: rawSpec}, nil
		},
	}
	service := NewChallengeService(nil, repo, imageRepo, topologyRepo, nil, SelfCheckConfig{}, zap.NewNop())

	imageID := int64(7)
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.UpdateChallengeWithContext(ctx, 9, &dto.UpdateChallengeReq{ImageID: &imageID, InstanceSharing: model.InstanceSharingShared}); err != nil {
		t.Fatalf("UpdateChallengeWithContext() error = %v", err)
	}
	if !findCalled || !imageCalled || !topologyCalled || !updateCalled {
		t.Fatalf("expected repository calls, got find=%v image=%v topology=%v update=%v", findCalled, imageCalled, topologyCalled, updateCalled)
	}
}

func TestChallengeServiceDeleteChallengeWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("delete")
	expectedCtxValue := "ctx-delete"
	findCalled := false
	hasRunningCalled := false
	deleteCalled := false

	repo := &challengeCommandContextRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Title: "Delete Me"}, nil
		},
		hasRunningInstancesWithCtxFn: func(ctx context.Context, challengeID int64) (bool, error) {
			hasRunningCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected has-running ctx value %v, got %v", expectedCtxValue, got)
			}
			return false, nil
		},
		deleteWithContextFn: func(ctx context.Context, id int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewChallengeService(nil, repo, &challengeCommandImageRepoStub{}, &challengeCommandTopologyRepoStub{}, nil, SelfCheckConfig{}, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteChallengeWithContext(ctx, 12); err != nil {
		t.Fatalf("DeleteChallengeWithContext() error = %v", err)
	}
	if !findCalled || !hasRunningCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v hasRunning=%v delete=%v", findCalled, hasRunningCalled, deleteCalled)
	}
}

type challengeCommandRuntimeProbeStub struct {
	createContainerFn func(ctx context.Context, imageName string, env map[string]string) (string, model.InstanceRuntimeDetails, error)
	createTopologyFn  func(ctx context.Context, req *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error)
	cleanupFn         func(ctx context.Context, details model.InstanceRuntimeDetails) error
}

func (s *challengeCommandRuntimeProbeStub) CreateTopology(ctx context.Context, req *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	if s.createTopologyFn != nil {
		return s.createTopologyFn(ctx, req)
	}
	return nil, nil
}

func (s *challengeCommandRuntimeProbeStub) CreateContainer(ctx context.Context, imageName string, env map[string]string) (string, model.InstanceRuntimeDetails, error) {
	if s.createContainerFn != nil {
		return s.createContainerFn(ctx, imageName, env)
	}
	return "", model.InstanceRuntimeDetails{}, nil
}

func (s *challengeCommandRuntimeProbeStub) CleanupRuntimeDetails(ctx context.Context, details model.InstanceRuntimeDetails) error {
	if s.cleanupFn != nil {
		return s.cleanupFn(ctx, details)
	}
	return nil
}

func TestChallengeServiceRequestPublishCheckPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("request-publish-check")
	expectedCtxValue := "ctx-request-publish-check"
	findCalled := false
	activeCalled := false
	createCalled := false

	repo := &challengeCommandContextRepoStub{
		findByIDFn: func(id int64) (*model.Challenge, error) {
			t.Fatal("unexpected legacy find-by-id without context")
			return nil, nil
		},
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Title: "Publish Me", Status: model.ChallengeStatusDraft}, nil
		},
		findActivePublishCheckJobByIDFn: func(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
			activeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find active job ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, gorm.ErrRecordNotFound
		},
		createPublishCheckJobFn: func(ctx context.Context, job *model.ChallengePublishCheckJob) error {
			createCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create job ctx value %v, got %v", expectedCtxValue, got)
			}
			job.ID = 101
			return nil
		},
	}
	service := NewChallengeService(nil, repo, &challengeCommandImageRepoStub{}, &challengeCommandTopologyRepoStub{}, nil, SelfCheckConfig{}, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.RequestPublishCheckWithContext(ctx, 1001, 9)
	if err != nil {
		t.Fatalf("RequestPublishCheckWithContext() error = %v", err)
	}
	if !findCalled || !activeCalled || !createCalled {
		t.Fatalf("expected repository calls, got find=%v active=%v create=%v", findCalled, activeCalled, createCalled)
	}
	if resp == nil || resp.ID != 101 || resp.Status != "queued" || !resp.Active {
		t.Fatalf("unexpected publish check resp: %+v", resp)
	}
}

func TestChallengeServiceGetLatestPublishCheckPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("latest-publish-check")
	expectedCtxValue := "ctx-latest-publish-check"
	findCalled := false
	latestCalled := false
	now := time.Now()

	repo := &challengeCommandContextRepoStub{
		findByIDFn: func(id int64) (*model.Challenge, error) {
			t.Fatal("unexpected legacy find-by-id without context")
			return nil, nil
		},
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Title: "Publish Me", UpdatedAt: now}, nil
		},
		findLatestPublishCheckJobByIDFn: func(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
			latestCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find latest job ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengePublishCheckJob{ID: 21, ChallengeID: challengeID, Status: model.ChallengePublishCheckStatusPassed, UpdatedAt: now}, nil
		},
	}
	service := NewChallengeService(nil, repo, &challengeCommandImageRepoStub{}, &challengeCommandTopologyRepoStub{}, nil, SelfCheckConfig{}, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetLatestPublishCheckWithContext(ctx, 9)
	if err != nil {
		t.Fatalf("GetLatestPublishCheckWithContext() error = %v", err)
	}
	if !findCalled || !latestCalled {
		t.Fatalf("expected repository calls, got find=%v latest=%v", findCalled, latestCalled)
	}
	if resp == nil || resp.ID != 21 || resp.Status != "succeeded" {
		t.Fatalf("unexpected latest publish check resp: %+v", resp)
	}
}

func TestChallengeServiceSelfCheckChallengePropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("self-check")
	expectedCtxValue := "ctx-self-check"
	findCalled := false
	imageCalled := false
	topologyCalled := false
	createCalled := false
	cleanupCalled := false

	repo := &challengeCommandContextRepoStub{
		findByIDFn: func(id int64) (*model.Challenge, error) {
			t.Fatal("unexpected legacy find-by-id without context")
			return nil, nil
		},
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Title: "Self Check", ImageID: 7, FlagType: model.FlagTypeStatic, FlagHash: "flag{ok}", FlagSalt: "salt"}, nil
		},
	}
	imageRepo := &challengeCommandImageRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			imageCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected image find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "latest", Status: model.ImageStatusAvailable}, nil
		},
	}
	topologyRepo := &challengeCommandTopologyRepoStub{
		findChallengeTopologyByChallengeIDFn: func(challengeID int64) (*model.ChallengeTopology, error) {
			t.Fatal("unexpected legacy topology find without context")
			return nil, nil
		},
		findChallengeTopologyByChallengeIDCtxFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
			topologyCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected topology find ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, gorm.ErrRecordNotFound
		},
	}
	probe := &challengeCommandRuntimeProbeStub{
		createContainerFn: func(ctx context.Context, imageName string, env map[string]string) (string, model.InstanceRuntimeDetails, error) {
			createCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected runtime create ctx value %v, got %v", expectedCtxValue, got)
			}
			return "http://127.0.0.1:30001", model.InstanceRuntimeDetails{Containers: []model.InstanceRuntimeContainer{{ContainerID: "ctr-1"}}, Networks: []model.InstanceRuntimeNetwork{{NetworkID: "net-1"}}}, nil
		},
		cleanupFn: func(ctx context.Context, details model.InstanceRuntimeDetails) error {
			cleanupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected runtime cleanup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewChallengeService(nil, repo, imageRepo, topologyRepo, probe, SelfCheckConfig{RuntimeCreateTimeout: time.Second}, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.SelfCheckChallengeWithContext(ctx, 9)
	if err != nil {
		t.Fatalf("SelfCheckChallengeWithContext() error = %v", err)
	}
	if !findCalled || !imageCalled || !topologyCalled || !createCalled || !cleanupCalled {
		t.Fatalf("expected calls, got find=%v image=%v topology=%v create=%v cleanup=%v", findCalled, imageCalled, topologyCalled, createCalled, cleanupCalled)
	}
	if resp == nil || !resp.Precheck.Passed || !resp.Runtime.Passed {
		t.Fatalf("unexpected self-check resp: %+v", resp)
	}
}

func TestChallengeServicePublishChallengeWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("publish")
	expectedCtxValue := "ctx-publish"
	findCalled := false
	updateCalled := false

	repo := &challengeCommandContextRepoStub{
		findByIDFn: func(id int64) (*model.Challenge, error) {
			t.Fatal("unexpected legacy find-by-id without context")
			return nil, nil
		},
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Title: "Publish Me", Status: model.ChallengeStatusDraft}, nil
		},
		updateWithContextFn: func(ctx context.Context, challenge *model.Challenge) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			if challenge.Status != model.ChallengeStatusPublished {
				t.Fatalf("unexpected challenge payload: %+v", challenge)
			}
			return nil
		},
	}
	service := NewChallengeService(nil, repo, &challengeCommandImageRepoStub{}, &challengeCommandTopologyRepoStub{}, nil, SelfCheckConfig{}, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.PublishChallengeWithContext(ctx, 15); err != nil {
		t.Fatalf("PublishChallengeWithContext() error = %v", err)
	}
	if !findCalled || !updateCalled {
		t.Fatalf("expected repository calls, got find=%v update=%v", findCalled, updateCalled)
	}
}

func TestChallengeServiceProcessPublishCheckJobPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := challengeCommandContextKey("process-publish-job")
	expectedCtxValue := "ctx-process-publish-job"
	loadJobCalled := false
	findChallengeCalled := false
	publishUpdateCalled := false
	updateJobCalled := 0

	repo := &challengeCommandContextRepoStub{
		findPublishCheckJobByIDFn: func(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error) {
			loadJobCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected load job ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengePublishCheckJob{ID: id, ChallengeID: 21, RequestedBy: 1001, Status: model.ChallengePublishCheckStatusRunning}, nil
		},
		findByIDFn: func(id int64) (*model.Challenge, error) {
			t.Fatal("unexpected legacy find-by-id without context")
			return nil, nil
		},
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id, Title: "Attachment Only", AttachmentURL: "/tmp/source.zip", Status: model.ChallengeStatusDraft, FlagType: model.FlagTypeStatic, FlagHash: "flag{ok}", FlagSalt: "salt"}, nil
		},
		updateWithContextFn: func(ctx context.Context, challenge *model.Challenge) error {
			publishUpdateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected publish update ctx value %v, got %v", expectedCtxValue, got)
			}
			if challenge.Status != model.ChallengeStatusPublished {
				t.Fatalf("unexpected published challenge payload: %+v", challenge)
			}
			return nil
		},
		updatePublishCheckJobFn: func(ctx context.Context, job *model.ChallengePublishCheckJob) error {
			updateJobCalled++
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update job ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	topologyRepo := &challengeCommandTopologyRepoStub{
		findChallengeTopologyByChallengeIDFn: func(challengeID int64) (*model.ChallengeTopology, error) {
			t.Fatal("unexpected legacy topology find without context")
			return nil, nil
		},
		findChallengeTopologyByChallengeIDCtxFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected topology find ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, gorm.ErrRecordNotFound
		},
	}
	service := NewChallengeService(nil, repo, &challengeCommandImageRepoStub{}, topologyRepo, &challengeCommandRuntimeProbeStub{}, SelfCheckConfig{}, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	service.processPublishCheckJob(ctx, 51)
	if !loadJobCalled || !findChallengeCalled || !publishUpdateCalled || updateJobCalled == 0 {
		t.Fatalf("expected process job calls, got load=%v find=%v publish=%v updateJob=%d", loadJobCalled, findChallengeCalled, publishUpdateCalled, updateJobCalled)
	}
}
