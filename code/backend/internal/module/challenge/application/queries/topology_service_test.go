package queries

import (
	"context"
	"encoding/json"
	"testing"

	"ctf-platform/internal/model"
)

type stubChallengeTopologyRepository struct {
	findByIDWithContextFn                  func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn   func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	upsertChallengeTopologyFn              func(ctx context.Context, topology *model.ChallengeTopology) error
	deleteChallengeTopologyByChallengeIDFn func(ctx context.Context, challengeID int64) error
}

func (s *stubChallengeTopologyRepository) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengeTopologyRepository) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *stubChallengeTopologyRepository) UpsertChallengeTopology(ctx context.Context, topology *model.ChallengeTopology) error {
	if s.upsertChallengeTopologyFn != nil {
		return s.upsertChallengeTopologyFn(ctx, topology)
	}
	return nil
}

func (s *stubChallengeTopologyRepository) DeleteChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) error {
	if s.deleteChallengeTopologyByChallengeIDFn != nil {
		return s.deleteChallengeTopologyByChallengeIDFn(ctx, challengeID)
	}
	return nil
}

type stubEnvironmentTemplateRepository struct {
	createFn            func(ctx context.Context, template *model.EnvironmentTemplate) error
	updateFn            func(ctx context.Context, template *model.EnvironmentTemplate) error
	deleteFn            func(ctx context.Context, id int64) error
	findByIDFn          func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	listFn              func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	incrementUsageFn    func(ctx context.Context, id int64) error
}

func (s *stubEnvironmentTemplateRepository) Create(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.createFn != nil {
		return s.createFn(ctx, template)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) Update(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, template)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) Delete(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) FindByID(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *stubEnvironmentTemplateRepository) List(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
	if s.listFn != nil {
		return s.listFn(ctx, keyword)
	}
	return nil, nil
}

func (s *stubEnvironmentTemplateRepository) IncrementUsage(ctx context.Context, id int64) error {
	if s.incrementUsageFn != nil {
		return s.incrementUsageFn(ctx, id)
	}
	return nil
}

type challengeTopologyContextKey string

func TestTopologyServiceGetChallengeTopologyWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeTopologyContextKey("topology")
	expectedCtxValue := "ctx-topology"
	findChallengeCalled := false
	findTopologyCalled := false
	spec, err := json.Marshal(model.TopologySpec{
		Nodes: []model.TopologyNode{{Key: "web", Name: "Web", ServicePort: 8080}},
	})
	if err != nil {
		t.Fatalf("marshal topology spec: %v", err)
	}
	repo := &stubChallengeTopologyRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
			findChallengeCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Challenge{ID: id}, nil
		},
		findChallengeTopologyByChallengeIDFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
			findTopologyCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengeTopology{ChallengeID: challengeID, EntryNodeKey: "web", Spec: string(spec)}, nil
		},
	}
	service := NewTopologyService(repo, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetChallengeTopology(ctx, 11)
	if err != nil {
		t.Fatalf("GetChallengeTopology() error = %v", err)
	}
	if !findChallengeCalled || !findTopologyCalled {
		t.Fatalf("expected repository calls, got challenge=%v topology=%v", findChallengeCalled, findTopologyCalled)
	}
	if resp == nil || resp.EntryNodeKey != "web" || len(resp.Nodes) != 1 || resp.Nodes[0].Key != "web" {
		t.Fatalf("unexpected topology resp: %+v", resp)
	}
}

func TestTopologyServiceGetTemplateWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeTopologyContextKey("template")
	expectedCtxValue := "ctx-template"
	findTemplateCalled := false
	templateRepo := &stubEnvironmentTemplateRepository{
		findByIDFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
			findTemplateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.EnvironmentTemplate{ID: id, Name: "Base Web", EntryNodeKey: "web"}, nil
		},
	}
	service := NewTopologyService(nil, templateRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetTemplate(ctx, 21)
	if err != nil {
		t.Fatalf("GetTemplate() error = %v", err)
	}
	if !findTemplateCalled {
		t.Fatal("expected template repository find to be called")
	}
	if resp == nil || resp.ID != 21 || resp.Name != "Base Web" || resp.EntryNodeKey != "web" {
		t.Fatalf("unexpected template resp: %+v", resp)
	}
}

func TestTopologyServiceListTemplatesWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeTopologyContextKey("template-list")
	expectedCtxValue := "ctx-template-list"
	listCalled := false
	templateRepo := &stubEnvironmentTemplateRepository{
		listFn: func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list-templates ctx value %v, got %v", expectedCtxValue, got)
			}
			if keyword != "web" {
				t.Fatalf("unexpected keyword: %q", keyword)
			}
			return []*model.EnvironmentTemplate{{ID: 1, Name: "Base Web", EntryNodeKey: "web"}}, nil
		},
	}
	service := NewTopologyService(nil, templateRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListTemplates(ctx, " web ")
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected template repository list to be called")
	}
	if len(resp) != 1 || resp[0].ID != 1 || resp[0].Name != "Base Web" {
		t.Fatalf("unexpected template list resp: %+v", resp)
	}
}
