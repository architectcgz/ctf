package app

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"ctf-platform/internal/shared/taxonomy"
	fullrouterawdstate "ctf-platform/tests/system/http/fullrouterawdstate"
)

func TestFullRouter_InstanceHintAndProxyStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	fullrouterawdstate.VerifyInstanceHintAndProxyStateMatrix(t, fullrouterawdstate.InstanceHintAndProxyStateMatrixDriver{
		Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		StudentHeaders: bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
		PeerHeaders:    bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123")),
		TeacherHeaders: bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd)),
		ChallengeID:    env.challenge.ID,
		InstanceID:     env.instance.ID,
		CreateDraftChallenge: func(t *testing.T) int64 {
			return createDraftChallengeRecord(t, env, "Draft Hint Challenge").ID
		},
		ResetInstance: func(t *testing.T) {
			resetInstanceForAccessMatrix(t, env, env.instance.ID)
		},
		SetInstanceAccess: func(t *testing.T, accessURL string, status string) {
			updateInstanceAccessState(t, env, env.instance.ID, map[string]any{
				"access_url": accessURL,
				"status":     status,
			})
		},
	})
}

func TestFullRouter_AWDTrafficAdminStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	awdTeam := createContestTeam(t, env, env.awdContest.ID, env.student.ID, "AWD-Traffic-Blue", 4)
	trafficServiceID := contesttestsupport.DefaultAWDContestServiceID(env.awdContest.ID, env.challenge.ID)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		env.db,
		env.awdContest.ID,
		env.challenge.ID,
		"traffic-service",
		contestcontracts.AWDCheckerTypeHTTPStandard,
		`{"method":"GET","path":"/ping"}`,
		100,
		60,
		40,
		time.Now(),
	)

	fullrouterawdstate.VerifyAWDTrafficAdminStateMatrix(t, fullrouterawdstate.AWDTrafficAdminStateMatrixDriver{
		Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		AdminHeaders:   bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
		StudentHeaders: bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
		ContestID:      env.awdContest.ID,
		RoundID:        env.awdRound.ID,
		InstanceID:     env.instance.ID,
		SetInstanceAccess: func(t *testing.T, accessURL string) {
			updateInstanceAccessState(t, env, env.instance.ID, map[string]any{
				"user_id":      env.student.ID,
				"contest_id":   env.awdContest.ID,
				"team_id":      awdTeam.ID,
				"service_id":   trafficServiceID,
				"challenge_id": env.challenge.ID,
				"access_url":   accessURL,
				"status":       instancecontracts.InstanceStatusRunning,
			})
		},
	})
}

func TestFullRouter_VisibleAWDContestChallengesIncludeAWDServiceID(t *testing.T) {
	env := newFullRouterTestEnv(t)

	awdChallenge := &challengecontracts.AWDChallenge{
		Name:           "Visible AWD Challenge",
		Category:       taxonomy.DimensionWeb,
		Difficulty:     taxonomy.DifficultyMedium,
		ServiceType:    challengecontracts.AWDServiceTypeWebHTTP,
		DeploymentMode: challengecontracts.AWDDeploymentModeSingleContainer,
		Status:         challengecontracts.AWDChallengeStatusPublished,
	}
	if err := env.db.Create(awdChallenge).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	contest := createFullRouterContest(t, env, "Visible AWD Contest", contestcontracts.ContestStatusRunning)
	contest.Mode = contestcontracts.ContestModeAWD
	if err := env.db.Save(contest).Error; err != nil {
		t.Fatalf("update contest mode: %v", err)
	}

	now := time.Now()
	awdService := &contestcontracts.ContestAWDService{
		ContestID:       contest.ID,
		AWDChallengeID:  awdChallenge.ID,
		DisplayName:     "Visible Bank Portal",
		Order:           0,
		IsVisible:       true,
		ScoreConfig:     `{"points":260}`,
		RuntimeConfig:   fmt.Sprintf(`{"awd_challenge_id":%d}`, awdChallenge.ID),
		ServiceSnapshot: "{\n\t\t\t\"name\":\"Visible Bank Portal\",\n\t\t\t\"category\":\"web\",\n\t\t\t\"difficulty\":\"medium\"\n\t\t}",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := env.db.Create(awdService).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}

	createContestRegistration(t, env, contest.ID, env.student.ID, contestcontracts.ContestRegistrationStatusApproved, nil)

	fullrouterawdstate.VerifyVisibleAWDContestChallengesIncludeAWDServiceID(
		t,
		func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
		contest.ID,
		awdService.ID,
	)
}

func TestFullRouter_AWDContestLegacyChallengeInstanceRouteRejected(t *testing.T) {
	env := newFullRouterTestEnv(t)

	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	awdTeam := createContestTeam(t, env, env.awdContest.ID, env.student.ID, "AWD Legacy Route Team", 4)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		env.db,
		env.awdContest.ID,
		env.challenge.ID,
		"Matrix AWD Service",
		contestcontracts.AWDCheckerTypeHTTPStandard,
		`{"get_flag":{"path":"/health"}}`,
		100,
		18,
		28,
		time.Now(),
	)

	fullrouterawdstate.VerifyAWDContestLegacyChallengeInstanceRouteRejected(
		t,
		func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		studentHeaders,
		env.awdContest.ID,
		env.challenge.ID,
	)

	var awdInstanceCount int64
	if err := env.db.Model(&instancecontracts.Instance{}).
		Where("contest_id = ? AND team_id = ? AND user_id = ?", env.awdContest.ID, awdTeam.ID, env.student.ID).
		Count(&awdInstanceCount).Error; err != nil {
		t.Fatalf("count awd instances after challenge-based route request: %v", err)
	}
	if awdInstanceCount != 0 {
		t.Fatalf("expected no awd instance created through challenge-based route, got %d", awdInstanceCount)
	}
}

func TestFullRouter_AWDChallengeAuthoringStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	fullrouterawdstate.VerifyAWDChallengeAuthoringStateMatrix(
		t,
		func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
		bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd)),
		bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
	)
}

func updateInstanceAccessState(t *testing.T, env *fullRouterTestEnv, instanceID int64, attrs map[string]any) {
	t.Helper()

	if err := env.db.Model(&instancecontracts.Instance{}).Where("id = ?", instanceID).Updates(attrs).Error; err != nil {
		t.Fatalf("update instance access state: %v", err)
	}
}
