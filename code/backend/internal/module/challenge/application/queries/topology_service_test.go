package queries

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
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
	createFn         func(ctx context.Context, template *model.EnvironmentTemplate) error
	updateFn         func(ctx context.Context, template *model.EnvironmentTemplate) error
	deleteFn         func(ctx context.Context, id int64) error
	findByIDFn       func(ctx context.Context, id int64) (*model.EnvironmentTemplate, error)
	listFn           func(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error)
	incrementUsageFn func(ctx context.Context, id int64) error
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

type stubChallengePackageRevisionRepository struct {
	createFn     func(context.Context, *model.ChallengePackageRevision) error
	findByIDFn   func(context.Context, int64) (*model.ChallengePackageRevision, error)
	findLatestFn func(context.Context, int64) (*model.ChallengePackageRevision, error)
	listFn       func(context.Context, int64) ([]*model.ChallengePackageRevision, error)
}

func (s *stubChallengePackageRevisionRepository) CreateChallengePackageRevision(ctx context.Context, revision *model.ChallengePackageRevision) error {
	if s.createFn != nil {
		return s.createFn(ctx, revision)
	}
	return nil
}

func (s *stubChallengePackageRevisionRepository) FindChallengePackageRevisionByID(ctx context.Context, id int64) (*model.ChallengePackageRevision, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengePackageRevisionRepository) FindLatestChallengePackageRevisionByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePackageRevision, error) {
	if s.findLatestFn != nil {
		return s.findLatestFn(ctx, challengeID)
	}
	return nil, nil
}

func (s *stubChallengePackageRevisionRepository) ListChallengePackageRevisionsByChallengeID(ctx context.Context, challengeID int64) ([]*model.ChallengePackageRevision, error) {
	if s.listFn != nil {
		return s.listFn(ctx, challengeID)
	}
	return nil, nil
}

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

func TestTopologyServiceGetChallengeTopologyTreatsChallengeNotFoundAsChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewTopologyService(&stubChallengeTopologyRepository{
		findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, challengeports.ErrChallengeTopologyChallengeNotFound
		},
	}, nil)

	_, err := service.GetChallengeTopology(context.Background(), 404)
	if err == nil {
		t.Fatal("expected challenge not found")
	}
	if err.Error() != errcode.ErrChallengeNotFound.Error() {
		t.Fatalf("expected errcode.ErrChallengeNotFound, got %v", err)
	}
}

func TestTopologyServiceGetChallengeTopologyTreatsTopologyNotFoundAsNotFound(t *testing.T) {
	t.Parallel()

	service := NewTopologyService(&stubChallengeTopologyRepository{
		findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
			return &model.Challenge{ID: 9}, nil
		},
		findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*model.ChallengeTopology, error) {
			return nil, challengeports.ErrChallengeTopologyNotFound
		},
	}, nil)

	_, err := service.GetChallengeTopology(context.Background(), 9)
	if err == nil {
		t.Fatal("expected topology not found")
	}
	if err.Error() != errcode.ErrNotFound.Error() {
		t.Fatalf("expected errcode.ErrNotFound, got %v", err)
	}
}

func TestTopologyServiceGetChallengeTopologyIgnoresMissingPackageRevision(t *testing.T) {
	t.Parallel()

	spec, err := json.Marshal(model.TopologySpec{
		Nodes: []model.TopologyNode{{Key: "web", Name: "Web"}},
	})
	if err != nil {
		t.Fatalf("marshal topology spec: %v", err)
	}
	revisionID := int64(33)
	service := NewTopologyService(&stubChallengeTopologyRepository{
		findByIDWithContextFn: func(context.Context, int64) (*model.Challenge, error) {
			return &model.Challenge{ID: 9}, nil
		},
		findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*model.ChallengeTopology, error) {
			return &model.ChallengeTopology{
				ChallengeID:       9,
				EntryNodeKey:      "web",
				Spec:              string(spec),
				PackageRevisionID: &revisionID,
			}, nil
		},
	}, nil, &stubChallengePackageRevisionRepository{
		listFn: func(context.Context, int64) ([]*model.ChallengePackageRevision, error) {
			return []*model.ChallengePackageRevision{{ID: revisionID, ChallengeID: 9, RevisionNo: 1}}, nil
		},
		findByIDFn: func(context.Context, int64) (*model.ChallengePackageRevision, error) {
			return nil, challengeports.ErrChallengeTopologyPackageRevisionNotFound
		},
	})

	resp, err := service.GetChallengeTopology(context.Background(), 9)
	if err != nil {
		t.Fatalf("GetChallengeTopology() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected topology response")
	}
	if len(resp.PackageRevisions) != 1 {
		t.Fatalf("expected package revisions, got %+v", resp.PackageRevisions)
	}
	if resp.PackageFiles != nil {
		t.Fatalf("expected package files to be omitted, got %+v", resp.PackageFiles)
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

func TestTopologyServiceGetTemplateTreatsTemplateNotFoundAsNotFound(t *testing.T) {
	t.Parallel()

	service := NewTopologyService(nil, &stubEnvironmentTemplateRepository{
		findByIDFn: func(context.Context, int64) (*model.EnvironmentTemplate, error) {
			return nil, challengeports.ErrChallengeTopologyTemplateNotFound
		},
	})

	_, err := service.GetTemplate(context.Background(), 404)
	if err == nil {
		t.Fatal("expected template not found")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrNotFound.Code {
		t.Fatalf("expected errcode.ErrNotFound, got %v", err)
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
