package contest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeModule "ctf-platform/internal/module/challenge"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func TestSubmissionServiceSubmitFlagInContestAppliesDynamicScoreAndFirstBlood(t *testing.T) {
	service, redisClient, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(1)
	challengeID := int64(101)
	firstTeamID := int64(11)
	secondTeamID := int64(12)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	createContestTeamRegistration(t, db, contestID, firstTeamID, 1001, "Alpha", now)
	createContestTeamRegistration(t, db, contestID, secondTeamID, 1002, "Beta", now)

	firstResp, err := service.SubmitFlagInContest(context.Background(), 1001, contestID, challengeID, "flag{contest-dynamic}")
	if err != nil {
		t.Fatalf("first SubmitFlagInContest() error = %v", err)
	}
	if !firstResp.IsCorrect || firstResp.Points != 495 {
		t.Fatalf("unexpected first response: %+v", firstResp)
	}

	secondResp, err := service.SubmitFlagInContest(context.Background(), 1002, contestID, challengeID, "flag{contest-dynamic}")
	if err != nil {
		t.Fatalf("second SubmitFlagInContest() error = %v", err)
	}
	if !secondResp.IsCorrect || secondResp.Points != 405 {
		t.Fatalf("unexpected second response: %+v", secondResp)
	}

	var submissions []model.Submission
	if err := db.Order("submitted_at ASC, id ASC").Find(&submissions).Error; err != nil {
		t.Fatalf("list submissions: %v", err)
	}
	if len(submissions) != 2 || submissions[0].Score != 446 || submissions[1].Score != 405 {
		t.Fatalf("unexpected submissions: %+v", submissions)
	}

	var contestChallenge model.ContestChallenge
	if err := db.Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).First(&contestChallenge).Error; err != nil {
		t.Fatalf("load contest challenge: %v", err)
	}
	if contestChallenge.FirstBloodBy == nil || *contestChallenge.FirstBloodBy != firstTeamID {
		t.Fatalf("unexpected first blood team: %+v", contestChallenge.FirstBloodBy)
	}

	assertContestTeamScore(t, db, firstTeamID, 446)
	assertContestTeamScore(t, db, secondTeamID, 405)

	scoreboardKey := rediskeys.RankContestTeamKey(contestID)
	firstScore, err := redisClient.ZScore(context.Background(), scoreboardKey, teamIDToMember(firstTeamID)).Result()
	if err != nil {
		t.Fatalf("load first team score from redis: %v", err)
	}
	if firstScore != 446 {
		t.Fatalf("unexpected first redis score: %v", firstScore)
	}
	secondScore, err := redisClient.ZScore(context.Background(), scoreboardKey, teamIDToMember(secondTeamID)).Result()
	if err != nil {
		t.Fatalf("load second team score from redis: %v", err)
	}
	if secondScore != 405 {
		t.Fatalf("unexpected second redis score: %v", secondScore)
	}
}

func TestSubmissionServiceSubmitFlagInContestUsesContestScoreAsDynamicBase(t *testing.T) {
	service, _, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(2)
	challengeID := int64(102)
	teamID := int64(21)
	overrideScore := 300

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	if err := db.Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Update("contest_score", overrideScore).Error; err != nil {
		t.Fatalf("set contest score override: %v", err)
	}
	createContestTeamRegistration(t, db, contestID, teamID, 2001, "Gamma", now)

	resp, err := service.SubmitFlagInContest(context.Background(), 2001, contestID, challengeID, "flag{contest-dynamic}")
	if err != nil {
		t.Fatalf("SubmitFlagInContest() error = %v", err)
	}
	if !resp.IsCorrect || resp.Points != 297 {
		t.Fatalf("unexpected response: %+v", resp)
	}

	var submission model.Submission
	if err := db.First(&submission).Error; err != nil {
		t.Fatalf("load submission: %v", err)
	}
	if submission.Score != 297 {
		t.Fatalf("unexpected submission score: %+v", submission)
	}
}

