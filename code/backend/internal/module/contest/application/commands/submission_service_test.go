package commands_test

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
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

type stubContestFlagValidator struct {
	validateFlagFn func(userID, challengeID int64, input string, nonce string) (bool, error)
}

func (s *stubContestFlagValidator) ValidateFlag(userID, challengeID int64, input string, nonce string) (bool, error) {
	if s.validateFlagFn == nil {
		return false, nil
	}
	return s.validateFlagFn(userID, challengeID, input, nonce)
}

func TestSubmissionServiceSubmitFlagInContestAppliesDynamicScoreAndFirstBlood(t *testing.T) {
	service, redisClient, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(1)
	challengeID := int64(101)
	firstTeamID := int64(11)
	secondTeamID := int64(12)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	testsupport.CreateContestTeamRegistration(t, db, contestID, firstTeamID, 1001, "Alpha", now)
	testsupport.CreateContestTeamRegistration(t, db, contestID, secondTeamID, 1002, "Beta", now)

	firstResp, err := service.SubmitFlagInContest(context.Background(), 1001, contestID, challengeID, "flag{contest-dynamic}")
	if err != nil {
		t.Fatalf("first SubmitFlagInContest() error = %v", err)
	}
	if !firstResp.IsCorrect || firstResp.Points != 495 {
		t.Fatalf("unexpected first response: %+v", firstResp)
	}
	if firstResp.Message != "" {
		t.Fatalf("expected contest correct response message to be omitted, got %+v", firstResp)
	}

	secondResp, err := service.SubmitFlagInContest(context.Background(), 1002, contestID, challengeID, "flag{contest-dynamic}")
	if err != nil {
		t.Fatalf("second SubmitFlagInContest() error = %v", err)
	}
	if !secondResp.IsCorrect || secondResp.Points != 405 {
		t.Fatalf("unexpected second response: %+v", secondResp)
	}
	if secondResp.Message != "" {
		t.Fatalf("expected contest correct response message to be omitted, got %+v", secondResp)
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

	testsupport.AssertContestTeamScore(t, db, firstTeamID, 446)
	testsupport.AssertContestTeamScore(t, db, secondTeamID, 405)

	scoreboardKey := rediskeys.RankContestTeamKey(contestID)
	firstScore, err := redisClient.ZScore(context.Background(), scoreboardKey, contestdomain.TeamIDToMember(firstTeamID)).Result()
	if err != nil {
		t.Fatalf("load first team score from redis: %v", err)
	}
	if firstScore != 446 {
		t.Fatalf("unexpected first redis score: %v", firstScore)
	}
	secondScore, err := redisClient.ZScore(context.Background(), scoreboardKey, contestdomain.TeamIDToMember(secondTeamID)).Result()
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
	testsupport.CreateContestTeamRegistration(t, db, contestID, teamID, 2001, "Gamma", now)

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

func TestSubmissionServiceSubmitFlagInContestOmitsIncorrectMessage(t *testing.T) {
	service, _, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(6)
	challengeID := int64(106)
	teamID := int64(61)
	userID := int64(6001)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	testsupport.CreateContestTeamRegistration(t, db, contestID, teamID, userID, "Wrong", now)

	resp, err := service.SubmitFlagInContest(context.Background(), userID, contestID, challengeID, "flag{wrong}")
	if err != nil {
		t.Fatalf("SubmitFlagInContest() error = %v", err)
	}
	if resp.IsCorrect {
		t.Fatalf("expected incorrect response, got %+v", resp)
	}
	if resp.Message != "" {
		t.Fatalf("expected contest incorrect response message to be omitted, got %+v", resp)
	}
}

func TestSubmissionServiceSubmitFlagInContestRejectsManualReviewChallenges(t *testing.T) {
	service, _, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(3)
	challengeID := int64(103)
	teamID := int64(31)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	testsupport.CreateContestTeamRegistration(t, db, contestID, teamID, 3001, "Manual", now)

	flagService, err := challengecmd.NewFlagService(challengeinfra.NewRepository(db), "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("new flag service: %v", err)
	}
	if err := flagService.ConfigureManualReviewFlag(challengeID); err != nil {
		t.Fatalf("configure manual review flag: %v", err)
	}

	_, err = service.SubmitFlagInContest(context.Background(), 3001, contestID, challengeID, "answer text")
	if err == nil {
		t.Fatal("expected manual review challenge submit in contest to fail")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestSubmissionServiceSubmitFlagInContestRejectsSecondSolveFromSameTeam(t *testing.T) {
	service, redisClient, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(4)
	challengeID := int64(104)
	teamID := int64(41)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	testsupport.CreateContestTeamRegistration(t, db, contestID, teamID, 4001, "Delta", now)
	testsupport.CreateContestRegistrationForExistingTeam(t, db, contestID, teamID, 4002, now)

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

	testsupport.AssertContestTeamScore(t, db, teamID, 495)

	scoreboardKey := rediskeys.RankContestTeamKey(contestID)
	score, err := redisClient.ZScore(context.Background(), scoreboardKey, contestdomain.TeamIDToMember(teamID)).Result()
	if err != nil {
		t.Fatalf("load team score from redis: %v", err)
	}
	if score != 495 {
		t.Fatalf("unexpected redis score: %v", score)
	}
}

func TestSubmissionServiceSubmitFlagInContestAcceptsSharedStaticFlagChallenge(t *testing.T) {
	service, _, db := newContestSubmissionTestService(t)

	now := time.Now()
	contestID := int64(5)
	challengeID := int64(105)
	teamID := int64(51)
	userID := int64(5001)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	testsupport.CreateContestTeamRegistration(t, db, contestID, teamID, userID, "Shared", now)

	flagService, err := challengecmd.NewFlagService(challengeinfra.NewRepository(db), "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("new flag service: %v", err)
	}
	if err := db.Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Update("instance_sharing", model.InstanceSharingShared).Error; err != nil {
		t.Fatalf("set shared instance scope: %v", err)
	}
	if err := flagService.ConfigureStaticFlag(challengeID, "flag{contest-shared-static}", "flag"); err != nil {
		t.Fatalf("configure shared static flag: %v", err)
	}

	resp, err := service.SubmitFlagInContest(context.Background(), userID, contestID, challengeID, "flag{contest-shared-static}")
	if err != nil {
		t.Fatalf("SubmitFlagInContest() error = %v", err)
	}
	if !resp.IsCorrect {
		t.Fatalf("expected shared static contest submission success, got %+v", resp)
	}
}

func TestScoreboardServiceRebuildScoreboardUsesTeamTotals(t *testing.T) {
	db := testsupport.SetupContestTestDB(t)

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
	service := contestcmd.NewScoreboardAdminService(contestinfra.NewRepository(db), redisClient, &cfg.Contest)
	scoreboardKey := rediskeys.RankContestTeamKey(contestID)
	if err := redisClient.ZAdd(context.Background(), scoreboardKey, redis.Z{Score: 999, Member: contestdomain.TeamIDToMember(99)}).Err(); err != nil {
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
	if contestdomain.MemberToTeamID(members[0].Member) != 31 || members[0].Score != 600 {
		t.Fatalf("unexpected first redis member: %+v", members[0])
	}
	if contestdomain.MemberToTeamID(members[1].Member) != 33 || members[1].Score != 450 {
		t.Fatalf("unexpected second redis member: %+v", members[1])
	}
}

func TestScoreboardServiceGetScoreboardUsesAWDAttackStats(t *testing.T) {
	db := testsupport.SetupContestTestDB(t)
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
		redis.Z{Score: 260, Member: contestdomain.TeamIDToMember(131)},
		redis.Z{Score: 120, Member: contestdomain.TeamIDToMember(132)},
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
	service := contestqry.NewScoreboardService(contestinfra.NewRepository(db), redisClient, &cfg.Contest, zap.NewNop())

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
	db := testsupport.SetupContestTestDB(t)
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
		redis.Z{Score: 300, Member: contestdomain.TeamIDToMember(141)},
		redis.Z{Score: 180, Member: contestdomain.TeamIDToMember(142)},
	).Err(); err != nil {
		t.Fatalf("seed live scoreboard: %v", err)
	}
	if err := redisClient.ZAdd(context.Background(), frozenKey,
		redis.Z{Score: 120, Member: contestdomain.TeamIDToMember(142)},
		redis.Z{Score: 100, Member: contestdomain.TeamIDToMember(141)},
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
	service := contestqry.NewScoreboardService(contestinfra.NewRepository(db), redisClient, &cfg.Contest, zap.NewNop())

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

func TestSubmissionServiceUsesChallengeFlagValidator(t *testing.T) {
	db := testsupport.SetupContestTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()
	contestID := int64(91)
	challengeID := int64(901)
	teamID := int64(9001)
	userID := int64(90001)

	createContestSubmissionFixture(t, db, contestID, challengeID, now)
	testsupport.CreateContestTeamRegistration(t, db, contestID, teamID, userID, "Validator", now)

	called := false
	service := contestcmd.NewSubmissionService(
		contestinfra.NewRepository(db),
		contestinfra.NewSubmissionRepository(db),
		redisClient,
		&stubContestFlagValidator{
			validateFlagFn: func(gotUserID, gotChallengeID int64, input string, nonce string) (bool, error) {
				called = true
				if gotUserID != userID || gotChallengeID != challengeID {
					t.Fatalf("unexpected validator args: user=%d challenge=%d", gotUserID, gotChallengeID)
				}
				if input != "flag{through-contract}" {
					t.Fatalf("unexpected validator input: %s", input)
				}
				if nonce != "" {
					t.Fatalf("unexpected validator nonce: %s", nonce)
				}
				return true, nil
			},
		},
		contestinfra.NewTeamRepository(db),
		nil,
		&config.Config{
			Contest: config.ContestConfig{
				BaseScore:       1000,
				MinScore:        100,
				Decay:           0.9,
				FirstBloodBonus: 0.1,
			},
		},
	)

	resp, err := service.SubmitFlagInContest(context.Background(), userID, contestID, challengeID, "flag{through-contract}")
	if err != nil {
		t.Fatalf("SubmitFlagInContest() error = %v", err)
	}
	if !called {
		t.Fatal("expected challenge flag validator to be used")
	}
	if !resp.IsCorrect {
		t.Fatalf("expected correct submission via contract validator, got %+v", resp)
	}
}

func newContestSubmissionTestService(t *testing.T) (*contestcmd.SubmissionService, *redis.Client, *gorm.DB) {
	t.Helper()

	db := testsupport.SetupContestTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	flagService, err := challengeqry.NewFlagService(challengeinfra.NewRepository(db), "0123456789abcdef0123456789abcdef")
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
	contestRepo := contestinfra.NewRepository(db)
	scoreboardService := contestcmd.NewScoreboardAdminService(contestRepo, redisClient, &cfg.Contest)
	service := contestcmd.NewSubmissionService(contestRepo, contestinfra.NewSubmissionRepository(db), redisClient, flagService, contestinfra.NewTeamRepository(db), scoreboardService, cfg)
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

	flagService, err := challengecmd.NewFlagService(challengeinfra.NewRepository(db), "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("new flag service: %v", err)
	}
	if err := flagService.ConfigureStaticFlag(challengeID, "flag{contest-dynamic}", "flag"); err != nil {
		t.Fatalf("configure static flag: %v", err)
	}
}
