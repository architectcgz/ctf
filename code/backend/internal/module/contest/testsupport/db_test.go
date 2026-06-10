package testsupport

import (
	"testing"

	runtimeentity "ctf-platform/internal/module/contest/entity"
)

func TestSetupAWDTestDBAutoMigratesAWDDefenseWorkspace(t *testing.T) {
	t.Parallel()

	db := SetupAWDTestDB(t)
	workspace := &runtimeentity.AWDDefenseWorkspace{
		ContestID:         1001,
		TeamID:            1002,
		ServiceID:         1003,
		InstanceID:        1004,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr-auto",
		SeedSignature:     "seed:auto",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("create defense workspace in awd test db: %v", err)
	}

	var stored runtimeentity.AWDDefenseWorkspace
	if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ?", 1001, 1002, 1003).First(&stored).Error; err != nil {
		t.Fatalf("load defense workspace from awd test db: %v", err)
	}
	if stored.ContainerID != "workspace-ctr-auto" || stored.WorkspaceRevision != 1 {
		t.Fatalf("unexpected stored workspace: %+v", stored)
	}
}
