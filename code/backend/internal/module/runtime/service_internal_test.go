package runtime

import (
	"testing"
	"time"

	"ctf-platform/internal/config"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

func TestSelectOrphanContainersSkipsActiveAndGracePeriod(t *testing.T) {
	t.Parallel()

	now := time.Now()
	managedContainers := []ManagedContainer{
		{ID: "active", Name: "ctf-instance-active", CreatedAt: now.Add(-10 * time.Minute)},
		{ID: "fresh", Name: "ctf-instance-fresh", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "orphan", Name: "ctf-instance-orphan", CreatedAt: now.Add(-12 * time.Minute)},
	}
	activeContainerIDs := map[string]struct{}{
		"active": {},
	}

	orphanContainers := selectOrphanContainers(managedContainers, activeContainerIDs, 5*time.Minute, now)
	if len(orphanContainers) != 1 {
		t.Fatalf("expected 1 orphan container, got %d (%v)", len(orphanContainers), orphanContainers)
	}
	if orphanContainers[0].ID != "orphan" {
		t.Fatalf("unexpected orphan container: %+v", orphanContainers[0])
	}
}

func TestManagedContainerLabels(t *testing.T) {
	t.Parallel()

	labels := managedContainerLabels()
	if labels[managedByLabelKey] != managedByLabelValue {
		t.Fatalf("expected managed-by label, got %v", labels)
	}
	if labels[challengeInstanceLabelKey] != challengeInstanceLabelValue {
		t.Fatalf("expected component label, got %v", labels)
	}
}

func TestManagedNetworkLabels(t *testing.T) {
	t.Parallel()

	labels := managedNetworkLabels()
	if labels[managedByLabelKey] != managedByLabelValue {
		t.Fatalf("expected managed-by label, got %v", labels)
	}
	if labels[challengeInstanceLabelKey] != challengeInstanceLabelValue {
		t.Fatalf("expected component label, got %v", labels)
	}
}

func TestNewServiceTreatsTypedNilEngineAsNil(t *testing.T) {
	t.Parallel()

	cfg := &config.ContainerConfig{
		PortRangeStart:       30000,
		PortRangeEnd:         30010,
		DefaultExposedPort:   8080,
		PublicHost:           "127.0.0.1",
		DefaultTTL:           time.Hour,
		MaxExtends:           2,
		MaxConcurrentPerUser: 3,
		CreateTimeout:        time.Second,
	}

	var typedNil *runtimeinfra.Engine
	service := NewService(nil, typedNil, cfg, nil)
	if service.engine != nil {
		t.Fatalf("expected typed nil engine to be normalized to nil, got %#v", service.engine)
	}
}
