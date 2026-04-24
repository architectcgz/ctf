package commands

import (
	"context"
	"encoding/json"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type topologyCommandRepoStub struct {
	findByIDWithContextFn                  func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn   func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	upsertChallengeTopologyFn              func(ctx context.Context, topology *model.ChallengeTopology) error
	deleteChallengeTopologyByChallengeIDFn func(ctx context.Context, challengeID int64) error
}

func (s *topologyCommandRepoStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *topologyCommandRepoStub) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *topologyCommandRepoStub) UpsertChallengeTopology(ctx context.Context, topology *model.ChallengeTopology) error {
	if s.upsertChallengeTopologyFn != nil {
		return s.upsertChallengeTopologyFn(ctx, topology)
	}
	return nil
}

func (s *topologyCommandRepoStub) DeleteChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) error {
	if s.deleteChallengeTopologyByChallengeIDFn != nil {
		return s.deleteChallengeTopologyByChallengeIDFn(ctx, challengeID)
	}
	return nil
}

type topologyTemplateRepoStub struct {
	createFn            func(ctx context.Context, template *model.EnvironmentTemplate) error
	updateFn            func(ctx context.Context, template *model.EnvironmentTemplate) error
	deleteWithContextFn func(ctx context.Context, id int64) error
	findByIDFn          func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	listFn              func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	incrementUsageFn    func(ctx context.Context, id int64) error
}

func (s *topologyTemplateRepoStub) Create(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.createFn != nil {
		return s.createFn(ctx, template)
	}
	return nil
}

func (s *topologyTemplateRepoStub) Update(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, template)
	}
	return nil
}

func (s *topologyTemplateRepoStub) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return nil
}

func (s *topologyTemplateRepoStub) FindByID(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *topologyTemplateRepoStub) List(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
	if s.listFn != nil {
		return s.listFn(ctx, keyword)
	}
	return nil, nil
}

func (s *topologyTemplateRepoStub) IncrementUsage(ctx context.Context, id int64) error {
	if s.incrementUsageFn != nil {
		return s.incrementUsageFn(ctx, id)
	}
	return nil
}

type topologyImageRepoStub struct {
	createFn            func(ctx context.Context, image *model.Image) error
	findByIDFn          func(ctx context.Context, id int64) (*model.Image, error)
	findByNameTagFn     func(ctx context.Context, name, tag string) (*model.Image, error)
	listFn              func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	updateFn            func(ctx context.Context, image *model.Image) error
	deleteWithContextFn func(ctx context.Context, id int64) error
}

func (s *topologyImageRepoStub) Create(ctx context.Context, image *model.Image) error {
	if s.createFn != nil {
		return s.createFn(ctx, image)
	}
	return nil
}

func (s *topologyImageRepoStub) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *topologyImageRepoStub) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	if s.findByNameTagFn != nil {
		return s.findByNameTagFn(ctx, name, tag)
	}
	return nil, nil
}

func (s *topologyImageRepoStub) List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, name, status, offset, limit)
	}
	return nil, 0, nil
}

func (s *topologyImageRepoStub) Update(ctx context.Context, image *model.Image) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, image)
	}
	return nil
}

func (s *topologyImageRepoStub) Delete(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return nil
}

type topologyCommandContextKey string

func TestTopologyServiceSaveChallengeTopologyPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("save-topology")
	expectedCtxValue := "ctx-save-topology"
	templateID := int64(5)
	spec, err := json.Marshal(model.TopologySpec{Nodes: []model.TopologyNode{{Key: "web", Name: "Web"}}})
	if err != nil {
		t.Fatalf("marshal topology spec: %v", err)
	}

	findChallengeCalled := false
	findTemplateCalled := false
	upsertCalled := false
	incrementCalled := false
	findSavedCalled := false

	repo := &topologyCommandRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		upsertChallengeTopologyFn: func(ctx context.Context, topology *model.ChallengeTopology) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			if topology.TemplateID == nil || *topology.TemplateID != templateID {
				t.Fatalf("unexpected topology payload: %+v", topology)
			}
			return nil
		},
		findChallengeTopologyByChallengeIDFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
			findSavedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-saved-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengeTopology{ChallengeID: challengeID, TemplateID: &templateID, EntryNodeKey: "web", Spec: string(spec)}, nil
		},
	}
	templateRepo := &topologyTemplateRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
			findTemplateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.EnvironmentTemplate{ID: id, EntryNodeKey: "web", Spec: string(spec)}, nil
		},
		incrementUsageFn: func(ctx context.Context, id int64) error {
			incrementCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected increment-usage ctx value %v, got %v", expectedCtxValue, got)
			}
			if id != templateID {
				t.Fatalf("unexpected template id: %d", id)
			}
			return nil
		},
	}
	service := NewTopologyService(repo, templateRepo, &topologyImageRepoStub{})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.SaveChallengeTopology(ctx, 9, &dto.SaveChallengeTopologyReq{TemplateID: &templateID})
	if err != nil {
		t.Fatalf("SaveChallengeTopology() error = %v", err)
	}
	if !findChallengeCalled || !findTemplateCalled || !upsertCalled || !incrementCalled || !findSavedCalled {
		t.Fatalf("expected repository calls, got challenge=%v template=%v upsert=%v increment=%v saved=%v", findChallengeCalled, findTemplateCalled, upsertCalled, incrementCalled, findSavedCalled)
	}
	if resp == nil || resp.TemplateID == nil || *resp.TemplateID != templateID || resp.EntryNodeKey != "web" {
		t.Fatalf("unexpected topology resp: %+v", resp)
	}
}

