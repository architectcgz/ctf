package model

import "testing"

func TestAWDDefenseWorkspaceTableName(t *testing.T) {
	if got := (AWDDefenseWorkspace{}).TableName(); got != "awd_defense_workspaces" {
		t.Fatalf("AWDDefenseWorkspace.TableName() = %q, want %q", got, "awd_defense_workspaces")
	}
}
