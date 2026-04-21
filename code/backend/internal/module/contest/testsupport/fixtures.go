package testsupport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
		FlagPrefix: "awd",
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

	var contest model.Contest
	if err := db.WithContext(context.Background()).Where("id = ?", contestID).First(&contest).Error; err != nil {
		return
	}
	if contest.Mode != model.ContestModeAWD {
		return
	}

	var count int64
	if err := db.WithContext(context.Background()).
		Model(&model.ContestAWDService{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Count(&count).Error; err != nil {
		t.Fatalf("count awd contest services: %v", err)
	}
	if count > 0 {
		return
	}

	serviceSnapshot := buildAWDContestServiceFixtureSnapshot(t, db, challengeID)

	if err := db.Create(&model.ContestAWDService{
		ID:                DefaultAWDContestServiceID(contestID, challengeID),
		ContestID:         contestID,
		ChallengeID:       challengeID,
		DisplayName:       fmt.Sprintf("awd-service-%d", challengeID),
		ServiceSnapshot:   serviceSnapshot,
		Order:             0,
		IsVisible:         true,
		ScoreConfig:       `{"points":100}`,
		RuntimeConfig:     `{}`,
		ValidationState:   model.AWDCheckerValidationStatePending,
		LastPreviewResult: "",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create awd contest service fixture: %v", err)
	}
}

func SyncAWDContestServiceFixture(
	t *testing.T,
	db *gorm.DB,
	contestID, challengeID int64,
	displayName string,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	points, slaScore, defenseScore int,
	now time.Time,
) {
	t.Helper()

	if strings.TrimSpace(displayName) == "" {
		displayName = fmt.Sprintf("awd-service-%d", challengeID)
	}

	runtimeConfig := map[string]any{}
	if checkerType != "" {
		runtimeConfig["checker_type"] = string(checkerType)
	}
	trimmedCheckerConfig := strings.TrimSpace(checkerConfig)
	if trimmedCheckerConfig != "" {
		var checkerValue any
		if err := json.Unmarshal([]byte(trimmedCheckerConfig), &checkerValue); err == nil {
			runtimeConfig["checker_config"] = checkerValue
		} else {
			runtimeConfig["checker_config_raw"] = checkerConfig
		}
	}

	scoreConfig := map[string]any{
		"points":            points,
		"awd_sla_score":     slaScore,
		"awd_defense_score": defenseScore,
	}

	runtimeConfigRaw, err := json.Marshal(runtimeConfig)
	if err != nil {
		t.Fatalf("marshal awd runtime config: %v", err)
	}
	scoreConfigRaw, err := json.Marshal(scoreConfig)
	if err != nil {
		t.Fatalf("marshal awd score config: %v", err)
	}
	serviceSnapshot := buildAWDContestServiceFixtureSnapshot(t, db, challengeID)

	updates := map[string]any{
		"display_name":     displayName,
		"is_visible":       true,
		"score_config":     string(scoreConfigRaw),
		"runtime_config":   string(runtimeConfigRaw),
		"service_snapshot": serviceSnapshot,
		"updated_at":       now,
	}
	result := db.Model(&model.ContestAWDService{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Updates(updates)
	if result.Error != nil {
		t.Fatalf("update awd contest service fixture: %v", result.Error)
	}
	if result.RowsAffected > 0 {
		return
	}

	if err := db.Create(&model.ContestAWDService{
		ID:                DefaultAWDContestServiceID(contestID, challengeID),
		ContestID:         contestID,
		ChallengeID:       challengeID,
		DisplayName:       displayName,
		ServiceSnapshot:   serviceSnapshot,
		Order:             0,
		IsVisible:         true,
		ScoreConfig:       string(scoreConfigRaw),
		RuntimeConfig:     string(runtimeConfigRaw),
		ValidationState:   model.AWDCheckerValidationStatePending,
		LastPreviewResult: "",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create awd contest service fixture: %v", err)
	}
}

func SyncAWDContestServiceReadinessFixture(
	t *testing.T,
	db *gorm.DB,
	contestID, challengeID int64,
	state model.AWDCheckerValidationState,
	lastPreviewAt *time.Time,
	lastPreviewResult string,
) {
	t.Helper()

	updates := map[string]any{
		"awd_checker_validation_state":    state,
		"awd_checker_last_preview_at":     lastPreviewAt,
		"awd_checker_last_preview_result": lastPreviewResult,
	}
	result := db.Model(&model.ContestAWDService{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Updates(updates)
	if result.Error != nil {
		t.Fatalf("update awd contest service readiness fixture: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		t.Fatalf("missing awd contest service fixture for readiness sync: contest=%d challenge=%d", contestID, challengeID)
	}
}

func DefaultAWDContestServiceID(contestID, challengeID int64) int64 {
	return contestID*1_000_000_000 + challengeID
}

func buildAWDContestServiceFixtureSnapshot(t *testing.T, db *gorm.DB, challengeID int64) string {
	t.Helper()

	var challenge model.Challenge
	if err := db.WithContext(context.Background()).Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		t.Fatalf("load awd challenge fixture: %v", err)
	}
	snapshot := model.ContestAWDServiceSnapshot{
		Name:       challenge.Title,
		Category:   challenge.Category,
		Difficulty: challenge.Difficulty,
		FlagConfig: map[string]any{
			"flag_type":   challenge.FlagType,
			"flag_prefix": firstFixtureValue(challenge.FlagPrefix, "flag"),
		},
		RuntimeConfig: map[string]any{},
	}
	if challenge.ImageID > 0 {
		snapshot.RuntimeConfig["image_id"] = challenge.ImageID
	}
	raw, err := model.EncodeContestAWDServiceSnapshot(snapshot)
	if err != nil {
		t.Fatalf("encode awd service snapshot fixture: %v", err)
	}
	return raw
}

func firstFixtureValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
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

func AssertAWDServiceStatusCache(t *testing.T, redisClient *redis.Client, contestID, teamID, serviceID int64, expected string) {
	t.Helper()

	value, err := redisClient.HGet(context.Background(), rediskeys.AWDServiceStatusKey(contestID), rediskeys.AWDRoundFlagServiceField(teamID, serviceID)).Result()
	if err != nil {
		t.Fatalf("load awd service status cache for %d/%d: %v", teamID, serviceID, err)
	}
	if value != expected {
		t.Fatalf("unexpected awd service status cache for %d/%d: got %q want %q", teamID, serviceID, value, expected)
	}
}

func AssertAWDServiceStatusCacheMissing(t *testing.T, redisClient *redis.Client, contestID, teamID, serviceID int64) {
	t.Helper()

	_, err := redisClient.HGet(context.Background(), rediskeys.AWDServiceStatusKey(contestID), rediskeys.AWDRoundFlagServiceField(teamID, serviceID)).Result()
	if err == nil {
		t.Fatalf("expected missing awd service status cache for %d/%d", teamID, serviceID)
	}
	if !errors.Is(err, redis.Nil) {
		t.Fatalf("unexpected awd service status cache error for %d/%d: %v", teamID, serviceID, err)
	}
}
