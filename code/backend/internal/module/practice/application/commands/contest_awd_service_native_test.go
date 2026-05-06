package commands_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
)

func TestServiceStartContestAWDServiceCanProvisionFromContestAWDServiceSnapshot(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceAWDContest(t, db, 3901, 0, now)
	seedContestInstanceTeam(t, db, 3901, 4901, 5901, now)
	seedContestInstanceRegistration(t, db, 3901, 5901, 4901, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, 3901, 4901, 5901, now)

	if err := db.Create(&model.Image{
		ID:          9901,
		Name:        "ctf/bank-portal",
		Tag:         "v1",
		Status:      model.ImageStatusAvailable,
		Description: "bank portal",
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.ContestAWDService{
		ID:             7901,
		ContestID:      3901,
		AWDChallengeID: 2801,
		DisplayName:    "Bank Portal",
		Order:          1,
		IsVisible:      true,
		ScoreConfig:    `{"points":180}`,
		RuntimeConfig:  `{"checker_type":"http_standard","checker_config":{"get_flag":{"path":"/ready"}}}`,
		ServiceSnapshot: mustEncodeContestInstanceAWDServiceSnapshot(t, model.ContestAWDServiceSnapshot{
			Name:       "Bank Portal",
			Category:   "web",
			Difficulty: "medium",
			RuntimeConfig: map[string]any{
				"image_id":         int64(9901),
				"instance_sharing": string(model.InstanceSharingPerTeam),
				"defense_workspace": map[string]any{
					"entry_mode":      "ssh",
					"seed_root":       "runtime/workspace",
					"workspace_roots": []string{"runtime/workspace/app"},
					"writable_roots":  []string{"runtime/workspace/app"},
					"readonly_roots":  []string{},
					"runtime_mounts": []map[string]any{
						{"source": "runtime/workspace/app", "target": "/workspace/app", "mode": "rw"},
					},
				},
			},
			FlagConfig: map[string]any{
				"flag_type":   model.FlagTypeStatic,
				"flag_prefix": "awd",
			},
		}),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	resp, err := service.StartContestAWDService(context.Background(), 5901, 3901, 7901)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID == 0 || resp.ChallengeID != 2801 {
		t.Fatalf("expected awd snapshot-backed instance, got %+v", resp)
	}
}
