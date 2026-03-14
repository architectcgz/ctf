package contest

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func TestParticipationServiceRegisterContestCreatesPendingRegistration(t *testing.T) {
	t.Parallel()

	db := newContestTestDB(t)
	contestRepo := NewRepository(db)
	teamRepo := NewTeamRepository(db)
	service := NewParticipationService(db, contestRepo, teamRepo)

	now := time.Now()
	if err := db.Create(&model.Contest{
		ID:        1,
		Title:     "spring-ctf",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		Status:    model.ContestStatusRegistration,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	if err := service.RegisterContest(context.Background(), 1, 1001); err != nil {
		t.Fatalf("RegisterContest() error = %v", err)
	}
	if err := service.RegisterContest(context.Background(), 1, 1001); err != nil {
		t.Fatalf("RegisterContest() second call error = %v", err)
	}

	var registration model.ContestRegistration
	if err := db.Where("contest_id = ? AND user_id = ?", 1, 1001).First(&registration).Error; err != nil {
		t.Fatalf("load registration: %v", err)
	}
	if registration.Status != model.ContestRegistrationStatusPending {
		t.Fatalf("unexpected registration status: %s", registration.Status)
	}
	if registration.TeamID != nil {
		t.Fatalf("expected nil team id, got %v", *registration.TeamID)
	}
}

func TestParticipationServiceRegisterContestRequeuesRejectedRegistration(t *testing.T) {
	t.Parallel()

	db := newContestTestDB(t)
	contestRepo := NewRepository(db)
	teamRepo := NewTeamRepository(db)
	service := NewParticipationService(db, contestRepo, teamRepo)

	now := time.Now()
	reviewedBy := int64(9001)
	reviewedAt := now.Add(-time.Hour)
	if err := db.Create(&model.Contest{
		ID:        10,
		Title:     "retry-ctf",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		Status:    model.ContestStatusRegistration,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestRegistration{
		ContestID:  10,
		UserID:     1002,
		Status:     model.ContestRegistrationStatusRejected,
		ReviewedBy: &reviewedBy,
		ReviewedAt: &reviewedAt,
		CreatedAt:  now.Add(-2 * time.Hour),
		UpdatedAt:  now.Add(-time.Hour),
	}).Error; err != nil {
		t.Fatalf("create rejected registration: %v", err)
	}

	if err := service.RegisterContest(context.Background(), 10, 1002); err != nil {
		t.Fatalf("RegisterContest() error = %v", err)
	}

	var registration model.ContestRegistration
	if err := db.Where("contest_id = ? AND user_id = ?", 10, 1002).First(&registration).Error; err != nil {
		t.Fatalf("load registration: %v", err)
	}
	if registration.Status != model.ContestRegistrationStatusPending {
		t.Fatalf("unexpected registration status: %s", registration.Status)
	}
	if registration.ReviewedBy != nil || registration.ReviewedAt != nil {
		t.Fatalf("expected review metadata reset, got %+v", registration)
	}
}

func TestTeamRepositoryCreateWithMemberSyncsContestRegistration(t *testing.T) {
	t.Parallel()

	db := newContestTestDB(t)
	repo := NewTeamRepository(db)
	now := time.Now()
	if err := db.Create(&model.ContestRegistration{
		ContestID: 2,
		UserID:    2001,
		Status:    model.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	team := &model.Team{
		ContestID:  2,
		Name:       "Blue Team",
		CaptainID:  2001,
		InviteCode: "ABC123",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := repo.CreateWithMember(team, 2001); err != nil {
		t.Fatalf("CreateWithMember() error = %v", err)
	}

	var registration model.ContestRegistration
	if err := db.Where("contest_id = ? AND user_id = ?", 2, 2001).First(&registration).Error; err != nil {
		t.Fatalf("load registration: %v", err)
	}
	if registration.TeamID == nil || *registration.TeamID != team.ID {
		t.Fatalf("unexpected team binding: %+v", registration)
	}
	if registration.Status != model.ContestRegistrationStatusApproved {
		t.Fatalf("unexpected registration status: %s", registration.Status)
	}
}

func TestTeamServiceCreateTeamRequiresApprovedRegistration(t *testing.T) {
	t.Parallel()

	db := newContestTestDB(t)
	contestRepo := NewRepository(db)
	teamRepo := NewTeamRepository(db)
	service := NewTeamService(teamRepo, contestRepo)

	now := time.Now()
	if err := db.Create(&model.Contest{
		ID:        20,
		Title:     "team-ctf",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		Status:    model.ContestStatusRegistration,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestRegistration{
		ContestID: 20,
		UserID:    2002,
		Status:    model.ContestRegistrationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create pending registration: %v", err)
	}

	_, err := service.CreateTeam(context.Background(), 20, 2002, &dto.CreateTeamReq{Name: "Pending Team"})
	if err != errcode.ErrContestRegistrationPending {
		t.Fatalf("expected ErrContestRegistrationPending, got %v", err)
	}

	var count int64
	if err := db.Model(&model.Team{}).Count(&count).Error; err != nil {
		t.Fatalf("count teams: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no teams to be created, got %d", count)
	}
}

func TestParticipationServiceAnnouncementsAndMyProgress(t *testing.T) {
	t.Parallel()

	db := newContestTestDB(t)
	contestRepo := NewRepository(db)
	teamRepo := NewTeamRepository(db)
	service := NewParticipationService(db, contestRepo, teamRepo)

	now := time.Now()
	contest := &model.Contest{
		ID:        3,
		Title:     "autumn-ctf",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Status:    model.ContestStatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	teamID := int64(31)
	if err := db.Create(&model.ContestRegistration{
		ContestID: contest.ID,
		UserID:    3001,
		TeamID:    &teamID,
		Status:    model.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	if err := db.Create(&model.ContestChallenge{
		ID:          11,
		ContestID:   contest.ID,
		ChallengeID: 501,
		Points:      150,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}
	if err := db.Create(&model.Submission{
		UserID:      3001,
		ChallengeID: 501,
		ContestID:   &contest.ID,
		TeamID:      &teamID,
		IsCorrect:   true,
		Score:       150,
		SubmittedAt: now,
	}).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}

	created, err := service.CreateAnnouncement(context.Background(), contest.ID, 9001, &dto.CreateContestAnnouncementReq{
		Title:   "比赛开始",
		Content: "欢迎来到比赛。",
	})
	if err != nil {
		t.Fatalf("CreateAnnouncement() error = %v", err)
	}

	announcements, err := service.ListAnnouncements(context.Background(), contest.ID)
	if err != nil {
		t.Fatalf("ListAnnouncements() error = %v", err)
	}
	if len(announcements) != 1 || announcements[0].ID != created.ID {
		t.Fatalf("unexpected announcements: %+v", announcements)
	}

	progress, err := service.GetMyProgress(context.Background(), contest.ID, 3001)
	if err != nil {
		t.Fatalf("GetMyProgress() error = %v", err)
	}
	if progress.TeamID == nil || *progress.TeamID != teamID {
		t.Fatalf("unexpected team id: %+v", progress)
	}
	if len(progress.Solved) != 1 || progress.Solved[0].ContestChallengeID != 11 || progress.Solved[0].PointsEarned != 150 {
		t.Fatalf("unexpected solved progress: %+v", progress.Solved)
	}
}

func TestParticipationServiceListAndReviewRegistrations(t *testing.T) {
	t.Parallel()

	db := newContestTestDB(t)
	contestRepo := NewRepository(db)
	teamRepo := NewTeamRepository(db)
	service := NewParticipationService(db, contestRepo, teamRepo)

	now := time.Now()
	if err := db.Create(&model.Contest{
		ID:        30,
		Title:     "review-ctf",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		Status:    model.ContestStatusRegistration,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.User{
		ID:        3001,
		Username:  "alice",
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.ContestRegistration{
		ContestID: 30,
		UserID:    3001,
		Status:    model.ContestRegistrationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}

	status := model.ContestRegistrationStatusPending
	page, err := service.ListRegistrations(context.Background(), 30, &dto.ContestRegistrationQuery{
		Status: &status,
		Page:   1,
		Size:   10,
	})
	if err != nil {
		t.Fatalf("ListRegistrations() error = %v", err)
	}
	items, ok := page.List.([]*dto.ContestRegistrationResp)
	if !ok {
		t.Fatalf("unexpected list type: %T", page.List)
	}
	if len(items) != 1 || items[0].Username != "alice" || items[0].Status != model.ContestRegistrationStatusPending {
		t.Fatalf("unexpected registrations: %+v", items)
	}

	reviewed, err := service.ReviewRegistration(context.Background(), 30, items[0].ID, 9001, &dto.ReviewContestRegistrationReq{
		Status: model.ContestRegistrationStatusApproved,
	})
	if err != nil {
		t.Fatalf("ReviewRegistration() error = %v", err)
	}
	if reviewed.Status != model.ContestRegistrationStatusApproved || reviewed.ReviewedBy == nil || *reviewed.ReviewedBy != 9001 || reviewed.ReviewedAt == nil {
		t.Fatalf("unexpected reviewed registration: %+v", reviewed)
	}
}

func newContestTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Contest{},
		&model.Challenge{},
		&model.User{},
		&model.Team{},
		&model.TeamMember{},
		&model.ContestRegistration{},
		&model.ContestAnnouncement{},
		&model.ContestChallenge{},
		&model.Submission{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_user_challenge_correct
		ON submissions(contest_id, user_id, challenge_id)
		WHERE is_correct = 1 AND contest_id IS NOT NULL AND team_id IS NULL
	`).Error; err != nil {
		t.Fatalf("create contest submission user unique index: %v", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_team_challenge_correct
		ON submissions(contest_id, team_id, challenge_id)
		WHERE is_correct = 1 AND contest_id IS NOT NULL AND team_id IS NOT NULL
	`).Error; err != nil {
		t.Fatalf("create contest submission team unique index: %v", err)
	}
	return db
}
