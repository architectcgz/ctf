package infrastructure

import (
	"context"
	"testing"
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

func TestContainerInventoryRepositoryReturnsRuntimeNodeID(t *testing.T) {
	t.Parallel()

	db := newInstanceRepositoryTestDB(t)
	repo := NewContainerInventoryRepository(db)
	runtimeNodeID := int64(6201)
	instance := instancecontracts.Instance{
		ID:             62001,
		UserID:         7,
		ChallengeID:    41,
		RuntimeNodeID:  &runtimeNodeID,
		Status:         instancecontracts.InstanceStatusRunning,
		ContainerID:    "ctr-runtime-node-inventory",
		RuntimeDetails: `{"containers":[{"id":"ctr-runtime-node-inventory"}]}`,
		ExpiresAt:      time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}

	active, err := repo.ListActiveContainerInventory(context.Background())
	if err != nil {
		t.Fatalf("ListActiveContainerInventory() error = %v", err)
	}
	if len(active) != 1 || active[0].RuntimeNodeID == nil || *active[0].RuntimeNodeID != runtimeNodeID {
		t.Fatalf("active inventory runtime node id = %+v, want %d", active, runtimeNodeID)
	}

	candidates, err := repo.ListContainerNodeLookupCandidates(context.Background(), "ctr-runtime-node-inventory")
	if err != nil {
		t.Fatalf("ListContainerNodeLookupCandidates() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].RuntimeNodeID == nil || *candidates[0].RuntimeNodeID != runtimeNodeID {
		t.Fatalf("lookup candidate runtime node id = %+v, want %d", candidates, runtimeNodeID)
	}
}
