package testsupport

import (
	"testing"

	"ctf-platform/internal/model"
)

func TestSetupAWDTestDBAutoMigratesAWDDefenseWorkspace(t *testing.T) {
	t.Parallel()

	db := SetupAWDTestDB(t)
	workspace := &model.AWDDefenseWorkspace{
		ContestID:         1001,
		TeamID:            1002,
		ServiceID:         1003,
		InstanceID:        1004,
		WorkspaceRevision: 1,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr-auto",
		SeedSignature:     "seed:auto",
	}
	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("create defense workspace in awd test db: %v", err)
	}

	var stored model.AWDDefenseWorkspace
	if err := db.Where("contest_id = ? AND team_id = ? AND service_id = ?", 1001, 1002, 1003).First(&stored).Error; err != nil {
		t.Fatalf("load defense workspace from awd test db: %v", err)
	}
	if stored.ContainerID != "workspace-ctr-auto" || stored.WorkspaceRevision != 1 {
		t.Fatalf("unexpected stored workspace: %+v", stored)
	}
}
