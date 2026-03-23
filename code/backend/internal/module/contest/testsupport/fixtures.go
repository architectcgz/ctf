package testsupport

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/crypto"
)

const (
	AWDCheckSourceScheduler      = "scheduler"
	AWDCheckSourceManualCurrent  = "manual_current_round"
	AWDCheckSourceManualSelected = "manual_selected_round"
	AWDCheckSourceManualService  = "manual_service_check"
)

func CreateContestTeamRegistration(t *testing.T, db *gorm.DB, contestID, teamID, userID int64, teamName string, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       teamName,
		CaptainID:  userID,
		InviteCode: teamName,
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := db.Create(&model.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    &teamID,
		Status:    model.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
}

func CreateContestRegistrationForExistingTeam(t *testing.T, db *gorm.DB, contestID, teamID, userID int64, now time.Time) {
	t.Helper()

	if err := db.Create(&model.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    &teamID,
		Status:    model.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
}

func AssertContestTeamScore(t *testing.T, db *gorm.DB, teamID int64, expected int) {
	t.Helper()

	var team model.Team
	if err := db.First(&team, teamID).Error; err != nil {
		t.Fatalf("load team %d: %v", teamID, err)
	}
	if team.TotalScore != expected {
		t.Fatalf("unexpected team score for %d: %+v", teamID, team)
	}
	if team.LastSolveAt == nil {
		t.Fatalf("expected last solve time for team %d", teamID)
	}
}

func CreateAWDContestFixture(t *testing.T, db *gorm.DB, contestID int64, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "awd-contest",
		Mode:      model.ContestModeAWD,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Status:    model.ContestStatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create awd contest: %v", err)
	}
}

func CreateAWDRoundFixture(t *testing.T, db *gorm.DB, roundID, contestID int64, roundNumber, attackScore, defenseScore int, now time.Time) {
	t.Helper()

	if err := db.Create(&model.AWDRound{
		ID:           roundID,
		ContestID:    contestID,
		RoundNumber:  roundNumber,
		Status:       model.AWDRoundStatusRunning,
		AttackScore:  attackScore,
		DefenseScore: defenseScore,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error; err != nil {
		t.Fatalf("create awd round: %v", err)
	}
}

func CreateAWDRoundFixtureWithWindow(t *testing.T, db *gorm.DB, roundID, contestID int64, roundNumber, attackScore, defenseScore int, startedAt time.Time, endedAt time.Time) {
	t.Helper()

	round := &model.AWDRound{
		ID:           roundID,
		ContestID:    contestID,
		RoundNumber:  roundNumber,
		Status:       model.AWDRoundStatusRunning,
		StartedAt:    &startedAt,
		AttackScore:  attackScore,
		DefenseScore: defenseScore,
		CreatedAt:    startedAt,
		UpdatedAt:    startedAt,
	}
	if !endedAt.IsZero() {
		round.EndedAt = &endedAt
		round.Status = model.AWDRoundStatusFinished
	}
	if err := db.Create(round).Error; err != nil {
		t.Fatalf("create awd round with window: %v", err)
	}
}

func CreateAWDChallengeFixture(t *testing.T, db *gorm.DB, challengeID int64, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Challenge{
		ID:         challengeID,
		Title:      "awd-service",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}
}

func CreateAWDContestChallengeFixture(t *testing.T, db *gorm.DB, contestID, challengeID int64, now time.Time) {
	t.Helper()

	if err := db.Create(&model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: challengeID,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd contest challenge: %v", err)
	}
}

func CreateAWDTeamFixture(t *testing.T, db *gorm.DB, teamID, contestID int64, name string, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       name,
		CaptainID:  teamID + 1000,
		InviteCode: name,
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create awd team: %v", err)
	}
}

func CreateAWDTeamMemberFixture(t *testing.T, db *gorm.DB, contestID, teamID, userID int64, now time.Time) {
	t.Helper()

	if err := db.Create(&model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    userID,
		JoinedAt:  now,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create awd team member: %v", err)
	}
}

func BuildAWDRoundFlag(contestID int64, roundNumber int, teamID, challengeID int64, secret, prefix string) string {
	nonce := fmt.Sprintf("awd:%d:%d:%d:%d", contestID, roundNumber, teamID, challengeID)
	return crypto.GenerateDynamicFlag(teamID, challengeID, secret, nonce, prefix)
}

func AssertTeamTotalScore(t *testing.T, db *gorm.DB, teamID int64, expected int) {
	t.Helper()

	var team model.Team
	if err := db.Where("id = ?", teamID).First(&team).Error; err != nil {
		t.Fatalf("load team %d: %v", teamID, err)
	}
	if team.TotalScore != expected {
		t.Fatalf("unexpected team total_score for %d: got %d want %d", teamID, team.TotalScore, expected)
	}
}

func AssertContestRedisScore(t *testing.T, redisClient *redis.Client, contestID, teamID int64, expected float64) {
	t.Helper()

	score, err := redisClient.ZScore(context.Background(), rediskeys.RankContestTeamKey(contestID), strconv.FormatInt(teamID, 10)).Result()
	if err != nil {
		t.Fatalf("load redis score for team %d: %v", teamID, err)
	}
	if score != expected {
		t.Fatalf("unexpected redis score for team %d: got %v want %v", teamID, score, expected)
	}
}

func AssertContestRedisScoreMissing(t *testing.T, redisClient *redis.Client, contestID, teamID int64) {
	t.Helper()

	_, err := redisClient.ZScore(context.Background(), rediskeys.RankContestTeamKey(contestID), strconv.FormatInt(teamID, 10)).Result()
	if err == nil {
		t.Fatalf("expected missing redis score for team %d", teamID)
	}
	if !errors.Is(err, redis.Nil) {
		t.Fatalf("unexpected redis score error for team %d: %v", teamID, err)
	}
}

func AssertAWDServiceStatusCache(t *testing.T, redisClient *redis.Client, contestID, teamID, challengeID int64, expected string) {
	t.Helper()

	value, err := redisClient.HGet(context.Background(), rediskeys.AWDServiceStatusKey(contestID), rediskeys.AWDRoundFlagField(teamID, challengeID)).Result()
	if err != nil {
		t.Fatalf("load awd service status cache for %d/%d: %v", teamID, challengeID, err)
	}
	if value != expected {
		t.Fatalf("unexpected awd service status cache for %d/%d: got %q want %q", teamID, challengeID, value, expected)
	}
}

func AssertAWDServiceStatusCacheMissing(t *testing.T, redisClient *redis.Client, contestID, teamID, challengeID int64) {
	t.Helper()

	_, err := redisClient.HGet(context.Background(), rediskeys.AWDServiceStatusKey(contestID), rediskeys.AWDRoundFlagField(teamID, challengeID)).Result()
	if err == nil {
		t.Fatalf("expected missing awd service status cache for %d/%d", teamID, challengeID)
	}
	if !errors.Is(err, redis.Nil) {
		t.Fatalf("unexpected awd service status cache error for %d/%d: %v", teamID, challengeID, err)
	}
}
