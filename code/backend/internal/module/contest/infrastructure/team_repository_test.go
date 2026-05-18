package infrastructure_test

import (
	"context"
	"testing"
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
)

func TestTeamRepositoryCreateWithMemberSyncsContestRegistration(t *testing.T) {
	t.Parallel()

	db := testsupport.SetupContestTestDB(t)
	repo := contestinfra.NewTeamRepository(db)
	now := time.Now()
	if err := db.Create(&contestentity.ContestRegistration{
		ContestID: 2,
		UserID:    2001,
		Status:    contestentity.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	team := &contestentity.Team{
		ContestID:  2,
		Name:       "Blue Team",
		CaptainID:  2001,
		InviteCode: "ABC123",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := repo.CreateWithMember(context.Background(), team, 2001); err != nil {
		t.Fatalf("CreateWithMember() error = %v", err)
	}

	var registration contestentity.ContestRegistration
	if err := db.Where("contest_id = ? AND user_id = ?", 2, 2001).First(&registration).Error; err != nil {
		t.Fatalf("load registration: %v", err)
	}
	if registration.TeamID == nil || *registration.TeamID != team.ID {
		t.Fatalf("unexpected team binding: %+v", registration)
	}
	if registration.Status != contestentity.ContestRegistrationStatusApproved {
		t.Fatalf("unexpected registration status: %s", registration.Status)
	}
}
