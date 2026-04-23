package queries

import (
	"context"
	"encoding/json"
	"testing"

	"ctf-platform/internal/model"
)

type stubChallengeTopologyRepository struct {
	findByIDFn                              func(id int64) (*model.Challenge, error)
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn    func(challengeID int64) (*model.ChallengeTopology, error)
	findChallengeTopologyByChallengeIDCtxFn func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	upsertChallengeTopologyFn               func(topology *model.ChallengeTopology) error
	deleteChallengeTopologyByChallengeIDFn  func(challengeID int64) error
}

func (s *stubChallengeTopologyRepository) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *stubChallengeTopologyRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *stubChallengeTopologyRepository) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(challengeID)
	}
	return nil, nil
}

func (s *stubChallengeTopologyRepository) FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDCtxFn != nil {
		return s.findChallengeTopologyByChallengeIDCtxFn(ctx, challengeID)
	}
	return s.FindChallengeTopologyByChallengeID(challengeID)
}

func (s *stubChallengeTopologyRepository) UpsertChallengeTopology(topology *model.ChallengeTopology) error {
	if s.upsertChallengeTopologyFn != nil {
		return s.upsertChallengeTopologyFn(topology)
	}
	return nil
}

func (s *stubChallengeTopologyRepository) DeleteChallengeTopologyByChallengeID(challengeID int64) error {
	if s.deleteChallengeTopologyByChallengeIDFn != nil {
		return s.deleteChallengeTopologyByChallengeIDFn(challengeID)
	}
	return nil
}

type stubEnvironmentTemplateRepository struct {
	createFn          func(template *model.EnvironmentTemplate) error
	updateFn          func(template *model.EnvironmentTemplate) error
	deleteFn          func(id int64) error
	findByIDFn        func(id int64) (*model.EnvironmentTemplate, error)
	findByIDWithCtxFn func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	listFn            func(keyword string) ([]*model.EnvironmentTemplate, error)
	listWithCtxFn     func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	incrementUsageFn  func(id int64) error
}

func (s *stubEnvironmentTemplateRepository) Create(template *model.EnvironmentTemplate) error {
	if s.createFn != nil {
		return s.createFn(template)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) Update(template *model.EnvironmentTemplate) error {
	if s.updateFn != nil {
		return s.updateFn(template)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) Delete(id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) FindByID(id int64) (*model.EnvironmentTemplate, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *stubEnvironmentTemplateRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
	if s.findByIDWithCtxFn != nil {
		return s.findByIDWithCtxFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *stubEnvironmentTemplateRepository) List(keyword string) ([]*model.EnvironmentTemplate, error) {
	if s.listFn != nil {
		return s.listFn(keyword)
	}
	return nil, nil
}

func (s *stubEnvironmentTemplateRepository) ListWithContext(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
	if s.listWithCtxFn != nil {
		return s.listWithCtxFn(ctx, keyword)
	}
	return s.List(keyword)
}

func (s *stubEnvironmentTemplateRepository) IncrementUsage(id int64) error {
	if s.incrementUsageFn != nil {
		return s.incrementUsageFn(id)
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
		findChallengeTopologyByChallengeIDCtxFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
			findTopologyCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-topology ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.ChallengeTopology{ChallengeID: challengeID, EntryNodeKey: "web", Spec: string(spec)}, nil
		},
	}
	service := NewTopologyService(repo, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetChallengeTopologyWithContext(ctx, 11)
	if err != nil {
		t.Fatalf("GetChallengeTopologyWithContext() error = %v", err)
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
		findByIDWithCtxFn: func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
			findTemplateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-template ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.EnvironmentTemplate{ID: id, Name: "Base Web", EntryNodeKey: "web"}, nil
		},
	}
	service := NewTopologyService(nil, templateRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetTemplateWithContext(ctx, 21)
	if err != nil {
		t.Fatalf("GetTemplateWithContext() error = %v", err)
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
		listWithCtxFn: func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
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
	resp, err := service.ListTemplatesWithContext(ctx, " web ")
	if err != nil {
		t.Fatalf("ListTemplatesWithContext() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected template repository list to be called")
	}
	if len(resp) != 1 || resp[0].ID != 1 || resp[0].Name != "Base Web" {
		t.Fatalf("unexpected template list resp: %+v", resp)
	}
}
