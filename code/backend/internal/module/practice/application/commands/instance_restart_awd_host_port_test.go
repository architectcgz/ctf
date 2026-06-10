package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/config"
	runtimeentity "ctf-platform/internal/module/contest/entity"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
)

func TestRestartContestAWDServicePreservesHostPortWhenAccessHostConfigured(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4116)
	serviceID := int64(7116)
	contestID := int64(3116)
	userID := int64(5116)
	contestEnd := now.Add(3 * time.Hour).UTC()
	instance := &instanceentity.Instance{
		ID:             9116,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2116,
		ServiceID:      &serviceID,
		HostPort:       32116,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container","host_port":32116}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://host-gateway.internal:32116",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var cleanupInstanceID int64
	var resetStatus string
	var preserveHostPortArg bool
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2116,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":116,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			return instance, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			preserveHostPortArg = preserveHostPort
			resetStatus = status
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				cleanupInstanceID = got.ID
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup must not release preserved host port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if cleanupInstanceID != instance.ID {
		t.Fatalf("expected cleanup for instance %d, got %d", instance.ID, cleanupInstanceID)
	}
	if resetStatus != instanceentity.InstanceStatusPending {
		t.Fatalf("expected reset to pending, got %q", resetStatus)
	}
	if !preserveHostPortArg {
		t.Fatal("expected restart to preserve awd host port when access host is configured")
	}
	if !resp.ExpiresAt.Equal(contestEnd) || !instance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("restart should refresh awd instance expiry to contest end time, resp=%s instance=%s expected=%s", resp.ExpiresAt, instance.ExpiresAt, contestEnd)
	}
	if instance.HostPort != 32116 {
		t.Fatalf("expected preserved awd host port, got %+v", instance)
	}
	if instance.ContainerID != "" || instance.NetworkID != "" || instance.RuntimeDetails != "" || instance.AccessURL != "" {
		t.Fatalf("restart should clear runtime fields, got %+v", instance)
	}
}

func TestRestartContestAWDServiceAllocatesHostPortWhenAccessHostConfiguredAndInstanceHasNone(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4117)
	serviceID := int64(7117)
	contestID := int64(3117)
	userID := int64(5117)
	contestEnd := now.Add(2 * time.Hour).UTC()
	instance := &instanceentity.Instance{
		ID:             9117,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2117,
		ServiceID:      &serviceID,
		HostPort:       0,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container"}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://awd-c3117-t4117-s7117:8080",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var reserved bool
	var bound bool
	var preserveHostPortArg bool
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2117,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":117,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			return instance, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			reserved = true
			return 32117, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			bound = true
			if port != 32117 || instanceID != instance.ID {
				t.Fatalf("unexpected reserved host port bind: port=%d instance=%d", port, instanceID)
			}
			return nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			preserveHostPortArg = preserveHostPort
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup must not release newly reserved host port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				PortRangeStart:       32000,
				PortRangeEnd:         32150,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if !reserved || !bound {
		t.Fatalf("expected restart to reserve and bind a host port, reserved=%v bound=%v", reserved, bound)
	}
	if !preserveHostPortArg {
		t.Fatal("expected reset path to preserve the newly reserved host port")
	}
	if !resp.ExpiresAt.Equal(contestEnd) || !instance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("restart should refresh awd instance expiry to contest end time, resp=%s instance=%s expected=%s", resp.ExpiresAt, instance.ExpiresAt, contestEnd)
	}
	if instance.HostPort != 32117 {
		t.Fatalf("expected instance host port to be backfilled before reprovision, got %+v", instance)
	}
}

func TestRestartContestAWDServiceReallocatesStaleHostPortWhenOwnedByAnotherInstance(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4118)
	serviceID := int64(7118)
	contestID := int64(3118)
	userID := int64(5118)
	contestEnd := now.Add(2 * time.Hour).UTC()
	instance := &instanceentity.Instance{
		ID:             9118,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2118,
		ServiceID:      &serviceID,
		HostPort:       32118,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container","host_port":32118}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://host-gateway.internal:32118",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var reusableChecked bool
	var reserved bool
	var bound bool
	var preserveHostPortArg bool
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2118,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":118,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			return instance, nil
		},
		isHostPortReusableForRestartFn: func(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
			reusableChecked = true
			if instanceID != instance.ID || hostPort != 32118 {
				t.Fatalf("unexpected host port reuse check: instance=%d host_port=%d", instanceID, hostPort)
			}
			return false, nil
		},
		reserveAvailablePortExcludingFn: func(ctx context.Context, start, end, excludedPort int) (int, error) {
			reserved = true
			if excludedPort != 32118 {
				t.Fatalf("expected stale host port 32118 to be excluded, got %d", excludedPort)
			}
			return 32119, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			bound = true
			if port != 32119 || instanceID != instance.ID {
				t.Fatalf("unexpected rebound host port bind: port=%d instance=%d", port, instanceID)
			}
			return nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			preserveHostPortArg = preserveHostPort
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup must not release stale host port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				PortRangeStart:       32000,
				PortRangeEnd:         32150,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if !reusableChecked || !reserved || !bound {
		t.Fatalf("expected stale port to be checked and reallocated, checked=%v reserved=%v bound=%v", reusableChecked, reserved, bound)
	}
	if !preserveHostPortArg {
		t.Fatal("expected reset path to preserve the rebound host port")
	}
	if instance.HostPort != 32119 {
		t.Fatalf("expected stale host port to be replaced before reprovision, got %+v", instance)
	}
}
