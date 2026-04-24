package commands

import (
	"context"
	"encoding/json"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type topologyCommandRepoStub struct {
	findByIDFn                                    func(id int64) (*model.Challenge, error)
	findByIDWithContextFn                         func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn          func(challengeID int64) (*model.ChallengeTopology, error)
	findChallengeTopologyByChallengeIDWithCtxFn   func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	upsertChallengeTopologyFn                     func(topology *model.ChallengeTopology) error
	upsertChallengeTopologyWithContextFn          func(ctx context.Context, topology *model.ChallengeTopology) error
	deleteChallengeTopologyByChallengeIDFn        func(challengeID int64) error
	deleteChallengeTopologyByChallengeIDWithCtxFn func(ctx context.Context, challengeID int64) error
}

func (s *topologyCommandRepoStub) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *topologyCommandRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *topologyCommandRepoStub) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(challengeID)
	}
	return nil, nil
}

func (s *topologyCommandRepoStub) FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDWithCtxFn != nil {
		return s.findChallengeTopologyByChallengeIDWithCtxFn(ctx, challengeID)
	}
	return s.FindChallengeTopologyByChallengeID(challengeID)
}

func (s *topologyCommandRepoStub) UpsertChallengeTopology(topology *model.ChallengeTopology) error {
	if s.upsertChallengeTopologyFn != nil {
		return s.upsertChallengeTopologyFn(topology)
	}
	return nil
}

func (s *topologyCommandRepoStub) UpsertChallengeTopologyWithContext(ctx context.Context, topology *model.ChallengeTopology) error {
	if s.upsertChallengeTopologyWithContextFn != nil {
		return s.upsertChallengeTopologyWithContextFn(ctx, topology)
	}
	return s.UpsertChallengeTopology(topology)
}

func (s *topologyCommandRepoStub) DeleteChallengeTopologyByChallengeID(challengeID int64) error {
	if s.deleteChallengeTopologyByChallengeIDFn != nil {
		return s.deleteChallengeTopologyByChallengeIDFn(challengeID)
	}
	return nil
}

func (s *topologyCommandRepoStub) DeleteChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) error {
	if s.deleteChallengeTopologyByChallengeIDWithCtxFn != nil {
		return s.deleteChallengeTopologyByChallengeIDWithCtxFn(ctx, challengeID)
	}
	return s.DeleteChallengeTopologyByChallengeID(challengeID)
}

type topologyTemplateRepoStub struct {
	createWithContextFn     func(ctx context.Context, template *model.EnvironmentTemplate) error
	updateWithContextFn     func(ctx context.Context, template *model.EnvironmentTemplate) error
	deleteWithContextFn     func(ctx context.Context, id int64) error
	findByIDWithContextFn   func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	listWithContextFn       func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	incrementUsageWithCtxFn func(ctx context.Context, id int64) error
}

func (s *topologyTemplateRepoStub) CreateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, template)
	}
	return nil
}

func (s *topologyTemplateRepoStub) UpdateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, template)
	}
	return nil
}

func (s *topologyTemplateRepoStub) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return nil
}

func (s *topologyTemplateRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *topologyTemplateRepoStub) ListWithContext(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
	if s.listWithContextFn != nil {
		return s.listWithContextFn(ctx, keyword)
	}
	return nil, nil
}

func (s *topologyTemplateRepoStub) IncrementUsageWithContext(ctx context.Context, id int64) error {
	if s.incrementUsageWithCtxFn != nil {
		return s.incrementUsageWithCtxFn(ctx, id)
	}
	return nil
}

type topologyImageRepoStub struct {
	createFn                   func(image *model.Image) error
	createWithContextFn        func(ctx context.Context, image *model.Image) error
	findByIDFn                 func(id int64) (*model.Image, error)
	findByIDWithContextFn      func(ctx context.Context, id int64) (*model.Image, error)
	findByNameTagFn            func(name, tag string) (*model.Image, error)
	findByNameTagWithContextFn func(ctx context.Context, name, tag string) (*model.Image, error)
	listFn                     func(name, status string, offset, limit int) ([]*model.Image, int64, error)
	listWithContextFn          func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	updateFn                   func(image *model.Image) error
	updateWithContextFn        func(ctx context.Context, image *model.Image) error
	deleteFn                   func(id int64) error
	deleteWithContextFn        func(ctx context.Context, id int64) error
}

func (s *topologyImageRepoStub) Create(image *model.Image) error {
	if s.createFn != nil {
		return s.createFn(image)
	}
	return nil
}

func (s *topologyImageRepoStub) CreateWithContext(ctx context.Context, image *model.Image) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, image)
	}
	return s.Create(image)
}

func (s *topologyImageRepoStub) FindByID(id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *topologyImageRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *topologyImageRepoStub) FindByNameTag(name, tag string) (*model.Image, error) {
	if s.findByNameTagFn != nil {
		return s.findByNameTagFn(name, tag)
	}
	return nil, nil
}

func (s *topologyImageRepoStub) FindByNameTagWithContext(ctx context.Context, name, tag string) (*model.Image, error) {
	if s.findByNameTagWithContextFn != nil {
		return s.findByNameTagWithContextFn(ctx, name, tag)
	}
	return s.FindByNameTag(name, tag)
}

func (s *topologyImageRepoStub) List(name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listFn != nil {
		return s.listFn(name, status, offset, limit)
	}
	return nil, 0, nil
}

func (s *topologyImageRepoStub) ListWithContext(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listWithContextFn != nil {
		return s.listWithContextFn(ctx, name, status, offset, limit)
	}
	return s.List(name, status, offset, limit)
}

func (s *topologyImageRepoStub) Update(image *model.Image) error {
	if s.updateFn != nil {
		return s.updateFn(image)
	}
	return nil
}