func TestTopologyServiceDeleteChallengeTopologyPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("delete-topology")
	expectedCtxValue := "ctx-delete-topology"
	findCalled := false
	deleteCalled := false
	repo := &topologyCommandRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		deleteChallengeTopologyByChallengeIDFn: func(ctx context.Context, challengeID int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewTopologyService(repo, &topologyTemplateRepoStub{}, &topologyImageRepoStub{})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteChallengeTopology(ctx, 12); err != nil {
		t.Fatalf("DeleteChallengeTopology() error = %v", err)
	}
	if !findCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v delete=%v", findCalled, deleteCalled)
	}
}

func TestTopologyServiceCreateTemplatePropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("create-template")
	expectedCtxValue := "ctx-create-template"
	findImageCalled := false
	createCalled := false
	imageRepo := &topologyImageRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findImageCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1"}, nil
		},
	}
	templateRepo := &topologyTemplateRepoStub{
		createFn: func(ctx context.Context, template *model.EnvironmentTemplate) error {
			createCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create-template ctx value %v, got %v", expectedCtxValue, got)
			}
			if template.Name != "Base Web" {
				t.Fatalf("unexpected template payload: %+v", template)
			}
			return nil
		},
	}
	service := NewTopologyService(&topologyCommandRepoStub{}, templateRepo, imageRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.CreateTemplate(ctx, &dto.UpsertEnvironmentTemplateReq{
		Name:         "Base Web",
		EntryNodeKey: "web",
		Nodes:        []dto.TopologyNodeReq{{Key: "web", Name: "Web", ImageID: 7, ServicePort: 8080}},
	})
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}
	if !findImageCalled || !createCalled {
		t.Fatalf("expected repository calls, got image=%v create=%v", findImageCalled, createCalled)
	}
	if resp == nil || resp.Name != "Base Web" || resp.EntryNodeKey != "web" {
		t.Fatalf("unexpected template resp: %+v", resp)
	}
}

func TestTopologyServiceUpdateTemplatePropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("update-template")
	expectedCtxValue := "ctx-update-template"
	findTemplateCalled := false
	findImageCalled := false
	updateCalled := false
	templateRepo := &topologyTemplateRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
			findTemplateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.EnvironmentTemplate{ID: id, Name: "Old", EntryNodeKey: "old"}, nil
		},
		updateFn: func(ctx context.Context, template *model.EnvironmentTemplate) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update-template ctx value %v, got %v", expectedCtxValue, got)
			}
			if template.Name != "New Web" || template.EntryNodeKey != "web" {
				t.Fatalf("unexpected template payload: %+v", template)
			}
			return nil
		},
	}
	imageRepo := &topologyImageRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findImageCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v2"}, nil
		},
	}
	service := NewTopologyService(&topologyCommandRepoStub{}, templateRepo, imageRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpdateTemplate(ctx, 8, &dto.UpsertEnvironmentTemplateReq{
		Name:         "New Web",
		EntryNodeKey: "web",
		Nodes:        []dto.TopologyNodeReq{{Key: "web", Name: "Web", ImageID: 7, ServicePort: 8080}},
	})
	if err != nil {
		t.Fatalf("UpdateTemplate() error = %v", err)
	}
	if !findTemplateCalled || !findImageCalled || !updateCalled {
		t.Fatalf("expected repository calls, got template=%v image=%v update=%v", findTemplateCalled, findImageCalled, updateCalled)
	}
	if resp == nil || resp.ID != 8 || resp.Name != "New Web" {
		t.Fatalf("unexpected template resp: %+v", resp)
	}
}

func TestTopologyServiceDeleteTemplatePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("delete-template")
	expectedCtxValue := "ctx-delete-template"
	findCalled := false
	deleteCalled := false
	templateRepo := &topologyTemplateRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.EnvironmentTemplate{ID: id, Name: "Base Web"}, nil
		},
		deleteWithContextFn: func(ctx context.Context, id int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewTopologyService(&topologyCommandRepoStub{}, templateRepo, &topologyImageRepoStub{})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteTemplate(ctx, 3); err != nil {
		t.Fatalf("DeleteTemplate() error = %v", err)
	}
	if !findCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v delete=%v", findCalled, deleteCalled)
	}
}
