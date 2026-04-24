package queries

import (
	"context"
	"encoding/json"
	"testing"

	"ctf-platform/internal/model"
)

type stubChallengeTopologyRepository struct {
	findByIDWithContextFn                         func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDCtxFn       func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
	upsertChallengeTopologyWithContextFn          func(ctx context.Context, topology *model.ChallengeTopology) error
	deleteChallengeTopologyByChallengeIDWithCtxFn func(ctx context.Context, challengeID int64) error
}

func (s *stubChallengeTopologyRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengeTopologyRepository) FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDCtxFn != nil {
		return s.findChallengeTopologyByChallengeIDCtxFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *stubChallengeTopologyRepository) UpsertChallengeTopologyWithContext(ctx context.Context, topology *model.ChallengeTopology) error {
	if s.upsertChallengeTopologyWithContextFn != nil {
		return s.upsertChallengeTopologyWithContextFn(ctx, topology)
	}
	return nil
}

func (s *stubChallengeTopologyRepository) DeleteChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) error {
	if s.deleteChallengeTopologyByChallengeIDWithCtxFn != nil {
		return s.deleteChallengeTopologyByChallengeIDWithCtxFn(ctx, challengeID)
	}
	return nil
}

type stubEnvironmentTemplateRepository struct {
	createWithContextFn     func(ctx context.Context, template *model.EnvironmentTemplate) error
	updateWithContextFn     func(ctx context.Context, template *model.EnvironmentTemplate) error
	deleteWithContextFn     func(ctx context.Context, id int64) error
	findByIDWithCtxFn       func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	listWithCtxFn           func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	incrementUsageWithCtxFn func(ctx context.Context, id int64) error
}

func (s *stubEnvironmentTemplateRepository) CreateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, template)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) UpdateWithContext(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, template)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return nil
}

func (s *stubEnvironmentTemplateRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
	if s.findByIDWithCtxFn != nil {
		return s.findByIDWithCtxFn(ctx, id)
	}
	return nil, nil
}

func (s *stubEnvironmentTemplateRepository) ListWithContext(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
	if s.listWithCtxFn != nil {
		return s.listWithCtxFn(ctx, keyword)
	}
	return nil, nil
}

func (s *stubEnvironmentTemplateRepository) IncrementUsageWithContext(ctx context.Context, id int64) error {
	if s.incrementUsageWithCtxFn != nil {
		return s.incrementUsageWithCtxFn(ctx, id)
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
