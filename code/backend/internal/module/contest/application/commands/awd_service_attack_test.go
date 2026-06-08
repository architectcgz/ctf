package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/shared/taxonomy"
)

func TestAWDServiceSubmitAttackUsesCurrentRoundFlagAndRejectsStaleFlagAfterRotation(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 4, now)
	createAWDRoundFixture(t, db, 41, 4, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 401, now)
	createAWDContestChallengeFixture(t, db, 4, 401, now)
	createAWDTeamFixture(t, db, 411, 4, "Red", now)
	createAWDTeamFixture(t, db, 412, 4, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 4, 411, 4001, now)
	createContestRegistrationForExistingTeam(t, db, 4, 411, 4002, now)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 401).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}

	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(4), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}
	serviceID := defaultAWDContestServiceID(4, 401)
	flag := contestdomain.BuildAWDRoundFlag(4, 1, 412, 401, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(4, 41), map[string]any{
		rediskeys.AWDRoundFlagServiceField(412, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	first, err := service.SubmitAttack(context.Background(), 4001, 4, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() first error = %v", err)
	}
	if first.Source != contestentity.AWDAttackSourceSubmission {
		t.Fatalf("expected submission source, got %+v", first)
	}
	if !first.IsSuccess || first.ScoreGained != 80 || first.AttackerTeamID != 411 || first.VictimTeamID != 412 {
		t.Fatalf("unexpected first attack resp: %+v", first)
	}

	second, err := service.SubmitAttack(context.Background(), 4002, 4, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() second error = %v", err)
	}
	if second.Source != contestentity.AWDAttackSourceSubmission {
		t.Fatalf("expected submission source, got %+v", second)
	}
	if second.IsSuccess || second.ScoreGained != 0 {
		t.Fatalf("unexpected second attack resp: %+v", second)
	}

	var logs []contestentity.AWDAttackLog
	if err := db.Order("id ASC").Find(&logs).Error; err != nil {
		t.Fatalf("query attack logs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 attack logs, got %+v", logs)
	}
	if logs[0].SubmittedByUserID == nil || *logs[0].SubmittedByUserID != 4001 {
		t.Fatalf("expected first log submitted_by_user_id=4001, got %+v", logs[0])
	}
	if logs[1].SubmittedByUserID == nil || *logs[1].SubmittedByUserID != 4002 {
		t.Fatalf("expected second log submitted_by_user_id=4002, got %+v", logs[1])
	}
}

func TestAWDServiceSubmitAttackAcceptsServiceScopedRoundFlagField(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 24, now)
	createAWDRoundFixture(t, db, 241, 24, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 2401, now)
	createAWDContestChallengeFixture(t, db, 24, 2401, now)
	createAWDTeamFixture(t, db, 2411, 24, "Red", now)
	createAWDTeamFixture(t, db, 2412, 24, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 24, 2411, 24001, now)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 2401).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	serviceID := defaultAWDContestServiceID(24, 2401)
	if err := db.Model(&contestentity.ContestAWDService{}).
		Where("contest_id = ? AND awd_challenge_id = ?", 24, 2401).
		Updates(map[string]any{
			"display_name":   "Bank Portal",
			"order":          0,
			"is_visible":     true,
			"score_config":   `{"points":100,"awd_sla_score":1,"awd_defense_score":2}`,
			"runtime_config": `{"awd_challenge_id":2401,"checker_type":"legacy_probe","checker_config":{}}`,
			"updated_at":     now,
		}).Error; err != nil {
		t.Fatalf("update contest awd service: %v", err)
	}

	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(24), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}
	flag := contestdomain.BuildAWDRoundFlag(24, 1, 2412, 2401, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(24, 241), map[string]any{
		rediskeys.AWDRoundFlagServiceField(2412, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set service scoped round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 24001, 24, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 2412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() service scoped flag error = %v", err)
	}
	if resp.Source != contestentity.AWDAttackSourceSubmission || !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected service scoped submit resp: %+v", resp)
	}

	var logRecord contestentity.AWDAttackLog
	if err := db.Where("round_id = ? AND attacker_team_id = ? AND victim_team_id = ?", 241, 2411, 2412).First(&logRecord).Error; err != nil {
		t.Fatalf("load service scoped attack log: %v", err)
	}
	if logRecord.ServiceID != serviceID {
		t.Fatalf("expected attack log service_id=%d, got %+v", serviceID, logRecord)
	}

	var victimService contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 241, 2412, 2401).First(&victimService).Error; err != nil {
		t.Fatalf("load victim service after service scoped submit: %v", err)
	}
	if victimService.ServiceID != serviceID {
		t.Fatalf("expected victim service service_id=%d, got %+v", serviceID, victimService)
	}
}

