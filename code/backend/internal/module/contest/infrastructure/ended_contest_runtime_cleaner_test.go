package infrastructure

import (
	"context"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
)

type endedContestRuntimeCleanupStub struct {
	cleaned []*model.Instance
	err     error
}

func (s *endedContestRuntimeCleanupStub) CleanupRuntime(_ context.Context, instance *model.Instance) error {
	copied := *instance
	s.cleaned = append(s.cleaned, &copied)
	return s.err
}

func TestContestEndedRuntimeCleanerCleansOnlyCurrentContestAWDInstances(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupAWDTestDB(t)
	if err := db.AutoMigrate(&model.PortAllocation{}); err != nil {
		t.Fatalf("auto migrate port allocations: %v", err)
	}
	now := time.Now().UTC()
	contestID := int64(81)
	otherContestID := int64(82)
	serviceID := int64(91)
	secondServiceID := int64(92)
	otherServiceID := int64(93)
	teamID := int64(101)
	secondTeamID := int64(102)

	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Containers: []model.InstanceRuntimeContainer{
			{ContainerID: "ctr-runtime-details", HostPort: 32012, IsEntryPoint: true},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	for _, contest := range []model.Contest{
		{ID: contestID, Title: "ended-awd", Mode: model.ContestModeAWD, Status: model.ContestStatusEnded, StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: otherContestID, Title: "other-awd", Mode: model.ContestModeAWD, Status: model.ContestStatusEnded, StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Minute), CreatedAt: now, UpdatedAt: now},
	} {
		if err := db.Create(&contest).Error; err != nil {
			t.Fatalf("create contest: %v", err)
		}
	}

	for _, service := range []model.ContestAWDService{
		{ID: serviceID, ContestID: contestID, AWDChallengeID: 201, DisplayName: "svc-a", Order: 1, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{ID: secondServiceID, ContestID: contestID, AWDChallengeID: 202, DisplayName: "svc-b", Order: 2, IsVisible: true, CreatedAt: now, UpdatedAt: now},
		{ID: otherServiceID, ContestID: otherContestID, AWDChallengeID: 203, DisplayName: "svc-c", Order: 1, IsVisible: true, CreatedAt: now, UpdatedAt: now},
	} {
		if err := db.Create(&service).Error; err != nil {
			t.Fatalf("create awd service: %v", err)
		}
	}

	for _, instance := range []model.Instance{
		{
			ID:          1001,
			UserID:      1,
			ContestID:   &contestID,
			TeamID:      &teamID,
			ServiceID:   &serviceID,
			ChallengeID: 201,
			HostPort:    32011,
			ContainerID: "ctr-legacy",
			NetworkID:   "net-legacy",
			Status:      model.InstanceStatusRunning,
			AccessURL:   "http://127.0.0.1:32011",
			ExpiresAt:   now.Add(time.Hour),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:             1002,
			UserID:         2,
			ContestID:      &contestID,
			TeamID:         &secondTeamID,
			ServiceID:      &secondServiceID,
			ChallengeID:    202,
			Status:         model.InstanceStatusFailed,
			RuntimeDetails: runtimeDetails,
			AccessURL:      "http://awd-c81-t102-s92:8080",
			ExpiresAt:      now.Add(time.Hour),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:          1003,
			UserID:      3,
			ContestID:   &contestID,
			TeamID:      &teamID,
			ServiceID:   &serviceID,
			ChallengeID: 201,
			Status:      model.InstanceStatusStopped,
			ExpiresAt:   now.Add(time.Hour),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          1004,
			UserID:      4,
			ContestID:   &otherContestID,
			TeamID:      &teamID,
			ServiceID:   &otherServiceID,
			ChallengeID: 203,
			ContainerID: "ctr-other",
			Status:      model.InstanceStatusRunning,
			ExpiresAt:   now.Add(time.Hour),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	} {
		if err := db.Create(&instance).Error; err != nil {
			t.Fatalf("create instance: %v", err)
		}
	}

	for _, allocation := range []model.PortAllocation{
		{Port: 32011, InstanceID: int64Ptr(1001), CreatedAt: now, UpdatedAt: now},
		{Port: 32012, InstanceID: int64Ptr(1002), CreatedAt: now, UpdatedAt: now},
		{Port: 32013, InstanceID: int64Ptr(1004), CreatedAt: now, UpdatedAt: now},
	} {
		if err := db.Create(&allocation).Error; err != nil {
			t.Fatalf("create port allocation: %v", err)
		}
	}

	for _, workspace := range []model.AWDDefenseWorkspace{
		{
			ContestID:         contestID,
			TeamID:            teamID,
			ServiceID:         serviceID,
			InstanceID:        1001,
			WorkspaceRevision: 7,
			Status:            model.AWDDefenseWorkspaceStatusRunning,
			ContainerID:       "workspace-ctr-team-a",
			SeedSignature:     "seed:team-a",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ContestID:         otherContestID,
			TeamID:            teamID,
			ServiceID:         otherServiceID,
			InstanceID:        1004,
			WorkspaceRevision: 3,
			Status:            model.AWDDefenseWorkspaceStatusRunning,
			ContainerID:       "workspace-ctr-other",
			SeedSignature:     "seed:other",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	} {
		if err := db.Create(&workspace).Error; err != nil {
			t.Fatalf("create defense workspace: %v", err)
		}
	}

	for _, operation := range []model.AWDServiceOperation{
		{
			ID:            2001,
			ContestID:     contestID,
			TeamID:        teamID,
			ServiceID:     serviceID,
			InstanceID:    1001,
			OperationType: model.AWDServiceOperationTypeRestart,
			RequestedBy:   model.AWDServiceOperationRequestedByUser,
			Reason:        "user_restart",
			SLABillable:   true,
			Status:        model.AWDServiceOperationStatusProvisioning,
			StartedAt:     now.Add(-3 * time.Minute),
			CreatedAt:     now.Add(-3 * time.Minute),
			UpdatedAt:     now.Add(-2 * time.Minute),
		},
		{
			ID:            2002,
			ContestID:     contestID,
			TeamID:        secondTeamID,
			ServiceID:     secondServiceID,
			InstanceID:    1002,
			OperationType: model.AWDServiceOperationTypeRecover,
			RequestedBy:   model.AWDServiceOperationRequestedBySystem,
			Reason:        "container_not_running",
			SLABillable:   false,
			Status:        model.AWDServiceOperationStatusRecovering,
			StartedAt:     now.Add(-4 * time.Minute),
			CreatedAt:     now.Add(-4 * time.Minute),
			UpdatedAt:     now.Add(-3 * time.Minute),
		},
		{
			ID:            2003,
			ContestID:     otherContestID,
			TeamID:        teamID,
			ServiceID:     otherServiceID,
			InstanceID:    1004,
			OperationType: model.AWDServiceOperationTypeRestart,
			RequestedBy:   model.AWDServiceOperationRequestedByUser,
			Reason:        "other_contest",
			SLABillable:   true,
			Status:        model.AWDServiceOperationStatusProvisioning,
			StartedAt:     now.Add(-2 * time.Minute),
			CreatedAt:     now.Add(-2 * time.Minute),
			UpdatedAt:     now.Add(-time.Minute),
		},
	} {
		if err := db.Create(&operation).Error; err != nil {
			t.Fatalf("create awd service operation: %v", err)
		}
	}

	runtimeCleaner := &endedContestRuntimeCleanupStub{}
	awdRepo := NewAWDRepository(db)
	runtimeRepo := runtimeinfra.NewRepository(db)
	cleaner := NewContestEndedRuntimeCleaner(awdRepo, awdRepo, runtimeCleaner, runtimeRepo)

	if err := cleaner.CleanupEndedContestAWDInstances(context.Background(), contestID); err != nil {
		t.Fatalf("CleanupEndedContestAWDInstances() error = %v", err)
	}

	if len(runtimeCleaner.cleaned) != 3 {
		t.Fatalf("expected 3 cleaned runtime payloads, got %d", len(runtimeCleaner.cleaned))
	}
	cleanedContainers := collectCleanedContainerIDs(t, runtimeCleaner.cleaned)
	for _, containerID := range []string{"ctr-legacy", "ctr-runtime-details", "workspace-ctr-team-a"} {
		if _, ok := cleanedContainers[containerID]; !ok {
			t.Fatalf("expected cleanup to include container %q, got %+v", containerID, cleanedContainers)
		}
	}
	if _, ok := cleanedContainers["workspace-ctr-other"]; ok {
		t.Fatalf("expected other contest workspace container to stay untouched, got %+v", cleanedContainers)
	}
	if runtimeCleaner.cleaned[0].ID == 0 && runtimeCleaner.cleaned[1].ID == 0 && runtimeCleaner.cleaned[2].ID == 0 {
		t.Fatalf("expected cleanup payloads to preserve instance ids for diagnostics, got %+v", runtimeCleaner.cleaned)
	}

	for _, instanceID := range []int64{1001, 1002} {
		var row struct {
			Status         string     `gorm:"column:status"`
			HostPort       int        `gorm:"column:host_port"`
			ContainerID    string     `gorm:"column:container_id"`
			NetworkID      string     `gorm:"column:network_id"`
			RuntimeDetails string     `gorm:"column:runtime_details"`
			AccessURL      string     `gorm:"column:access_url"`
			DestroyedAt    *time.Time `gorm:"column:destroyed_at"`
		}
		if err := db.Table("instances").
			Select("status", "host_port", "container_id", "network_id", "runtime_details", "access_url", "destroyed_at").
			Where("id = ?", instanceID).
			Take(&row).Error; err != nil {
			t.Fatalf("load expired instance %d: %v", instanceID, err)
		}
		if row.Status != model.InstanceStatusExpired {
			t.Fatalf("instance %d status = %q, want %q", instanceID, row.Status, model.InstanceStatusExpired)
		}
		if row.HostPort != 0 || row.ContainerID != "" || row.NetworkID != "" || row.RuntimeDetails != "" || row.AccessURL != "" {
			t.Fatalf("expected instance %d runtime fields to be cleared, got %+v", instanceID, row)
		}
		if row.DestroyedAt == nil {
			t.Fatalf("expected instance %d destroyed_at to be set", instanceID)
		}
	}

	var stoppedInstance model.Instance
	if err := db.First(&stoppedInstance, 1003).Error; err != nil {
		t.Fatalf("load stopped instance: %v", err)
	}
	if stoppedInstance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped instance to remain stopped, got %+v", stoppedInstance)
	}

	var otherContestInstance model.Instance
	if err := db.First(&otherContestInstance, 1004).Error; err != nil {
		t.Fatalf("load other contest instance: %v", err)
	}
	if otherContestInstance.Status != model.InstanceStatusRunning || otherContestInstance.ContainerID != "ctr-other" {
		t.Fatalf("expected other contest instance to stay untouched, got %+v", otherContestInstance)
	}

	var workspace model.AWDDefenseWorkspace
	if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).First(&workspace).Error; err != nil {
		t.Fatalf("load ended contest workspace: %v", err)
	}
	if workspace.InstanceID != 1001 || workspace.WorkspaceRevision != 7 {
		t.Fatalf("expected workspace identity to stay scoped, got %+v", workspace)
	}
	if workspace.Status != model.AWDDefenseWorkspaceStatusFailed || workspace.ContainerID != "" {
		t.Fatalf("expected workspace runtime to be cleared into failed state, got %+v", workspace)
	}

	var otherWorkspace model.AWDDefenseWorkspace
	if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ?", otherContestID, teamID, otherServiceID).First(&otherWorkspace).Error; err != nil {
		t.Fatalf("load other contest workspace: %v", err)
	}
	if otherWorkspace.Status != model.AWDDefenseWorkspaceStatusRunning || otherWorkspace.ContainerID != "workspace-ctr-other" {
		t.Fatalf("expected other contest workspace to stay untouched, got %+v", otherWorkspace)
	}

	for _, operationID := range []int64{2001, 2002} {
		var operation model.AWDServiceOperation
		if err := db.First(&operation, operationID).Error; err != nil {
			t.Fatalf("load ended contest operation %d: %v", operationID, err)
		}
		if operation.Status != model.AWDServiceOperationStatusFailed {
			t.Fatalf("operation %d status = %q, want %q", operationID, operation.Status, model.AWDServiceOperationStatusFailed)
		}
		if operation.ErrorMessage != "contest_ended" {
			t.Fatalf("operation %d error_message = %q, want contest_ended", operationID, operation.ErrorMessage)
		}
		if operation.FinishedAt == nil {
			t.Fatalf("expected operation %d finished_at to be set", operationID)
		}
	}

	var otherContestOperation model.AWDServiceOperation
	if err := db.First(&otherContestOperation, 2003).Error; err != nil {
		t.Fatalf("load other contest operation: %v", err)
	}
	if otherContestOperation.Status != model.AWDServiceOperationStatusProvisioning || otherContestOperation.FinishedAt != nil {
		t.Fatalf("expected other contest operation to stay active, got %+v", otherContestOperation)
	}

	var remainingEndedContestAllocations int64
	if err := db.Model(&model.PortAllocation{}).
		Where("instance_id IN ?", []int64{1001, 1002}).
		Count(&remainingEndedContestAllocations).Error; err != nil {
		t.Fatalf("count ended contest port allocations: %v", err)
	}
	if remainingEndedContestAllocations != 0 {
		t.Fatalf("expected ended contest port allocations to be released, got %d", remainingEndedContestAllocations)
	}

	var otherContestAllocations int64
	if err := db.Model(&model.PortAllocation{}).
		Where("instance_id = ?", 1004).
		Count(&otherContestAllocations).Error; err != nil {
		t.Fatalf("count other contest port allocations: %v", err)
	}
	if otherContestAllocations != 1 {
		t.Fatalf("expected other contest port allocation to stay, got %d", otherContestAllocations)
	}
}

func collectCleanedContainerIDs(t *testing.T, instances []*model.Instance) map[string]struct{} {
	t.Helper()

	result := make(map[string]struct{})
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		if trimmed := strings.TrimSpace(instance.ContainerID); trimmed != "" {
			result[trimmed] = struct{}{}
		}
		if strings.TrimSpace(instance.RuntimeDetails) == "" {
			continue
		}
		details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
		if err != nil {
			t.Fatalf("decode runtime details for cleaned instance %d: %v", instance.ID, err)
		}
		for _, container := range details.Containers {
			if trimmed := strings.TrimSpace(container.ContainerID); trimmed != "" {
				result[trimmed] = struct{}{}
			}
		}
	}
	return result
}

func int64Ptr(value int64) *int64 {
	return &value
}
