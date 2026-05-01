package commands_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
)

func TestAWDServiceCreateRoundBlocksItemLevelReadinessReasons(t *testing.T) {
	testCases := []struct {
		name           string
		contestID      int64
		challengeID    int64
		blockingReason string
		checkerType    model.AWDCheckerType
		checkerConfig  string
		state          model.AWDCheckerValidationState
	}{
		{
			name:           "pending validation",
			contestID:      1210,
			challengeID:    12101,
			blockingReason: "pending_validation",
			checkerType:    model.AWDCheckerTypeLegacyProbe,
			checkerConfig:  `{}`,
			state:          model.AWDCheckerValidationStatePending,
		},
		{
			name:           "invalid checker config",
			contestID:      1211,
			challengeID:    12111,
			blockingReason: "invalid_checker_config",
			checkerType:    model.AWDCheckerTypeHTTPStandard,
			checkerConfig:  `{"get_flag":`,
			state:          model.AWDCheckerValidationStatePassed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := newAWDTestDB(t)
			service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
			now := time.Now()

			createAWDContestFixture(t, db, tc.contestID, now)
			seedCommandReadinessItem(t, db, tc.contestID, tc.challengeID, commandReadinessSeed{
				Now:             now,
				CheckerType:     tc.checkerType,
				CheckerConfig:   tc.checkerConfig,
				ValidationState: tc.state,
			})
			assertCommandReadinessBlockingReason(t, db, tc.contestID, tc.challengeID, tc.blockingReason)

			_, err := service.CreateRound(context.Background(), tc.contestID, contestcmd.CreateAWDRoundInput{
				RoundNumber: 1,
			})
			assertAWDReadinessBlocked(t, err)
		})
	}
}

func TestAWDServiceRunCurrentRoundChecksBlocksItemLevelReadinessReasons(t *testing.T) {
	testCases := []struct {
		name              string
		contestID         int64
		challengeID       int64
		blockingReason    string
		checkerType       model.AWDCheckerType
		checkerConfig     string
		state             model.AWDCheckerValidationState
		lastPreviewAt     *time.Time
		lastPreviewResult string
	}{
		{
			name:              "last preview failed",
			contestID:         2410,
			challengeID:       24101,
			blockingReason:    "last_preview_failed",
			checkerType:       model.AWDCheckerTypeHTTPStandard,
			checkerConfig:     `{"get_flag":{"path":"/health"}}`,
			state:             model.AWDCheckerValidationStateFailed,
			lastPreviewResult: `{"service_status":"down"}`,
		},
		{
			name:           "validation stale",
			contestID:      2411,
			challengeID:    24111,
			blockingReason: "validation_stale",
			checkerType:    model.AWDCheckerTypeLegacyProbe,
			checkerConfig:  `{}`,
			state:          model.AWDCheckerValidationStateStale,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := newAWDTestDB(t)
			now := time.Now()

			createAWDContestFixture(t, db, tc.contestID, now)
			createAWDRoundFixture(t, db, tc.contestID*10+1, tc.contestID, 1, 70, 35, now)
			seedCommandReadinessItem(t, db, tc.contestID, tc.challengeID, commandReadinessSeed{
				Now:               now,
				CheckerType:       tc.checkerType,
				CheckerConfig:     tc.checkerConfig,
				ValidationState:   tc.state,
				LastPreviewAt:     tc.lastPreviewAt,
				LastPreviewResult: tc.lastPreviewResult,
			})
			assertCommandReadinessBlockingReason(t, db, tc.contestID, tc.challengeID, tc.blockingReason)

			service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
			_, err := service.RunCurrentRoundChecks(context.Background(), tc.contestID, nil)
			assertAWDReadinessBlocked(t, err)
		})
	}
}

func TestContestServiceUpdateContestBlocksAWDStartForItemLevelReadinessReasons(t *testing.T) {
	testCases := []struct {
		name           string
		contestID      int64
		challengeID    int64
		blockingReason string
		checkerType    model.AWDCheckerType
		checkerConfig  string
		state          model.AWDCheckerValidationState
	}{
		{
			name:           "pending validation",
			contestID:      8210,
			challengeID:    82101,
			blockingReason: "pending_validation",
			checkerType:    model.AWDCheckerTypeLegacyProbe,
			checkerConfig:  `{}`,
			state:          model.AWDCheckerValidationStatePending,
		},
		{
			name:           "validation stale",
			contestID:      8211,
			challengeID:    82111,
			blockingReason: "validation_stale",
			checkerType:    model.AWDCheckerTypeLegacyProbe,
			checkerConfig:  `{}`,
			state:          model.AWDCheckerValidationStateStale,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, db := newContestCommandServiceForTest(t)
			now := time.Now()

			createContestForUpdateTest(t, db, &model.Contest{
				ID:        tc.contestID,
				Title:     "awd-start-item-block",
				Mode:      model.ContestModeAWD,
				Status:    model.ContestStatusRegistration,
				StartTime: now.Add(time.Hour),
				EndTime:   now.Add(2 * time.Hour),
				CreatedAt: now,
				UpdatedAt: now,
			})
			seedCommandReadinessItem(t, db, tc.contestID, tc.challengeID, commandReadinessSeed{
				Now:             now,
				CheckerType:     tc.checkerType,
				CheckerConfig:   tc.checkerConfig,
				ValidationState: tc.state,
			})
			assertCommandReadinessBlockingReason(t, db, tc.contestID, tc.challengeID, tc.blockingReason)

			_, err := service.UpdateContest(context.Background(), tc.contestID, contestcmd.UpdateContestInput{
				Status: strPtr(model.ContestStatusRunning),
			})
			assertContestReadinessBlocked(t, err)
		})
	}
}

type commandReadinessSeed struct {
	Now               time.Time
	CheckerType       model.AWDCheckerType
	CheckerConfig     string
	ValidationState   model.AWDCheckerValidationState
	LastPreviewAt     *time.Time
	LastPreviewResult string
}

func seedCommandReadinessItem(t *testing.T, db *gorm.DB, contestID, challengeID int64, seed commandReadinessSeed) {
	t.Helper()

	createAWDChallengeFixture(t, db, challengeID, seed.Now)
	createAWDContestChallengeFixture(t, db, contestID, challengeID, seed.Now)
	syncAWDContestServiceFixture(t, db, contestID, challengeID, "awd-service", seed.CheckerType, seed.CheckerConfig, 100, 0, 0, seed.Now)
	syncAWDContestServiceReadinessFixture(t, db, contestID, challengeID, seed.ValidationState, seed.LastPreviewAt, seed.LastPreviewResult)
}

func assertCommandReadinessBlockingReason(t *testing.T, db *gorm.DB, contestID, challengeID int64, want string) {
	t.Helper()

	service := contestqry.NewAWDService(contestinfra.NewAWDRepository(db), contestinfra.NewRepository(db))
	resp, err := service.GetReadiness(context.Background(), contestID)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if resp.Ready {
		t.Fatalf("expected readiness to remain blocked: %+v", resp)
	}
	if got := commandReadinessBlockingReasonByChallenge(resp.Items, challengeID); got != want {
		t.Fatalf("expected readiness item reason %q, got %q", want, got)
	}
}

func commandReadinessBlockingReasonByChallenge(items []contestqry.AWDReadinessItem, challengeID int64) string {
	for _, item := range items {
		if item.AWDChallengeID == challengeID {
			return item.BlockingReason
		}
	}
	return ""
}