func TestSubmissionServiceSubmitFlagInContestRejectsSecondSolveFromSameTeam(t *testing.T) {
	service, redisClient, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(4)
	challengeID := int64(104)
	teamID := int64(41)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	createContestTeamRegistration(t, db, contestID, teamID, 4001, "Delta", now)
	createContestRegistrationForExistingTeam(t, db, contestID, teamID, 4002, now)

	firstResp, err := service.SubmitFlagInContest(context.Background(), 4001, contestID, challengeID, "flag{contest-dynamic}")
	if err != nil {
		t.Fatalf("first SubmitFlagInContest() error = %v", err)
	}
	if !firstResp.IsCorrect || firstResp.Points != 495 {
		t.Fatalf("unexpected first response: %+v", firstResp)
	}

	secondResp, err := service.SubmitFlagInContest(context.Background(), 4002, contestID, challengeID, "flag{contest-dynamic}")
	if err == nil {
		t.Fatalf("expected same-team second solve conflict, got response: %+v", secondResp)
	}
	if !errors.Is(err, errcode.ErrContestChallengeSolved) {
		t.Fatalf("expected ErrContestChallengeSolved, got %v", err)
	}

	var submissions []model.Submission
	if err := db.Order("submitted_at ASC, id ASC").Find(&submissions).Error; err != nil {
		t.Fatalf("list submissions: %v", err)
	}
	if len(submissions) != 1 || submissions[0].UserID != 4001 || submissions[0].Score != 495 {
		t.Fatalf("unexpected submissions: %+v", submissions)
	}

	assertContestTeamScore(t, db, teamID, 495)

	scoreboardKey := rediskeys.RankContestTeamKey(contestID)
	score, err := redisClient.ZScore(context.Background(), scoreboardKey, teamIDToMember(teamID)).Result()
	if err != nil {
		t.Fatalf("load team score from redis: %v", err)
	}
	if score != 495 {
		t.Fatalf("unexpected redis score: %v", score)
	}
}

