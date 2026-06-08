package commands

import (
	"context"
	"fmt"
	"testing"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
)

func TestLoadRuntimeSubjectWithScopePropagatesContextToChallengeContract(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("runtime-subject")
	expectedCtxValue := "ctx-runtime-subject"
	challengeLookupCalled := false
	topologyLookupCalled := false
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			challengeLookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengecontracts.PracticeRuntimeChallenge{ID: id, Status: challengecontracts.ChallengeStatusPublished}, nil
		},
		findChallengeTopologyByChallengeIDFn: func(ctx context.Context, challengeID int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
			topologyLookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected topology lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, nil
		},
	}
	service := wirePracticeScopeAdapters(NewService(
		nil,

		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil),

		nil, challengeRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	challenge, topology, err := service.loadRuntimeSubjectWithScope(ctx, practiceports.InstanceScope{}, 42)
	if err != nil {
		t.Fatalf("loadRuntimeSubjectWithScope() error = %v", err)
	}
	if challenge == nil || challenge.ID != 42 {
		t.Fatalf("expected challenge 42, got %+v", challenge)
	}
	if topology != nil {
		t.Fatalf("expected nil topology, got %+v", topology)
	}
	if !challengeLookupCalled {
		t.Fatal("expected challenge lookup to be called")
	}
	if !topologyLookupCalled {
		t.Fatal("expected topology lookup to be called")
	}
}

func TestBuildTopologyCreateRequestPropagatesContextToImageRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("topology-image")
	expectedCtxValue := "ctx-topology-image"
	lookups := make([]int64, 0, 2)
	service := &Service{
		imageRepo: &stubPracticeImageStore{
			findByIDFn: func(ctx context.Context, id int64) (*challengecontracts.Image, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected image lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				lookups = append(lookups, id)
				return &challengecontracts.Image{ID: id, Name: fmt.Sprintf("repo/%d", id), Tag: "latest", Status: challengecontracts.ImageStatusAvailable}, nil
			},
		},
		config: &config.Config{},
	}

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	request, err := service.buildTopologyCreateRequest(ctx, 30001, false, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{ImageID: 1}), "web", challengecontracts.TopologySpec{
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", Name: "Web", ServicePort: 8080},
			{Key: "worker", Name: "Worker", ImageID: 2, ServicePort: 9000},
		},
	}, "flag{ctx-image}")
	if err != nil {
		t.Fatalf("buildTopologyCreateRequest() error = %v", err)
	}
	if len(request.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %+v", request.Nodes)
	}
	if len(lookups) != 2 || lookups[0] != 1 || lookups[1] != 2 {
		t.Fatalf("expected image lookups [1 2], got %v", lookups)
	}
}
