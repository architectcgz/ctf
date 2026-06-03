package app

import (
	"net/http/httptest"
	"testing"
	"time"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/internal/shared/taxonomy"
	fullrouterconteststate "ctf-platform/tests/system/http/fullrouterconteststate"
)

func TestFullRouter_ContestParticipationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	retryStudent := createFullRouterUser(t, env.db, "student_retry", "Password123", identitycontracts.RoleStudent, env.className)
	fillerStudent := createFullRouterUser(t, env.db, "student_filler", "Password123", identitycontracts.RoleStudent, env.className)
	registrationContest := createFullRouterContest(t, env, "Registration Matrix", contestcontracts.ContestStatusRegistration)

	fullrouterconteststate.VerifyContestParticipationStateMatrix(t, fullrouterconteststate.ContestParticipationStateMatrixDriver{
		Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		AdminHeaders:          bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
		StudentHeaders:        bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
		PeerHeaders:           bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123")),
		OtherHeaders:          bearerHeaders(loginForToken(t, env.router, env.otherStudent.Username, "Password123")),
		RetryHeaders:          bearerHeaders(loginForToken(t, env.router, retryStudent.Username, "Password123")),
		RegistrationContestID: registrationContest.ID,
		ExistingContestID:     env.contest.ID,
		ExistingTeamID:        env.team.ID,
		ExistingChallengeID:   env.challenge.ID,
		StudentID:             env.student.ID,
		PeerStudentID:         env.peerStudent.ID,
		OtherStudentID:        env.otherStudent.ID,
		RetryStudentID:        retryStudent.ID,
		FillerStudentID:       fillerStudent.ID,
		FindRegistration: func(t *testing.T, contestID, userID int64) fullrouterconteststate.RegistrationSnapshot {
			reg := findContestRegistration(t, env, contestID, userID)
			return fullrouterconteststate.RegistrationSnapshot{ID: reg.ID, Status: reg.Status}
		},
		CreateRegistration: func(t *testing.T, contestID, userID int64, status string) {
			createContestRegistration(t, env, contestID, userID, status, nil)
		},
		CreateTeam: func(t *testing.T, contestID, captainID int64, name string, maxMembers int) fullrouterconteststate.TeamSnapshot {
			team := createContestTeam(t, env, contestID, captainID, name, maxMembers)
			return fullrouterconteststate.TeamSnapshot{ID: team.ID, CaptainID: team.CaptainID}
		},
		AddTeamMember: func(t *testing.T, contestID, teamID, userID int64) {
			createContestTeamMember(t, env, contestID, teamID, userID)
		},
		CreateSubmission: func(t *testing.T, contestID, teamID, userID, challengeID int64, points int) {
			createContestSubmission(t, env, contestID, teamID, userID, challengeID, points)
		},
	})
}

func TestFullRouter_ContestAndReviewArchiveExportStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	createContestSubmission(t, env, env.contest.ID, env.team.ID, env.student.ID, env.challenge.ID, 100)
	secondChallenge := &appChallengeRow{
		Title:       "Export Matrix 2",
		Description: "contest export second solve",
		Category:    taxonomy.DimensionCrypto,
		Difficulty:  taxonomy.DifficultyEasy,
		Points:      150,
		Status:      challengecontracts.ChallengeStatusPublished,
		FlagType:    challengecontracts.FlagTypeStatic,
	}
	if err := env.db.Create(secondChallenge).Error; err != nil {
		t.Fatalf("create second challenge: %v", err)
	}
	secondContestChallenge := &contestcontracts.ContestChallenge{
		ContestID:   env.contest.ID,
		ChallengeID: secondChallenge.ID,
		Points:      150,
		Order:       2,
		IsVisible:   true,
	}
	if err := env.db.Create(secondContestChallenge).Error; err != nil {
		t.Fatalf("create second contest challenge: %v", err)
	}
	createContestSubmission(t, env, env.contest.ID, env.team.ID, env.student.ID, secondChallenge.ID, 150)

	fullrouterconteststate.VerifyContestAndReviewArchiveExportStateMatrix(t, fullrouterconteststate.ContestAndReviewArchiveExportStateMatrixDriver{
		Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		AdminHeaders:        bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
		TeacherHeaders:      bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd)),
		OtherTeacherHeaders: bearerHeaders(loginForToken(t, env.router, env.otherTeacher.Username, "Password123")),
		StudentHeaders:      bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
		ContestID:           env.contest.ID,
		StudentID:           env.student.ID,
		StudentUsername:     env.student.Username,
		WaitForReportStatus: func(t *testing.T, reportID int64, headers map[string]string) *assessmentcmd.ReportExportData {
			return waitForReportStatus(t, env, reportID, headers, assessmententity.ReportStatusReady, 5*time.Second)
		},
	})
}