func TestScoreboardServiceRebuildScoreboardUsesTeamTotals(t *testing.T) {
	db := newContestTestDB(t)

	now := time.Now()
	contestID := int64(3)
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "rebuild-ctf",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Status:    model.ContestStatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create([]*model.Team{
		{ID: 31, ContestID: contestID, Name: "Alpha", CaptainID: 3001, InviteCode: "A", TotalScore: 600, CreatedAt: now, UpdatedAt: now},
		{ID: 32, ContestID: contestID, Name: "Beta", CaptainID: 3002, InviteCode: "B", TotalScore: 0, CreatedAt: now, UpdatedAt: now},
		{ID: 33, ContestID: contestID, Name: "Gamma", CaptainID: 3003, InviteCode: "C", TotalScore: 450, CreatedAt: now, UpdatedAt: now},
	}).Error; err != nil {
		t.Fatalf("create teams: %v", err)
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	cfg := &config.Config{
		Contest: config.ContestConfig{
			BaseScore: 1000,
			MinScore:  100,
			Decay:     0.9,
		},
	}
	service := NewScoreboardService(NewRepository(db), redisClient, &cfg.Contest, zap.NewNop())
	scoreboardKey := rediskeys.RankContestTeamKey(contestID)
	if err := redisClient.ZAdd(context.Background(), scoreboardKey, redis.Z{Score: 999, Member: teamIDToMember(99)}).Err(); err != nil {
		t.Fatalf("seed redis scoreboard: %v", err)
	}

	if err := service.RebuildScoreboard(context.Background(), contestID); err != nil {
		t.Fatalf("RebuildScoreboard() error = %v", err)
	}

	members, err := redisClient.ZRevRangeWithScores(context.Background(), scoreboardKey, 0, -1).Result()
	if err != nil {
		t.Fatalf("load redis scoreboard: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("unexpected redis members: %+v", members)
	}
	if memberToTeamID(members[0].Member) != 31 || members[0].Score != 600 {
		t.Fatalf("unexpected first redis member: %+v", members[0])
	}
	if memberToTeamID(members[1].Member) != 33 || members[1].Score != 450 {
		t.Fatalf("unexpected second redis member: %+v", members[1])
	}
}

func TestScoreboardServiceGetScoreboardUsesAWDAttackStats(t *testing.T) {
	db := newContestTestDB(t)
	if err := db.AutoMigrate(&model.AWDRound{}, &model.AWDAttackLog{}); err != nil {
		t.Fatalf("auto migrate awd models: %v", err)
	}

	now := time.Now()
	contestID := int64(13)
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "awd-scoreboard",
		Mode:      model.ContestModeAWD,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Status:    model.ContestStatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create([]*model.Team{
		{ID: 131, ContestID: contestID, Name: "Alpha", CaptainID: 3001, InviteCode: "A13", TotalScore: 260, CreatedAt: now, UpdatedAt: now},
		{ID: 132, ContestID: contestID, Name: "Beta", CaptainID: 3002, InviteCode: "B13", TotalScore: 120, CreatedAt: now, UpdatedAt: now},
	}).Error; err != nil {
		t.Fatalf("create teams: %v", err)
	}
	if err := db.Create(&model.AWDRound{
		ID:           1301,
		ContestID:    contestID,
		RoundNumber:  1,
		Status:       model.AWDRoundStatusRunning,
		AttackScore:  80,
		DefenseScore: 40,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error; err != nil {
		t.Fatalf("create awd round: %v", err)
	}
	if err := db.Create([]*model.AWDAttackLog{
		{
			ID:             13001,
			RoundID:        1301,
			AttackerTeamID: 131,
			VictimTeamID:   132,
			ChallengeID:    501,
			AttackType:     model.AWDAttackTypeFlagCapture,
			Source:         model.AWDAttackSourceSubmission,
			IsSuccess:      true,
			ScoreGained:    80,
			CreatedAt:      now.Add(-2 * time.Minute),
		},
		{
			ID:             13002,
			RoundID:        1301,
			AttackerTeamID: 131,
			VictimTeamID:   132,
			ChallengeID:    502,
			AttackType:     model.AWDAttackTypeServiceExploit,
			Source:         model.AWDAttackSourceSubmission,
			IsSuccess:      true,
			ScoreGained:    80,
			CreatedAt:      now.Add(-time.Minute),
		},
	}).Error; err != nil {
		t.Fatalf("create awd attacks: %v", err)
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})
	scoreboardKey := rediskeys.RankContestTeamKey(contestID)
	if err := redisClient.ZAdd(context.Background(), scoreboardKey,
		redis.Z{Score: 260, Member: teamIDToMember(131)},
		redis.Z{Score: 120, Member: teamIDToMember(132)},
	).Err(); err != nil {
		t.Fatalf("seed redis scoreboard: %v", err)
	}

	cfg := &config.Config{
		Contest: config.ContestConfig{
			BaseScore: 1000,
			MinScore:  100,
			Decay:     0.9,
		},
	}
	service := NewScoreboardService(NewRepository(db), redisClient, &cfg.Contest, zap.NewNop())

	resp, err := service.GetScoreboard(context.Background(), contestID, 1, 10)
	if err != nil {
		t.Fatalf("GetScoreboard() error = %v", err)
	}
	if resp.Scoreboard == nil || len(resp.Scoreboard.List) != 2 {
		t.Fatalf("unexpected scoreboard resp: %+v", resp)
	}
	if resp.Scoreboard.List[0].TeamID != 131 || resp.Scoreboard.List[0].SolvedCount != 2 || resp.Scoreboard.List[0].LastSubmissionAt == nil {
		t.Fatalf("unexpected first scoreboard item: %+v", resp.Scoreboard.List[0])
	}
	if resp.Scoreboard.List[1].TeamID != 132 || resp.Scoreboard.List[1].SolvedCount != 0 {
		t.Fatalf("unexpected second scoreboard item: %+v", resp.Scoreboard.List[1])
	}
}

func TestScoreboardServiceGetLiveScoreboardBypassesFrozenSnapshot(t *testing.T) {
	db := newContestTestDB(t)
	if err := db.AutoMigrate(&model.AWDRound{}, &model.AWDAttackLog{}); err != nil {
		t.Fatalf("auto migrate awd models: %v", err)
	}
	now := time.Now()
	contestID := int64(14)
	freezeTime := now.Add(-5 * time.Minute)

	if err := db.Create(&model.Contest{
		ID:         contestID,
		Title:      "awd-live-scoreboard",
		Mode:       model.ContestModeAWD,
		StartTime:  now.Add(-time.Hour),
		EndTime:    now.Add(time.Hour),
		FreezeTime: &freezeTime,
		Status:     model.ContestStatusFrozen,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create([]*model.Team{
		{ID: 141, ContestID: contestID, Name: "Alpha", CaptainID: 4001, InviteCode: "A14", TotalScore: 300, CreatedAt: now, UpdatedAt: now},
		{ID: 142, ContestID: contestID, Name: "Beta", CaptainID: 4002, InviteCode: "B14", TotalScore: 180, CreatedAt: now, UpdatedAt: now},
	}).Error; err != nil {
		t.Fatalf("create teams: %v", err)
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	liveKey := rediskeys.RankContestTeamKey(contestID)
	frozenKey := rediskeys.RankContestFrozenKey(contestID)
	if err := redisClient.ZAdd(context.Background(), liveKey,
		redis.Z{Score: 300, Member: teamIDToMember(141)},
		redis.Z{Score: 180, Member: teamIDToMember(142)},
	).Err(); err != nil {
		t.Fatalf("seed live scoreboard: %v", err)
	}
	if err := redisClient.ZAdd(context.Background(), frozenKey,
		redis.Z{Score: 120, Member: teamIDToMember(142)},
		redis.Z{Score: 100, Member: teamIDToMember(141)},
	).Err(); err != nil {
		t.Fatalf("seed frozen scoreboard: %v", err)
	}

	cfg := &config.Config{
		Contest: config.ContestConfig{
			BaseScore: 1000,
			MinScore:  100,
			Decay:     0.9,
		},
	}
	service := NewScoreboardService(NewRepository(db), redisClient, &cfg.Contest, zap.NewNop())

	publicResp, err := service.GetScoreboard(context.Background(), contestID, 1, 10)
	if err != nil {
		t.Fatalf("GetScoreboard() error = %v", err)
	}
	if !publicResp.Frozen || len(publicResp.Scoreboard.List) != 2 || publicResp.Scoreboard.List[0].TeamID != 142 {
		t.Fatalf("unexpected frozen scoreboard resp: %+v", publicResp)
	}

	liveResp, err := service.GetLiveScoreboard(context.Background(), contestID, 1, 10)
	if err != nil {
		t.Fatalf("GetLiveScoreboard() error = %v", err)
	}
	if liveResp.Frozen {
		t.Fatalf("expected live scoreboard to bypass frozen snapshot: %+v", liveResp)
	}
	if len(liveResp.Scoreboard.List) != 2 || liveResp.Scoreboard.List[0].TeamID != 141 {
		t.Fatalf("unexpected live scoreboard resp: %+v", liveResp)
	}
}

func newContestSubmissionTestService(t *testing.T) (*SubmissionService, *redis.Client, *gorm.DB) {
	t.Helper()

	db := newContestTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	t.Setenv("CTF_FLAG_SECRET", "0123456789abcdef0123456789abcdef")
	flagService, err := challengeModule.NewFlagService(db)
	if err != nil {
		t.Fatalf("new flag service: %v", err)
	}

	cfg := &config.Config{
		Contest: config.ContestConfig{
			BaseScore:       1000,
			MinScore:        100,
			Decay:           0.9,
			FirstBloodBonus: 0.1,
		},
	}
	contestRepo := NewRepository(db)
	scoreboardService := NewScoreboardService(contestRepo, redisClient, &cfg.Contest, zap.NewNop())
	service := NewSubmissionService(db, redisClient, flagService, NewTeamRepository(db), scoreboardService, cfg)
	return service, redisClient, db
}

func createContestSubmissionFixture(t *testing.T, db *gorm.DB, contestID, challengeID int64, now time.Time) {
	t.Helper()

	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "dynamic-ctf",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Status:    model.ContestStatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:         challengeID,
		Title:      "dynamic-web",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     500,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: challengeID,
		Points:      500,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}

	t.Setenv("CTF_FLAG_SECRET", "0123456789abcdef0123456789abcdef")
	flagService, err := challengeModule.NewFlagService(db)
	if err != nil {
		t.Fatalf("new flag service: %v", err)
	}
	if err := flagService.ConfigureStaticFlag(challengeID, "flag{contest-dynamic}", "flag"); err != nil {
		t.Fatalf("configure static flag: %v", err)
	}
}

func createContestTeamRegistration(t *testing.T, db *gorm.DB, contestID, teamID, userID int64, teamName string, now time.Time) {
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

func createContestRegistrationForExistingTeam(t *testing.T, db *gorm.DB, contestID, teamID, userID int64, now time.Time) {
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

func assertContestTeamScore(t *testing.T, db *gorm.DB, teamID int64, expected int) {
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
