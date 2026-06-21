package composition

import (
	"context"
	"errors"
	"testing"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type practiceRuntimeNodeSelectorDefaultStub struct {
	calls   int
	binding *runtimecontracts.RuntimeNodeBinding
	err     error
}

func (s *practiceRuntimeNodeSelectorDefaultStub) SelectDefaultNode(context.Context) (*runtimecontracts.RuntimeNodeBinding, error) {
	s.calls++
	return s.binding, s.err
}

type practiceRuntimeNodePlacementStoreStub struct {
	findBinding   *practiceports.RuntimeNodeBinding
	findExists    bool
	findErr       error
	ensureBinding *practiceports.RuntimeNodeBinding
	ensureErr     error
	ensureCalls   int
}

func (s *practiceRuntimeNodePlacementStoreStub) FindActiveContestRuntimePlacement(context.Context, int64) (*practiceports.RuntimeNodeBinding, bool, error) {
	return s.findBinding, s.findExists, s.findErr
}

func (s *practiceRuntimeNodePlacementStoreStub) EnsureActiveContestRuntimePlacement(_ context.Context, _ int64, _ int64) (*practiceports.RuntimeNodeBinding, error) {
	s.ensureCalls++
	return s.ensureBinding, s.ensureErr
}

type practiceRuntimeNodeHealthLookupStub struct {
	calls   int
	binding *practiceports.RuntimeNodeBinding
	err     error
}

func (s *practiceRuntimeNodeHealthLookupStub) FindHealthyRuntimeNodeByID(context.Context, int64) (*practiceports.RuntimeNodeBinding, error) {
	s.calls++
	return s.binding, s.err
}

func TestPracticeRuntimeNodeSelectorAdapterUsesExistingAWDPlacement(t *testing.T) {
	t.Parallel()

	defaultSelector := &practiceRuntimeNodeSelectorDefaultStub{binding: &runtimecontracts.RuntimeNodeBinding{NodeID: 202, NodeName: "node-b"}}
	placementStore := &practiceRuntimeNodePlacementStoreStub{
		findBinding: &practiceports.RuntimeNodeBinding{RuntimeNodeID: 101, NodeName: "node-a"},
		findExists:  true,
	}
	healthLookup := &practiceRuntimeNodeHealthLookupStub{binding: &practiceports.RuntimeNodeBinding{RuntimeNodeID: 101, NodeName: "node-a"}}
	adapter := newPracticeRuntimeNodeSelectorAdapter(defaultSelector, placementStore, healthLookup)
	contestID := int64(9001)

	binding, err := adapter.SelectRuntimeNode(context.Background(), practiceports.InstanceScope{
		ContestID:   &contestID,
		ContestMode: "awd",
	})
	if err != nil {
		t.Fatalf("SelectRuntimeNode() error = %v", err)
	}
	if binding == nil || binding.RuntimeNodeID != 101 {
		t.Fatalf("binding = %+v, want placement node 101", binding)
	}
	if defaultSelector.calls != 0 {
		t.Fatalf("default selector calls = %d, want 0", defaultSelector.calls)
	}
}

func TestPracticeRuntimeNodeSelectorAdapterRejectsUnhealthyAWDPlacementWithoutFallback(t *testing.T) {
	t.Parallel()

	defaultSelector := &practiceRuntimeNodeSelectorDefaultStub{binding: &runtimecontracts.RuntimeNodeBinding{NodeID: 202, NodeName: "node-b"}}
	placementStore := &practiceRuntimeNodePlacementStoreStub{
		findBinding: &practiceports.RuntimeNodeBinding{RuntimeNodeID: 101, NodeName: "node-a"},
		findExists:  true,
	}
	healthLookup := &practiceRuntimeNodeHealthLookupStub{err: runtimeports.ErrRuntimeNodeUnavailable}
	adapter := newPracticeRuntimeNodeSelectorAdapter(defaultSelector, placementStore, healthLookup)
	contestID := int64(9001)

	_, err := adapter.SelectRuntimeNode(context.Background(), practiceports.InstanceScope{
		ContestID:   &contestID,
		ContestMode: "awd",
	})
	if !errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) {
		t.Fatalf("SelectRuntimeNode() error = %v, want ErrRuntimeNodeUnavailable", err)
	}
	if defaultSelector.calls != 0 {
		t.Fatalf("default selector calls = %d, want 0", defaultSelector.calls)
	}
}

func TestPracticeRuntimeNodeSelectorAdapterRejectsAWDWhenPlacementDependenciesMissing(t *testing.T) {
	t.Parallel()

	defaultSelector := &practiceRuntimeNodeSelectorDefaultStub{binding: &runtimecontracts.RuntimeNodeBinding{NodeID: 202, NodeName: "node-b"}}
	adapter := newPracticeRuntimeNodeSelectorAdapter(defaultSelector, nil, nil)
	contestID := int64(9001)

	_, err := adapter.SelectRuntimeNode(context.Background(), practiceports.InstanceScope{
		ContestID:   &contestID,
		ContestMode: "awd",
	})
	if !errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) {
		t.Fatalf("SelectRuntimeNode() error = %v, want ErrRuntimeNodeUnavailable", err)
	}
	if defaultSelector.calls != 0 {
		t.Fatalf("default selector calls = %d, want 0", defaultSelector.calls)
	}
}

func TestPracticeRuntimeNodeSelectorAdapterCreatesAWDPlacementFromDefaultSelection(t *testing.T) {
	t.Parallel()

	defaultSelector := &practiceRuntimeNodeSelectorDefaultStub{binding: &runtimecontracts.RuntimeNodeBinding{NodeID: 202, NodeName: "node-b"}}
	placementStore := &practiceRuntimeNodePlacementStoreStub{
		ensureBinding: &practiceports.RuntimeNodeBinding{RuntimeNodeID: 202, NodeName: "node-b"},
	}
	healthLookup := &practiceRuntimeNodeHealthLookupStub{binding: &practiceports.RuntimeNodeBinding{RuntimeNodeID: 202, NodeName: "node-b"}}
	adapter := newPracticeRuntimeNodeSelectorAdapter(defaultSelector, placementStore, healthLookup)
	contestID := int64(9001)

	binding, err := adapter.SelectRuntimeNode(context.Background(), practiceports.InstanceScope{
		ContestID:   &contestID,
		ContestMode: "awd",
	})
	if err != nil {
		t.Fatalf("SelectRuntimeNode() error = %v", err)
	}
	if binding == nil || binding.RuntimeNodeID != 202 {
		t.Fatalf("binding = %+v, want default node 202 placement", binding)
	}
	if defaultSelector.calls != 1 || placementStore.ensureCalls != 1 {
		t.Fatalf("default calls = %d ensure calls = %d, want 1/1", defaultSelector.calls, placementStore.ensureCalls)
	}
}

func TestPracticeRuntimeNodeSelectorAdapterUsesDefaultForNonAWDScope(t *testing.T) {
	t.Parallel()

	defaultSelector := &practiceRuntimeNodeSelectorDefaultStub{binding: &runtimecontracts.RuntimeNodeBinding{NodeID: 303, NodeName: "node-c"}}
	adapter := newPracticeRuntimeNodeSelectorAdapter(defaultSelector, nil, nil)

	binding, err := adapter.SelectRuntimeNode(context.Background(), practiceports.InstanceScope{ContestMode: "jeopardy"})
	if err != nil {
		t.Fatalf("SelectRuntimeNode() error = %v", err)
	}
	if binding == nil || binding.RuntimeNodeID != 303 {
		t.Fatalf("binding = %+v, want default node 303", binding)
	}
}
