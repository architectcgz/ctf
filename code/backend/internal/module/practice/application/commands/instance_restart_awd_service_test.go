package commands

import (
	"context"
	"ctf-platform/internal/config"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	runtimeentity "ctf-platform/internal/module/contest/entity"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	"testing"
	"time"
)

func TestRestartOrStartScopedAWDServiceRecreatesActiveInstanceWhenCheckerTokenMetadataMissing(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	contestID := int64(3319)
	teamID := int64(4319)
	serviceID := int64(7319)
	challengeID := int64(2319)
	userID := int64(5319)
	existingInstance := &instanceentity.Instance{
		ID:          9319,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(10 * time.Minute),
		RuntimeDetails: `{"networks":[{"name":"ctf-awd-contest-3319","shared":true}],
			"containers":[{"container_id":"ctr-3319","is_entry_point":true,"service_port":8080}]}`,
	}
	scope := practiceports.InstanceScope{
		ContestMode: practiceports.ContestModeAWD,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
	}

	cleanupCalled := false
	resetCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest lookup: %d", gotContestID)
			}
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: now.Add(time.Hour),
			}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, gotScope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != challengeID {
				t.Fatalf("unexpected restartable lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			if gotScope.ServiceID == nil || *gotScope.ServiceID != serviceID {
				t.Fatalf("unexpected scope: %+v", gotScope)
			}
			return existingInstance, nil
		},
		findContestAWDServiceRuntimeSubjectFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected runtime subject lookup: contest=%d service=%d", gotContestID, gotServiceID)
			}
			return &practiceports.ContestAWDServiceRuntimeSubject{
				ServiceID:   serviceID,
				ChallengeID: challengeID,
				Visible:     true,
				WorkspaceConfig: &practiceports.ContestAWDDefenseWorkspaceConfig{
					CheckerTokenEnv: "CHECKER_TOKEN",
				},
			}, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			resetCalled = true
			if instanceID != existingInstance.ID {
				t.Fatalf("unexpected reset instance: %d", instanceID)
			}
			if status != instanceentity.InstanceStatusPending {
				t.Fatalf("unexpected reset status: %s", status)
			}
			if preserveHostPort {
				t.Fatal("expected awd restart to avoid preserving host port")
			}
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, operation *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(newServiceCore(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
				cleanupCalled = true
				if instance.ID != existingInstance.ID {
					t.Fatalf("unexpected cleanup instance: %+v", instance)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, nil)

	resp, err := service.restartOrStartScopedAWDService(context.Background(), awdScopedRuntimeRequest{
		OwnerUserID:  userID,
		ContestID:    contestID,
		ChallengeID:  challengeID,
		Scope:        scope,
		NoopIfActive: true,
		Audit: awdScopedRuntimeAudit{
			RestartOperationType: contestcontracts.AWDServiceOperationTypeRecreate,
			RequestedBy:          contestcontracts.AWDServiceOperationRequestedBySystem,
			Reason:               "desired_runtime_reconcile",
		},
	})
	if err != nil {
		t.Fatalf("restartOrStartScopedAWDService() error = %v", err)
	}
	if resp == nil || resp.ID != existingInstance.ID {
		t.Fatalf("expected restarted instance response, got %+v", resp)
	}
	if !cleanupCalled || !resetCalled {
		t.Fatalf("expected active instance to be recreated, cleanup=%v reset=%v", cleanupCalled, resetCalled)
	}
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending status after restart scheduling, got %+v", resp)
	}
}

func TestRestartContestAWDServiceRequeuesExistingTeamInstance(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4106)
	serviceID := int64(7106)
	contestID := int64(3106)
	userID := int64(5106)
	contestEnd := now.Add(4 * time.Hour).UTC()
	instance := &instanceentity.Instance{
		ID:             9106,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2106,
		ServiceID:      &serviceID,
		HostPort:       32106,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container"}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:32106",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var cleanupInstanceID int64
	var resetStatus string
	var operation *runtimeentity.AWDServiceOperation
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", gotContestID, gotServiceID)
			}
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2106,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":106,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != 2106 {
				t.Fatalf("unexpected scoped lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			if scope.ServiceID == nil || *scope.ServiceID != serviceID || scope.TeamID == nil || *scope.TeamID != teamID || scope.ShareScope != instanceentity.ShareScopePerTeam {
				t.Fatalf("unexpected restart scope: %+v", scope)
			}
			return instance, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if instanceID != instance.ID {
				t.Fatalf("unexpected reset instance id: %d", instanceID)
			}
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			if preserveHostPort {
				t.Fatal("AWD restart must clear historical host port instead of preserving it")
			}
			resetStatus = status
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			operation = got
			return nil
		},
	}

	service := wirePracticeScopeAdapters(newServiceCore(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				cleanupInstanceID = got.ID
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup should preserve port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
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
	if !resp.ExpiresAt.Equal(contestEnd) || !instance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("restart should refresh awd instance expiry to contest end time, resp=%s instance=%s expected=%s", resp.ExpiresAt, instance.ExpiresAt, contestEnd)
	}
	if instance.ServiceID == nil || *instance.ServiceID != serviceID || instance.Nonce != "nonce-keep" || instance.HostPort != 0 {
		t.Fatalf("restart should preserve identity fields, got %+v", instance)
	}
	if instance.ContainerID != "" || instance.NetworkID != "" || instance.RuntimeDetails != "" || instance.AccessURL != "" {
		t.Fatalf("restart should clear runtime fields, got %+v", instance)
	}
	if operation == nil || operation.OperationType != runtimeentity.AWDServiceOperationTypeRestart || operation.RequestedBy != runtimeentity.AWDServiceOperationRequestedByUser || !operation.SLABillable {
		t.Fatalf("expected billable user restart operation, got %+v", operation)
	}
}