func TestFullRouter_ContestChallengeAndScoreboardStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	challengeA := createRecommendationChallenge(t, env, "Contest Matrix A", taxonomy.DimensionWeb)
	challengeB := createRecommendationChallenge(t, env, "Contest Matrix B", taxonomy.DimensionWeb)
	editableContest := createFullRouterContest(t, env, "Editable Contest", contestcontracts.ContestStatusRegistration)
	conflictContest := createFullRouterContest(t, env, "Conflict Contest", contestcontracts.ContestStatusRegistration)
	scoreboardContest := createFullRouterContest(t, env, "Scoreboard Contest", contestcontracts.ContestStatusRunning)
	notFrozenContest := createFullRouterContest(t, env, "Not Frozen Contest", contestcontracts.ContestStatusRunning)

	createContestRegistration(t, env, scoreboardContest.ID, env.student.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, scoreboardContest.ID, env.peerStudent.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	teamAlpha := createContestTeam(t, env, scoreboardContest.ID, env.student.ID, "Alpha", 4)
	teamBeta := createContestTeam(t, env, scoreboardContest.ID, env.peerStudent.ID, "Beta", 4)

	fullrouterconteststate.VerifyContestChallengeAndScoreboardStateMatrix(t, fullrouterconteststate.ContestChallengeAndScoreboardStateMatrixDriver{
		Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		AdminHeaders:             bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
		StudentHeaders:           bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
		PeerHeaders:              bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123")),
		OtherHeaders:             bearerHeaders(loginForToken(t, env.router, env.otherStudent.Username, "Password123")),
		EditableContestID:        editableContest.ID,
		ConflictContestID:        conflictContest.ID,
		ScoreboardContestID:      scoreboardContest.ID,
		NotFrozenContestID:       notFrozenContest.ID,
		ChallengeAID:             challengeA.ID,
		ChallengeBID:             challengeB.ID,
		StudentID:                env.student.ID,
		PeerStudentID:            env.peerStudent.ID,
		TeamAlphaID:              teamAlpha.ID,
		TeamBetaID:               teamBeta.ID,
		ConflictSubmissionTeamID: env.team.ID,
		CreateRegistration: func(t *testing.T, contestID, userID int64, status string) {
			createContestRegistration(t, env, contestID, userID, status, nil)
		},
		SetContestStatus: func(t *testing.T, contestID int64, status string) {
			setContestStatus(t, env, contestID, status, nil)
		},
		CreateSubmission: func(t *testing.T, contestID, teamID, userID, challengeID int64, points int) {
			createContestSubmission(t, env, contestID, teamID, userID, challengeID, points)
		},
		DeleteSubmissions: func(t *testing.T, contestID, challengeID int64) {
			if err := env.db.Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).Delete(&contestcontracts.Submission{}).Error; err != nil {
				t.Fatalf("delete contest submissions: %v", err)
			}
		},
		SeedScore: func(t *testing.T, contestID, teamID int64, score int) {
			seedContestScore(t, env, contestID, teamID, float64(score))
		},
	})
}

func TestFullRouter_AdminContestListSupportsModeStatusesSortAndSummary(t *testing.T) {
	env := newFullRouterTestEnv(t)

	base := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)

	env.awdContest.Status = contestcontracts.ContestStatusDraft
	if err := env.db.Save(env.awdContest).Error; err != nil {
		t.Fatalf("update seeded awd contest: %v", err)
	}

	contestSpecs := []struct {
		title     string
		mode      string
		status    string
		startTime time.Time
	}{
		{title: "AWD Running", mode: contestcontracts.ContestModeAWD, status: contestcontracts.ContestStatusRunning, startTime: base.Add(4 * time.Hour)},
		{title: "AWD Registration", mode: contestcontracts.ContestModeAWD, status: contestcontracts.ContestStatusRegistration, startTime: base.Add(3 * time.Hour)},
		{title: "AWD Frozen", mode: contestcontracts.ContestModeAWD, status: contestcontracts.ContestStatusFrozen, startTime: base.Add(2 * time.Hour)},
		{title: "AWD Ended", mode: contestcontracts.ContestModeAWD, status: contestcontracts.ContestStatusEnded, startTime: base.Add(1 * time.Hour)},
		{title: "Jeopardy Running", mode: contestcontracts.ContestModeJeopardy, status: contestcontracts.ContestStatusRunning, startTime: base.Add(5 * time.Hour)},
		{title: "AWD Draft", mode: contestcontracts.ContestModeAWD, status: contestcontracts.ContestStatusDraft, startTime: base.Add(6 * time.Hour)},
	}

	for _, spec := range contestSpecs {
		contest := createFullRouterContest(t, env, spec.title, spec.status)
		contest.Mode = spec.mode
		contest.StartTime = spec.startTime
		contest.EndTime = spec.startTime.Add(2 * time.Hour)
		if err := env.db.Save(contest).Error; err != nil {
			t.Fatalf("update contest fixture %s: %v", spec.title, err)
		}
	}

	fullrouterconteststate.VerifyAdminContestListSupportsModeStatusesSortAndSummary(
		t,
		fullrouterconteststate.AdminContestListSupportsModeStatusesSortAndSummaryDriver{
			Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterRequest(t, env.router, method, target, payload, headers)
			},
			AdminHeaders: bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
		},
	)
}