func (s *topologyImageRepoStub) UpdateWithContext(ctx context.Context, image *model.Image) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, image)
	}
	return s.Update(image)
}

func (s *topologyImageRepoStub) Delete(id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

func (s *topologyImageRepoStub) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return s.Delete(id)
}

type topologyCommandContextKey string

func TestTopologyServiceSaveChallengeTopologyWithContextPropagatesContextToRepositories(t *testing.T) {
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
		upsertChallengeTopologyWithContextFn: func(ctx context.Context, topology *model.ChallengeTopology) error {
			upsertCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected upsert-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			if topology.TemplateID == nil || *topology.TemplateID != templateID {
				t.Fatalf("unexpected topology payload: %+v", topology)
			}
			return nil
		},
		findChallengeTopologyByChallengeIDWithCtxFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
			findSavedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-saved-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengeTopology{ChallengeID: challengeID, TemplateID: &templateID, EntryNodeKey: "web", Spec: string(spec)}, nil
		},
	}
	templateRepo := &topologyTemplateRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
			findTemplateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.EnvironmentTemplate{ID: id, EntryNodeKey: "web", Spec: string(spec)}, nil
		},
		incrementUsageWithCtxFn: func(ctx context.Context, id int64) error {
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
	resp, err := service.SaveChallengeTopologyWithContext(ctx, 9, &dto.SaveChallengeTopologyReq{TemplateID: &templateID})
	if err != nil {
		t.Fatalf("SaveChallengeTopologyWithContext() error = %v", err)
	}
	if !findChallengeCalled || !findTemplateCalled || !upsertCalled || !incrementCalled || !findSavedCalled {
		t.Fatalf("expected repository calls, got challenge=%v template=%v upsert=%v increment=%v saved=%v", findChallengeCalled, findTemplateCalled, upsertCalled, incrementCalled, findSavedCalled)
	}
	if resp == nil || resp.TemplateID == nil || *resp.TemplateID != templateID || resp.EntryNodeKey != "web" {
		t.Fatalf("unexpected topology resp: %+v", resp)
	}
}

func TestTopologyServiceDeleteChallengeTopologyWithContextPropagatesContextToRepository(t *testing.T) {
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
		deleteChallengeTopologyByChallengeIDWithCtxFn: func(ctx context.Context, challengeID int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewTopologyService(repo, &topologyTemplateRepoStub{}, &topologyImageRepoStub{})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteChallengeTopologyWithContext(ctx, 12); err != nil {
		t.Fatalf("DeleteChallengeTopologyWithContext() error = %v", err)
	}
	if !findCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v delete=%v", findCalled, deleteCalled)
	}
}

func TestTopologyServiceCreateTemplateWithContextPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("create-template")
	expectedCtxValue := "ctx-create-template"
	findImageCalled := false
	createCalled := false
	imageRepo := &topologyImageRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findImageCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1"}, nil
		},
	}
	templateRepo := &topologyTemplateRepoStub{
		createWithContextFn: func(ctx context.Context, template *model.EnvironmentTemplate) error {
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
	resp, err := service.CreateTemplateWithContext(ctx, &dto.UpsertEnvironmentTemplateReq{
		Name:         "Base Web",
		EntryNodeKey: "web",
		Nodes:        []dto.TopologyNodeReq{{Key: "web", Name: "Web", ImageID: 7, ServicePort: 8080}},
	})
	if err != nil {
		t.Fatalf("CreateTemplateWithContext() error = %v", err)
	}
	if !findImageCalled || !createCalled {
		t.Fatalf("expected repository calls, got image=%v create=%v", findImageCalled, createCalled)
	}
	if resp == nil || resp.Name != "Base Web" || resp.EntryNodeKey != "web" {
		t.Fatalf("unexpected template resp: %+v", resp)
	}
}

func TestTopologyServiceUpdateTemplateWithContextPropagatesContextToRepositories(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("update-template")
	expectedCtxValue := "ctx-update-template"
	findTemplateCalled := false
	findImageCalled := false
	updateCalled := false
	templateRepo := &topologyTemplateRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
			findTemplateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.EnvironmentTemplate{ID: id, Name: "Old", EntryNodeKey: "old"}, nil
		},
		updateWithContextFn: func(ctx context.Context, template *model.EnvironmentTemplate) error {
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
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findImageCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v2"}, nil
		},
	}
	service := NewTopologyService(&topologyCommandRepoStub{}, templateRepo, imageRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.UpdateTemplateWithContext(ctx, 8, &dto.UpsertEnvironmentTemplateReq{
		Name:         "New Web",
		EntryNodeKey: "web",
		Nodes:        []dto.TopologyNodeReq{{Key: "web", Name: "Web", ImageID: 7, ServicePort: 8080}},
	})
	if err != nil {
		t.Fatalf("UpdateTemplateWithContext() error = %v", err)
	}
	if !findTemplateCalled || !findImageCalled || !updateCalled {
		t.Fatalf("expected repository calls, got template=%v image=%v update=%v", findTemplateCalled, findImageCalled, updateCalled)
	}
	if resp == nil || resp.ID != 8 || resp.Name != "New Web" {
		t.Fatalf("unexpected template resp: %+v", resp)
	}
}

func TestTopologyServiceDeleteTemplateWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := topologyCommandContextKey("delete-template")
	expectedCtxValue := "ctx-delete-template"
	findCalled := false
	deleteCalled := false
	templateRepo := &topologyTemplateRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
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
	if err := service.DeleteTemplateWithContext(ctx, 3); err != nil {
		t.Fatalf("DeleteTemplateWithContext() error = %v", err)
	}
	if !findCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v delete=%v", findCalled, deleteCalled)
	}
}
