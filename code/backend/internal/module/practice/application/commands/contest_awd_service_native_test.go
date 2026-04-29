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
		ID:              7901,
		ContestID:       3901,
		AWDChallengeID:  2801,
		DisplayName:     "Bank Portal",
		Order:           1,
		IsVisible:       true,
		ScoreConfig:     `{"points":180}`,
		RuntimeConfig:   `{"checker_type":"http_standard","checker_config":{"get_flag":{"path":"/ready"}}}`,
		ServiceSnapshot: `{"name":"Bank Portal","category":"web","difficulty":"medium","runtime_config":{"image_id":9901,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"awd"}}`,
		CreatedAt:       now,
		UpdatedAt:       now,
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