func TestAWDServiceSubmitAttackPublishesAttackAcceptedEvent(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	bus := platformevents.NewBus()
	service.commands.SetEventBus(bus)

	now := time.Now()
	createAWDContestFixture(t, db, 14, now)
	createAWDRoundFixture(t, db, 141, 14, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 1401, now)
	createAWDContestChallengeFixture(t, db, 14, 1401, now)
	createAWDTeamFixture(t, db, 1411, 14, "Red", now)
	createAWDTeamFixture(t, db, 1412, 14, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 14, 1411, 14001, now)
	createContestRegistrationForExistingTeam(t, db, 14, 1411, 14002, now)
	serviceID := defaultAWDContestServiceID(14, 1401)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 1401).Updates(map[string]any{
		"flag_prefix": "awd",
		"category":    taxonomy.DimensionWeb,
	}).Error; err != nil {
		t.Fatalf("update challenge fields: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(14), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	flag := contestdomain.BuildAWDRoundFlag(14, 1, 1412, 1401, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(14, 141), map[string]any{
		rediskeys.AWDRoundFlagServiceField(1412, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	received := make(chan contestcontracts.AWDAttackAcceptedEvent, 2)
	bus.Subscribe(contestcontracts.EventAWDAttackAccepted, func(_ context.Context, evt platformevents.Event) error {
		payload, ok := evt.Payload.(contestcontracts.AWDAttackAcceptedEvent)
		if !ok {
			t.Fatalf("unexpected payload type: %T", evt.Payload)
		}
		received <- payload
		return nil
	})

	first, err := service.SubmitAttack(context.Background(), 14001, 14, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() first error = %v", err)
	}
	if !first.IsSuccess || first.ScoreGained != 80 {
		t.Fatalf("unexpected first attack resp: %+v", first)
	}

	second, err := service.SubmitAttack(context.Background(), 14002, 14, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() second error = %v", err)
	}
	if second.IsSuccess || second.ScoreGained != 0 {
		t.Fatalf("unexpected second attack resp: %+v", second)
	}

	select {
	case evt := <-received:
		if evt.UserID != 14001 || evt.AWDChallengeID != 1401 || evt.Dimension != taxonomy.DimensionWeb {
			t.Fatalf("unexpected event payload: %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected contest.awd.attack_accepted event to be published")
	}

	select {
	case evt := <-received:
		t.Fatalf("expected only one accepted event, got %+v", evt)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestAWDServiceSubmitAttackRotatesFlagImmediatelyAfterFirstSuccess(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})

	now := time.Now()
	createAWDContestFixture(t, db, 15, now)
	createAWDRoundFixture(t, db, 151, 15, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 1501, now)
	createAWDContestChallengeFixture(t, db, 15, 1501, now)
	createAWDTeamFixture(t, db, 1511, 15, "Red", now)
	createAWDTeamFixture(t, db, 1512, 15, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 15, 1511, 15001, now)
	createContestRegistrationForExistingTeam(t, db, 15, 1511, 15002, now)
	serviceID := defaultAWDContestServiceID(15, 1501)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 1501).Updates(map[string]any{
		"flag_prefix": "awd",
	}).Error; err != nil {
		t.Fatalf("update challenge fields: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(15), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	originalFlag := contestdomain.BuildAWDRoundFlag(15, 1, 1512, 1501, "awd-secret", "awd")
	flagField := rediskeys.AWDRoundFlagServiceField(1512, serviceID)
	roundKey := rediskeys.AWDRoundFlagsKey(15, 151)
	if err := redisClient.HSet(context.Background(), roundKey, map[string]any{
		flagField: originalFlag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	first, err := service.SubmitAttack(context.Background(), 15001, 15, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1512,
		Flag:         originalFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() first error = %v", err)
	}
	if !first.IsSuccess || first.ScoreGained != 80 {
		t.Fatalf("unexpected first attack resp: %+v", first)
	}

	rotatedFlag, err := redisClient.HGet(context.Background(), roundKey, flagField).Result()
	if err != nil {
		t.Fatalf("load rotated flag: %v", err)
	}
	if rotatedFlag == originalFlag {
		t.Fatalf("expected rotated flag to change, got same value %q", rotatedFlag)
	}

	second, err := service.SubmitAttack(context.Background(), 15002, 15, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1512,
		Flag:         originalFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() second error = %v", err)
	}
	if second.IsSuccess {
		t.Fatalf("expected stale flag to fail after rotation, got %+v", second)
	}
}

func TestAWDServiceSubmitAttackInjectsRotatedFlag(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	injector := &fakeAWDFlagInjector{}
	service.commands.SetFlagInjector(injector)

	now := time.Now()
	createAWDContestFixture(t, db, 16, now)
	createAWDRoundFixture(t, db, 161, 16, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 1601, now)
	createAWDContestChallengeFixture(t, db, 16, 1601, now)
	createAWDTeamFixture(t, db, 1611, 16, "Red", now)
	createAWDTeamFixture(t, db, 1612, 16, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 16, 1611, 16001, now)
	serviceID := defaultAWDContestServiceID(16, 1601)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 1601).Updates(map[string]any{
		"flag_prefix": "awd",
	}).Error; err != nil {
		t.Fatalf("update challenge fields: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(16), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	originalFlag := contestdomain.BuildAWDRoundFlag(16, 1, 1612, 1601, "awd-secret", "awd")
	flagField := rediskeys.AWDRoundFlagServiceField(1612, serviceID)
	roundKey := rediskeys.AWDRoundFlagsKey(16, 161)
	if err := redisClient.HSet(context.Background(), roundKey, map[string]any{
		flagField: originalFlag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 16001, 16, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1612,
		Flag:         originalFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() error = %v", err)
	}
	if !resp.IsSuccess {
		t.Fatalf("expected success, got %+v", resp)
	}
	if injector.callCount != 1 {
		t.Fatalf("expected injector to be called once, got %d", injector.callCount)
	}
	if len(injector.assignments) != 1 {
		t.Fatalf("expected one injected assignment, got %d", len(injector.assignments))
	}

	rotatedFlag, err := redisClient.HGet(context.Background(), roundKey, flagField).Result()
	if err != nil {
		t.Fatalf("load rotated flag: %v", err)
	}
	if injector.assignments[0].Flag != rotatedFlag {
		t.Fatalf("expected injected flag %q to match rotated redis flag %q", injector.assignments[0].Flag, rotatedFlag)
	}
}

func TestAWDServiceSubmitAttackPreservesOriginalFlagWhenRotationInjectionFails(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	service.commands.SetFlagInjector(&fakeAWDFlagInjector{err: errors.New("inject failed")})

	now := time.Now()
	createAWDContestFixture(t, db, 17, now)
	createAWDRoundFixture(t, db, 171, 17, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 1701, now)
	createAWDContestChallengeFixture(t, db, 17, 1701, now)
	createAWDTeamFixture(t, db, 1711, 17, "Red", now)
	createAWDTeamFixture(t, db, 1712, 17, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 17, 1711, 17001, now)
	serviceID := defaultAWDContestServiceID(17, 1701)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 1701).Updates(map[string]any{
		"flag_prefix": "awd",
	}).Error; err != nil {
		t.Fatalf("update challenge fields: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(17), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	originalFlag := contestdomain.BuildAWDRoundFlag(17, 1, 1712, 1701, "awd-secret", "awd")
	flagField := rediskeys.AWDRoundFlagServiceField(1712, serviceID)
	roundKey := rediskeys.AWDRoundFlagsKey(17, 171)
	if err := redisClient.HSet(context.Background(), roundKey, map[string]any{
		flagField: originalFlag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	if _, err := service.SubmitAttack(context.Background(), 17001, 17, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1712,
		Flag:         originalFlag,
	}); err == nil {
		t.Fatal("expected SubmitAttack() to fail when flag injection fails")
	}

	roundFlag, err := redisClient.HGet(context.Background(), roundKey, flagField).Result()
	if err != nil {
		t.Fatalf("load preserved flag: %v", err)
	}
	if roundFlag != originalFlag {
		t.Fatalf("expected original flag to be preserved, got %q want %q", roundFlag, originalFlag)
	}

	var count int64
	if err := db.Model(&contestentity.AWDAttackLog{}).Where("round_id = ?", 171).Count(&count).Error; err != nil {
		t.Fatalf("count attack logs: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no attack log to be created, got %d", count)
	}
}

func TestAWDServiceSubmitAttackTreatsAlreadyClaimedCurrentFlagAsFailure(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})

	now := time.Now()
	createAWDContestFixture(t, db, 18, now)
	createAWDRoundFixture(t, db, 181, 18, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 1801, now)
	createAWDContestChallengeFixture(t, db, 18, 1801, now)
	createAWDTeamFixture(t, db, 1811, 18, "Red", now)
	createAWDTeamFixture(t, db, 1812, 18, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 18, 1811, 18001, now)
	serviceID := defaultAWDContestServiceID(18, 1801)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 1801).Updates(map[string]any{
		"flag_prefix": "awd",
	}).Error; err != nil {
		t.Fatalf("update challenge fields: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(18), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	originalFlag := contestdomain.BuildAWDRoundFlag(18, 1, 1812, 1801, "awd-secret", "awd")
	competingFlag := "awd{competing-rotation}"
	flagField := rediskeys.AWDRoundFlagServiceField(1812, serviceID)
	roundKey := rediskeys.AWDRoundFlagsKey(18, 181)
	if err := redisClient.HSet(context.Background(), roundKey, map[string]any{
		flagField: originalFlag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	realStateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	service.commands = newAWDCommandServiceWithStateStoreForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{}, racingAWDRoundStateStore{
		AWDRoundStateStore: realStateStore,
		replaceIfMatchFn: func(ctx context.Context, contestID, roundID, teamID, targetServiceID int64, expectedFlag, nextFlag string, ttl time.Duration) (bool, error) {
			if err := realStateStore.SetAWDRoundFlag(ctx, contestID, roundID, teamID, targetServiceID, competingFlag, ttl); err != nil {
				return false, err
			}
			return false, nil
		},
	})

	resp, err := service.SubmitAttack(context.Background(), 18001, 18, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1812,
		Flag:         originalFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() error = %v", err)
	}
	if resp.IsSuccess || resp.ScoreGained != 0 {
		t.Fatalf("expected claimed stale flag to fail, got %+v", resp)
	}

	currentFlag, err := redisClient.HGet(context.Background(), roundKey, flagField).Result()
	if err != nil {
		t.Fatalf("load claimed flag: %v", err)
	}
	if currentFlag != competingFlag {
		t.Fatalf("expected competing flag %q to remain current, got %q", competingFlag, currentFlag)
	}

	var logs []contestentity.AWDAttackLog
	if err := db.Order("id ASC").Find(&logs).Error; err != nil {
		t.Fatalf("query attack logs: %v", err)
	}
	if len(logs) != 1 || logs[0].IsSuccess {
		t.Fatalf("expected one failed attack log, got %+v", logs)
	}
}

func TestAWDServiceSubmitAttackAcceptsPreviousRoundFlagWithinGrace(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		PreviousRoundGrace: time.Minute,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 5, now)
	createAWDRoundFixtureWithWindow(t, db, 51, 5, 1, 80, 20, now.Add(-5*time.Minute), now.Add(-10*time.Second))
	createAWDRoundFixtureWithWindow(t, db, 52, 5, 2, 80, 20, now.Add(-10*time.Second), time.Time{})
	createAWDChallengeFixture(t, db, 501, now)
	createAWDContestChallengeFixture(t, db, 5, 501, now)
	createAWDTeamFixture(t, db, 511, 5, "Red", now)
	createAWDTeamFixture(t, db, 512, 5, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 5, 511, 5001, now)
	serviceID := defaultAWDContestServiceID(5, 501)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 501).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(5), "2", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}
	currentFlag := contestdomain.BuildAWDRoundFlag(5, 2, 512, 501, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(5, 52), map[string]any{
		rediskeys.AWDRoundFlagServiceField(512, serviceID): currentFlag,
	}).Err(); err != nil {
		t.Fatalf("set current round flag: %v", err)
	}

	previousFlag := contestdomain.BuildAWDRoundFlag(5, 1, 512, 501, "awd-secret", "awd")
	resp, err := service.SubmitAttack(context.Background(), 5001, 5, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 512,
		Flag:         previousFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() previous round flag error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected previous round submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackAllowsFrozenContest(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 6, now)
	createAWDRoundFixture(t, db, 61, 6, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 601, now)
	createAWDContestChallengeFixture(t, db, 6, 601, now)
	createAWDTeamFixture(t, db, 611, 6, "Red", now)
	createAWDTeamFixture(t, db, 612, 6, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 6, 611, 6001, now)
	serviceID := defaultAWDContestServiceID(6, 601)

	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 6).Update("status", contestentity.ContestStatusFrozen).Error; err != nil {
		t.Fatalf("set contest frozen: %v", err)
	}
	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 601).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(6), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	flag := contestdomain.BuildAWDRoundFlag(6, 1, 612, 601, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(6, 61), map[string]any{
		rediskeys.AWDRoundFlagServiceField(612, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 6001, 6, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 612,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() frozen contest error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected frozen contest submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackIgnoresStaleCurrentRoundPointer(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 7, now)
	createAWDRoundFixtureWithWindow(t, db, 71, 7, 1, 80, 20, now.Add(-5*time.Minute), now.Add(-10*time.Second))
	createAWDRoundFixtureWithWindow(t, db, 72, 7, 2, 80, 20, now.Add(-10*time.Second), time.Time{})
	createAWDChallengeFixture(t, db, 701, now)
	createAWDContestChallengeFixture(t, db, 7, 701, now)
	createAWDTeamFixture(t, db, 711, 7, "Red", now)
	createAWDTeamFixture(t, db, 712, 7, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 7, 711, 7001, now)
	serviceID := defaultAWDContestServiceID(7, 701)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 701).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(7), "1", 0).Err(); err != nil {
		t.Fatalf("set stale current round: %v", err)
	}

	currentFlag := contestdomain.BuildAWDRoundFlag(7, 2, 712, 701, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(7, 72), map[string]any{
		rediskeys.AWDRoundFlagServiceField(712, serviceID): currentFlag,
	}).Err(); err != nil {
		t.Fatalf("set current round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 7001, 7, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 712,
		Flag:         currentFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with stale current round pointer error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 || resp.RoundID != 72 {
		t.Fatalf("unexpected stale pointer submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackUsesTimeDerivedCurrentRoundWhenRoundStatusLags(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		RoundInterval:      time.Minute,
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 8, now)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 8).Updates(map[string]any{
		"start_time": now.Add(-90 * time.Second),
		"end_time":   now.Add(90 * time.Second),
	}).Error; err != nil {
		t.Fatalf("update contest timing: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 81, 8, 1, 80, 20, now.Add(-90*time.Second), now.Add(-30*time.Second))
	createAWDRoundFixtureWithWindow(t, db, 82, 8, 2, 80, 20, now.Add(-30*time.Second), time.Time{})
	createAWDChallengeFixture(t, db, 801, now)
	createAWDContestChallengeFixture(t, db, 8, 801, now)
	createAWDTeamFixture(t, db, 811, 8, "Red", now)
	createAWDTeamFixture(t, db, 812, 8, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 8, 811, 8001, now)
	serviceID := defaultAWDContestServiceID(8, 801)

	if err := db.Model(&contestentity.AWDRound{}).Where("id = ?", 81).Updates(map[string]any{
		"status":   contestentity.AWDRoundStatusRunning,
		"ended_at": nil,
	}).Error; err != nil {
		t.Fatalf("mark stale round as running: %v", err)
	}
	if err := db.Model(&contestentity.AWDRound{}).Where("id = ?", 82).Updates(map[string]any{
		"status": contestentity.AWDRoundStatusPending,
	}).Error; err != nil {
		t.Fatalf("mark actual round as pending: %v", err)
	}
	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 801).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(8), "1", 0).Err(); err != nil {
		t.Fatalf("set stale current round: %v", err)
	}

	currentFlag := contestdomain.BuildAWDRoundFlag(8, 2, 812, 801, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(8, 82), map[string]any{
		rediskeys.AWDRoundFlagServiceField(812, serviceID): currentFlag,
	}).Err(); err != nil {
		t.Fatalf("set actual round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 8001, 8, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 812,
		Flag:         currentFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with lagging round status error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 || resp.RoundID != 82 {
		t.Fatalf("unexpected lagging status submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackRejectsPreviousFlagAfterMaterializingMissingCurrentRound(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		RoundInterval:      time.Minute,
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 9, now)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 9).Updates(map[string]any{
		"start_time": now.Add(-90 * time.Second),
		"end_time":   now.Add(90 * time.Second),
	}).Error; err != nil {
		t.Fatalf("update contest timing: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 91, 9, 1, 80, 20, now.Add(-90*time.Second), now.Add(-30*time.Second))
	createAWDChallengeFixture(t, db, 901, now)
	createAWDContestChallengeFixture(t, db, 9, 901, now)
	createAWDTeamFixture(t, db, 911, 9, "Red", now)
	createAWDTeamFixture(t, db, 912, 9, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 9, 911, 9001, now)
	serviceID := defaultAWDContestServiceID(9, 901)

	if err := db.Model(&contestentity.AWDRound{}).Where("id = ?", 91).Updates(map[string]any{
		"status":   contestentity.AWDRoundStatusRunning,
		"ended_at": nil,
	}).Error; err != nil {
		t.Fatalf("mark stale round as running: %v", err)
	}
	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 901).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(9), "1", 0).Err(); err != nil {
		t.Fatalf("set stale current round: %v", err)
	}

	previousFlag := contestdomain.BuildAWDRoundFlag(9, 1, 912, 901, "awd-secret", "awd")
	resp, err := service.SubmitAttack(context.Background(), 9001, 9, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 912,
		Flag:         previousFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with previous flag after materializing round error = %v", err)
	}
	if resp.IsSuccess || resp.ScoreGained != 0 {
		t.Fatalf("unexpected stale flag submit resp: %+v", resp)
	}

	var round contestentity.AWDRound
	if err := db.Where("contest_id = ? AND round_number = ?", 9, 2).First(&round).Error; err != nil {
		t.Fatalf("find materialized round: %v", err)
	}
	if resp.RoundID != round.ID {
		t.Fatalf("unexpected materialized round id for stale flag submit: resp=%d round=%d", resp.RoundID, round.ID)
	}
}

func TestAWDServiceSubmitAttackMaterializesMissingCurrentRound(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		RoundInterval:      time.Minute,
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 10, now)
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 10).Updates(map[string]any{
		"start_time": now.Add(-90 * time.Second),
		"end_time":   now.Add(90 * time.Second),
	}).Error; err != nil {
		t.Fatalf("update contest timing: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 101, 10, 1, 80, 20, now.Add(-90*time.Second), now.Add(-30*time.Second))
	createAWDChallengeFixture(t, db, 1001, now)
	createAWDContestChallengeFixture(t, db, 10, 1001, now)
	createAWDTeamFixture(t, db, 1011, 10, "Red", now)
	createAWDTeamFixture(t, db, 1012, 10, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 10, 1011, 10001, now)
	serviceID := defaultAWDContestServiceID(10, 1001)

	if err := db.Model(&contestCommandChallengeRow{}).Where("id = ?", 1001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}

	currentFlag := contestdomain.BuildAWDRoundFlag(10, 2, 1012, 1001, "awd-secret", "awd")
	resp, err := service.SubmitAttack(context.Background(), 10001, 10, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1012,
		Flag:         currentFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with missing current round error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected materialized round submit resp: %+v", resp)
	}

	var round contestentity.AWDRound
	if err := db.Where("contest_id = ? AND round_number = ?", 10, 2).First(&round).Error; err != nil {
		t.Fatalf("find materialized round: %v", err)
	}
	if resp.RoundID != round.ID {
		t.Fatalf("unexpected materialized round id: resp=%d round=%d", resp.RoundID, round.ID)
	}
	if round.AttackScore != 80 || round.DefenseScore != 20 {
		t.Fatalf("unexpected materialized round score: %+v", round)
	}

	currentRound, err := redisClient.Get(context.Background(), rediskeys.AWDCurrentRoundKey(10)).Result()
	if err != nil {
		t.Fatalf("load current round key: %v", err)
	}
	if currentRound != "2" {
		t.Fatalf("unexpected current round key: %s", currentRound)
	}

	flagValue, err := redisClient.HGet(
		context.Background(),
		rediskeys.AWDRoundFlagsKey(10, round.ID),
		rediskeys.AWDRoundFlagServiceField(1012, serviceID),
	).Result()
	if err != nil {
		t.Fatalf("load materialized round flag: %v", err)
	}
	if flagValue == "" {
		t.Fatal("expected materialized round flag to be stored after submit")
	}
	if flagValue == currentFlag {
		t.Fatalf("expected materialized round flag to rotate after first success, still got %q", flagValue)
	}
}
