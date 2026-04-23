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
