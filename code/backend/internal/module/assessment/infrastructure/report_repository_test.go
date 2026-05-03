package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	"ctf-platform/internal/teaching/evidence"
)

func newReportRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Challenge{},
		&model.AWDChallenge{},
		&model.Submission{},
		&model.AWDRound{},
		&model.AWDAttackLog{},
		&model.AWDTrafficEvent{},
		&model.Team{},
		&model.TeamMember{},
		&model.Instance{},
		&model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	return db
}

func findDimensionStat(rows []assessmentdomain.ReportDimensionStat, dimension string) *assessmentdomain.ReportDimensionStat {
	for index := range rows {
		if rows[index].Dimension == dimension {
			return &rows[index]
		}
	}
	return nil
}

func TestReportRepositoryGetPersonalStatsIncludesAWDSolvedAndAttempts(t *testing.T) {
	t.Parallel()

	db := newReportRepositoryTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC)

	users := []model.User{
		{ID: 1, Username: "alice", Role: model.RoleStudent, ClassName: "class-a", Status: model.UserStatusActive},
		{ID: 2, Username: "bob", Role: model.RoleStudent, ClassName: "class-a", Status: model.UserStatusActive},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	challenges := []model.Challenge{
		{ID: 101, Title: "web-entry", Category: "web", Difficulty: model.ChallengeDifficultyEasy, Points: 100, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 102, Title: "pwn-core", Category: "pwn", Difficulty: model.ChallengeDifficultyMedium, Points: 200, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(&challenges).Error; err != nil {
		t.Fatalf("seed challenges: %v", err)
	}

	submissions := []model.Submission{
		{ID: 1, UserID: 1, ChallengeID: 101, IsCorrect: true, SubmittedAt: now.Add(-5 * time.Minute), UpdatedAt: now.Add(-5 * time.Minute)},
		{ID: 2, UserID: 2, ChallengeID: 102, IsCorrect: true, SubmittedAt: now.Add(-4 * time.Minute), UpdatedAt: now.Add(-4 * time.Minute)},
	}
	if err := db.Create(&submissions).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	round := model.AWDRound{
		ID:          11,
		ContestID:   88,
		RoundNumber: 1,
		Status:      model.AWDRoundStatusFinished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&round).Error; err != nil {
		t.Fatalf("seed round: %v", err)
	}

	aliceID := int64(1)
	logs := []model.AWDAttackLog{
		{
			ID:                1,
			RoundID:           round.ID,
			AttackerTeamID:    301,
			VictimTeamID:      401,
			AWDChallengeID:    102,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &aliceID,
			IsSuccess:         true,
			ScoreGained:       200,
			CreatedAt:         now.Add(-3 * time.Minute),
		},
		{
			ID:                2,
			RoundID:           round.ID,
			AttackerTeamID:    301,
			VictimTeamID:      402,
			AWDChallengeID:    101,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &aliceID,
			IsSuccess:         true,
			ScoreGained:       0,
			CreatedAt:         now.Add(-2 * time.Minute),
		},
		{
			ID:                3,
			RoundID:           round.ID,
			AttackerTeamID:    301,
			VictimTeamID:      403,
			AWDChallengeID:    102,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &aliceID,
			IsSuccess:         false,
			ScoreGained:       0,
			CreatedAt:         now.Add(-1 * time.Minute),
		},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed awd attack logs: %v", err)
	}

	stats, err := repo.GetPersonalStats(ctx, 1)
	if err != nil {
		t.Fatalf("GetPersonalStats() error = %v", err)
	}

	if stats.TotalScore != 100 {
		t.Fatalf("expected ordinary total score 100 after awd separation, got %+v", stats)
	}
	if stats.TotalSolved != 1 {
		t.Fatalf("expected total solved 1 after awd separation, got %+v", stats)
	}
	if stats.TotalAttempts != 4 {
		t.Fatalf("expected total attempts 4, got %+v", stats)
	}
	if stats.Rank != 2 {
		t.Fatalf("expected rank 2 without awd score boost, got %+v", stats)
	}
}

func TestReportRepositoryListPersonalDimensionStatsDedupesPracticeAndAWD(t *testing.T) {
	t.Parallel()

	db := newReportRepositoryTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 13, 11, 0, 0, 0, time.UTC)

	user := model.User{ID: 7, Username: "neo", Role: model.RoleStudent, ClassName: "class-a", Status: model.UserStatusActive}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	challenges := []model.Challenge{
		{ID: 201, Title: "web-a", Category: "web", Difficulty: model.ChallengeDifficultyEasy, Points: 100, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 202, Title: "web-b", Category: "web", Difficulty: model.ChallengeDifficultyMedium, Points: 150, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 203, Title: "crypto-a", Category: "crypto", Difficulty: model.ChallengeDifficultyEasy, Points: 80, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(&challenges).Error; err != nil {
		t.Fatalf("seed challenges: %v", err)
	}

	submission := model.Submission{
		ID:          11,
		UserID:      user.ID,
		ChallengeID: 201,
		IsCorrect:   true,
		SubmittedAt: now.Add(-5 * time.Minute),
		UpdatedAt:   now.Add(-5 * time.Minute),
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}

	round := model.AWDRound{ID: 21, ContestID: 99, RoundNumber: 1, Status: model.AWDRoundStatusFinished, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&round).Error; err != nil {
		t.Fatalf("seed round: %v", err)
	}

	userID := user.ID
	logs := []model.AWDAttackLog{
		{
			ID:                21,
			RoundID:           round.ID,
			AttackerTeamID:    1,
			VictimTeamID:      2,
			AWDChallengeID:    201,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &userID,
			IsSuccess:         true,
			ScoreGained:       100,
			CreatedAt:         now.Add(-4 * time.Minute),
		},
		{
			ID:                22,
			RoundID:           round.ID,
			AttackerTeamID:    1,
			VictimTeamID:      3,
			AWDChallengeID:    202,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &userID,
			IsSuccess:         true,
			ScoreGained:       150,
			CreatedAt:         now.Add(-3 * time.Minute),
		},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed awd logs: %v", err)
	}

	rows, err := repo.ListPersonalDimensionStats(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListPersonalDimensionStats() error = %v", err)
	}

	web := findDimensionStat(rows, "web")
	if web == nil {
		t.Fatalf("expected web row, got %+v", rows)
	}
	if web.Solved != 1 || web.Total != 2 {
		t.Fatalf("expected web solved=1 total=2 after awd separation, got %+v", web)
	}

	crypto := findDimensionStat(rows, "crypto")
	if crypto == nil {
		t.Fatalf("expected crypto row, got %+v", rows)
	}
	if crypto.Solved != 0 || crypto.Total != 1 {
		t.Fatalf("expected crypto solved=0 total=1, got %+v", crypto)
	}
}

func TestReportRepositoryClassStatsIncludeAWDSolvedEvidence(t *testing.T) {
	t.Parallel()

	db := newReportRepositoryTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)

	users := []model.User{
		{ID: 1, Username: "alice", Role: model.RoleStudent, ClassName: "class-a", Status: model.UserStatusActive},
		{ID: 2, Username: "bob", Role: model.RoleStudent, ClassName: "class-a", Status: model.UserStatusActive},
		{ID: 3, Username: "charlie", Role: model.RoleStudent, ClassName: "class-b", Status: model.UserStatusActive},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	challenges := []model.Challenge{
		{ID: 301, Title: "web", Category: "web", Difficulty: model.ChallengeDifficultyEasy, Points: 100, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 302, Title: "pwn", Category: "pwn", Difficulty: model.ChallengeDifficultyMedium, Points: 200, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(&challenges).Error; err != nil {
		t.Fatalf("seed challenges: %v", err)
	}

	submissions := []model.Submission{
		{ID: 31, UserID: 1, ChallengeID: 301, IsCorrect: true, SubmittedAt: now.Add(-5 * time.Minute), UpdatedAt: now.Add(-5 * time.Minute)},
		{ID: 32, UserID: 2, ChallengeID: 301, IsCorrect: true, SubmittedAt: now.Add(-4 * time.Minute), UpdatedAt: now.Add(-4 * time.Minute)},
		{ID: 33, UserID: 3, ChallengeID: 302, IsCorrect: true, SubmittedAt: now.Add(-3 * time.Minute), UpdatedAt: now.Add(-3 * time.Minute)},
	}
	if err := db.Create(&submissions).Error; err != nil {
		t.Fatalf("seed submissions: %v", err)
	}

	round := model.AWDRound{ID: 41, ContestID: 100, RoundNumber: 1, Status: model.AWDRoundStatusFinished, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&round).Error; err != nil {
		t.Fatalf("seed round: %v", err)
	}

	aliceID := int64(1)
	log := model.AWDAttackLog{
		ID:                41,
		RoundID:           round.ID,
		AttackerTeamID:    11,
		VictimTeamID:      22,
		AWDChallengeID:    302,
		AttackType:        model.AWDAttackTypeFlagCapture,
		Source:            model.AWDAttackSourceSubmission,
		SubmittedByUserID: &aliceID,
		IsSuccess:         true,
		ScoreGained:       200,
		CreatedAt:         now.Add(-2 * time.Minute),
	}
	if err := db.Create(&log).Error; err != nil {
		t.Fatalf("seed awd log: %v", err)
	}

	avg, err := repo.GetClassAverageScore(ctx, "class-a")
	if err != nil {
		t.Fatalf("GetClassAverageScore() error = %v", err)
	}
	if avg != 100 {
		t.Fatalf("expected class average score 100 after awd separation, got %.2f", avg)
	}

	top, err := repo.ListClassTopStudents(ctx, "class-a", 10)
	if err != nil {
		t.Fatalf("ListClassTopStudents() error = %v", err)
	}
	if len(top) != 2 {
		t.Fatalf("expected 2 top students, got %+v", top)
	}
	if top[0].UserID != 1 || top[0].TotalScore != 100 || top[0].Rank != 1 {
		t.Fatalf("expected alice tied on ordinary score, got %+v", top[0])
	}
	if top[1].UserID != 2 || top[1].TotalScore != 100 || top[1].Rank != 2 {
		t.Fatalf("expected bob second, got %+v", top[1])
	}
}

func TestReportRepositoryGetStudentTimelineIncludesAWDAttackEvents(t *testing.T) {
	t.Parallel()

	db := newReportRepositoryTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 13, 13, 0, 0, 0, time.UTC)

	user := model.User{ID: 1, Username: "alice", Role: model.RoleStudent, ClassName: "class-a", Status: model.UserStatusActive}
	challenge := model.AWDChallenge{ID: 401, Name: "web-attack", Slug: "web-attack", Category: "web", Difficulty: model.ChallengeDifficultyEasy, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now}
	round := model.AWDRound{ID: 51, ContestID: 200, RoundNumber: 2, Status: model.AWDRoundStatusFinished, CreatedAt: now, UpdatedAt: now}
	teams := []model.Team{
		{ID: 501, ContestID: 200, Name: "red-team", CaptainID: user.ID, InviteCode: "invite-red", MaxMembers: 4, CreatedAt: now, UpdatedAt: now},
		{ID: 502, ContestID: 200, Name: "blue-team", CaptainID: user.ID, InviteCode: "invite-blue", MaxMembers: 4, CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	if err := db.Create(&round).Error; err != nil {
		t.Fatalf("seed round: %v", err)
	}
	if err := db.Create(&teams).Error; err != nil {
		t.Fatalf("seed teams: %v", err)
	}

	userID := user.ID
	logs := []model.AWDAttackLog{
		{
			ID:                51,
			RoundID:           round.ID,
			AttackerTeamID:    900,
			VictimTeamID:      501,
			AWDChallengeID:    challenge.ID,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &userID,
			IsSuccess:         false,
			ScoreGained:       0,
			CreatedAt:         now.Add(-2 * time.Minute),
		},
		{
			ID:                52,
			RoundID:           round.ID,
			AttackerTeamID:    900,
			VictimTeamID:      502,
			AWDChallengeID:    challenge.ID,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &userID,
			IsSuccess:         true,
			ScoreGained:       120,
			CreatedAt:         now.Add(-1 * time.Minute),
		},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed awd logs: %v", err)
	}

	events, err := repo.GetStudentTimeline(ctx, user.ID, 10, 0)
	if err != nil {
		t.Fatalf("GetStudentTimeline() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 timeline events, got %+v", events)
	}
	if events[0].Type != "awd_attack_submit" || events[0].Detail != "AWD 攻击命中 blue-team，得分 120" {
		t.Fatalf("expected latest AWD success event, got %+v", events[0])
	}
	if events[0].ChallengeID != 0 || events[0].AWDChallengeID != challenge.ID || events[0].Title != challenge.Name {
		t.Fatalf("expected AWD challenge identity to be separated, got %+v", events[0])
	}
	if events[0].IsCorrect == nil || !*events[0].IsCorrect {
		t.Fatalf("expected success event IsCorrect=true, got %+v", events[0])
	}
	if events[0].Points == nil || *events[0].Points != 120 {
		t.Fatalf("expected success points=120, got %+v", events[0])
	}
	if events[1].Type != "awd_attack_submit" || events[1].Detail != "AWD 攻击未命中 red-team" {
		t.Fatalf("expected previous AWD failure event, got %+v", events[1])
	}
	if events[1].IsCorrect == nil || *events[1].IsCorrect {
		t.Fatalf("expected failure event IsCorrect=false, got %+v", events[1])
	}
}

func TestReportRepositoryGetStudentEvidenceIncludesAWDAttackLogs(t *testing.T) {
	t.Parallel()

	db := newReportRepositoryTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 13, 14, 0, 0, 0, time.UTC)

	user := model.User{ID: 1, Username: "alice", Role: model.RoleStudent, ClassName: "class-a", Status: model.UserStatusActive}
	challenge := model.AWDChallenge{ID: 501, Name: "pwn-attack", Slug: "pwn-attack", Category: "pwn", Difficulty: model.ChallengeDifficultyMedium, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now}
	round := model.AWDRound{ID: 61, ContestID: 300, RoundNumber: 3, Status: model.AWDRoundStatusFinished, CreatedAt: now, UpdatedAt: now}
	teams := []model.Team{
		{ID: 601, ContestID: 300, Name: "green-team", CaptainID: user.ID, InviteCode: "invite-green", MaxMembers: 4, CreatedAt: now, UpdatedAt: now},
		{ID: 602, ContestID: 300, Name: "gold-team", CaptainID: user.ID, InviteCode: "invite-gold", MaxMembers: 4, CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	if err := db.Create(&round).Error; err != nil {
		t.Fatalf("seed round: %v", err)
	}
	if err := db.Create(&teams).Error; err != nil {
		t.Fatalf("seed teams: %v", err)
	}

	userID := user.ID
	logs := []model.AWDAttackLog{
		{
			ID:                61,
			RoundID:           round.ID,
			AttackerTeamID:    910,
			VictimTeamID:      601,
			AWDChallengeID:    challenge.ID,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &userID,
			IsSuccess:         false,
			ScoreGained:       0,
			CreatedAt:         now.Add(-2 * time.Minute),
		},
		{
			ID:                62,
			RoundID:           round.ID,
			AttackerTeamID:    910,
			VictimTeamID:      602,
			AWDChallengeID:    challenge.ID,
			AttackType:        model.AWDAttackTypeFlagCapture,
			Source:            model.AWDAttackSourceSubmission,
			SubmittedByUserID: &userID,
			IsSuccess:         true,
			ScoreGained:       150,
			CreatedAt:         now.Add(-1 * time.Minute),
		},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed awd logs: %v", err)
	}

	events, err := repo.GetStudentEvidence(ctx, user.ID, evidence.Query{})
	if err != nil {
		t.Fatalf("GetStudentEvidence() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 evidence events, got %+v", events)
	}
	if events[0].Type != "awd_attack_submission" || events[0].Detail != "AWD 攻击提交未命中" {
		t.Fatalf("expected first AWD failure evidence, got %+v", events[0])
	}
	if events[0].ChallengeID != challenge.ID || events[0].AWDChallengeID != challenge.ID || events[0].Title != challenge.Name {
		t.Fatalf("expected AWD evidence identity to align with realtime workspace, got %+v", events[0])
	}
	if events[0].Meta["event_stage"] != "exploit" {
		t.Fatalf("expected exploit stage meta, got %+v", events[0].Meta)
	}
	if success, ok := events[0].Meta["is_success"].(bool); !ok || success {
		t.Fatalf("expected failure meta, got %+v", events[0].Meta)
	}
	if victimName, ok := events[0].Meta["victim_team_name"].(string); !ok || victimName != "green-team" {
		t.Fatalf("expected victim team name green-team, got %+v", events[0].Meta)
	}
	if scope, ok := events[0].Meta["scope"].(string); !ok || scope != "student" {
		t.Fatalf("expected scope=student, got %+v", events[0].Meta)
	}

	if events[1].Type != "awd_attack_submission" || events[1].Detail != "AWD 攻击提交成功" {
		t.Fatalf("expected second AWD success evidence, got %+v", events[1])
	}
	if score, ok := events[1].Meta["score_gained"].(int); !ok || score != 150 {
		t.Fatalf("expected score_gained=150, got %+v", events[1].Meta)
	}
	if roundID, ok := events[1].Meta["round_id"].(int64); !ok || roundID != round.ID {
		t.Fatalf("expected round_id=%d, got %+v", round.ID, events[1].Meta)
	}
}
