package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"ctf-platform/internal/shared/taxonomy"
)

func TestFullRouter_ContestParticipationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))
	otherHeaders := bearerHeaders(loginForToken(t, env.router, env.otherStudent.Username, "Password123"))

	registrationContest := createFullRouterContest(t, env, "Registration Matrix", contestcontracts.ContestStatusRegistration)
	retryStudent := createFullRouterUser(t, env.db, "student_retry", "Password123", identitycontracts.RoleStudent, env.className)
	retryHeaders := bearerHeaders(loginForToken(t, env.router, retryStudent.Username, "Password123"))
	fillerStudent := createFullRouterUser(t, env.db, "student_filler", "Password123", identitycontracts.RoleStudent, env.className)

	resp := performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", registrationContest.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	peerRegistration := findContestRegistration(t, env, registrationContest.ID, env.peerStudent.ID)
	if peerRegistration.Status != contestcontracts.ContestRegistrationStatusPending {
		t.Fatalf("expected pending registration, got %s", peerRegistration.Status)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", registrationContest.ID), map[string]any{
		"name":        "PendingTeam",
		"max_members": 3,
	}, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/registrations?status=pending", registrationContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	var registrationPage map[string]any
	decodeFullRouterData(t, resp, &registrationPage)
	if total := int(registrationPage["total"].(float64)); total != 1 {
		t.Fatalf("expected 1 pending registration, got %d", total)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/contests/%d/registrations/%d", registrationContest.ID, peerRegistration.ID), map[string]any{
		"status": contestcontracts.ContestRegistrationStatusApproved,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", env.contest.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	createContestRegistration(t, env, registrationContest.ID, env.student.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, registrationContest.ID, env.otherStudent.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, registrationContest.ID, fillerStudent.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, registrationContest.ID, retryStudent.ID, contestcontracts.ContestRegistrationStatusRejected, nil)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", registrationContest.ID), nil, retryHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	retryRegistration := findContestRegistration(t, env, registrationContest.ID, retryStudent.ID)
	if retryRegistration.Status != contestcontracts.ContestRegistrationStatusPending {
		t.Fatalf("expected rejected registration to requeue as pending, got %s", retryRegistration.Status)
	}

	fullTeam := createContestTeam(t, env, registrationContest.ID, env.otherStudent.ID, "FullTeam", 2)
	createContestTeamMember(t, env, registrationContest.ID, fullTeam.ID, fillerStudent.ID)
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams/%d/join", registrationContest.ID, fullTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", registrationContest.ID), map[string]any{
		"name":        "AlphaTeam",
		"max_members": 4,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdTeam contesthttp.TeamResp
	decodeFullRouterData(t, resp, &createdTeam)
	if createdTeam.CaptainID != env.student.ID {
		t.Fatalf("expected student captain id %d, got %d", env.student.ID, createdTeam.CaptainID)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", registrationContest.ID), map[string]any{
		"name":        "DuplicateTeam",
		"max_members": 4,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams/%d/join", registrationContest.ID, createdTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d", registrationContest.ID, createdTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d/leave", registrationContest.ID, createdTeam.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d/leave", registrationContest.ID, createdTeam.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d", registrationContest.ID, createdTeam.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	createContestSubmission(t, env, env.contest.ID, env.team.ID, env.student.ID, env.challenge.ID, 100)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/my-progress", env.contest.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var progress contesthttp.ContestMyProgressResp
	decodeFullRouterData(t, resp, &progress)
	if progress.ContestID != env.contest.ID || len(progress.Solved) == 0 {
		t.Fatalf("expected existing contest progress, got %+v", progress)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/my-progress", env.contest.ID), nil, otherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var emptyProgress contesthttp.ContestMyProgressResp
	decodeFullRouterData(t, resp, &emptyProgress)
	if emptyProgress.TeamID != nil || len(emptyProgress.Solved) != 0 {
		t.Fatalf("expected empty progress for unregistered student, got %+v", emptyProgress)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/announcements", env.contest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var announcements []contesthttp.ContestAnnouncementResp
	decodeFullRouterData(t, resp, &announcements)
	if len(announcements) == 0 {
		t.Fatalf("expected seeded announcement")
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/announcements", env.contest.ID), map[string]any{
		"title":   "新的公告",
		"content": "integration notice",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdAnnouncement contesthttp.ContestAnnouncementResp
	decodeFullRouterData(t, resp, &createdAnnouncement)
	if createdAnnouncement.Title != "新的公告" {
		t.Fatalf("unexpected announcement title: %+v", createdAnnouncement)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/announcements/%d", env.contest.ID, createdAnnouncement.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
}

func TestFullRouter_ContestAndReviewArchiveExportStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	otherTeacherHeaders := bearerHeaders(loginForToken(t, env.router, env.otherTeacher.Username, "Password123"))

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

	resp := performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/export", env.contest.ID), map[string]any{
		"format": "json",
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	missingContestResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/admin/contests/999999/export", map[string]any{
		"format": "json",
	}, adminHeaders)
	assertFullRouterStatus(t, missingContestResp, http.StatusNotFound)

	var contestExport assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &contestExport)
	if contestExport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected contest export to start in processing state, got %+v", contestExport)
	}

	contestReady := waitForReportStatus(t, env, contestExport.ReportID, adminHeaders, assessmententity.ReportStatusReady, 5*time.Second)
	if contestReady.DownloadURL == nil {
		t.Fatalf("expected contest export download url after ready, got %+v", contestReady)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", contestExport.ReportID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type for contest export, got %q", contentType)
	}
	contestBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read contest export body: %v", err)
	}
	if !bytes.Contains(contestBody, []byte(`"contest"`)) || !bytes.Contains(contestBody, []byte(env.contest.Title)) {
		t.Fatalf("expected contest export payload to contain contest metadata, got %s", string(contestBody))
	}
	if !bytes.Contains(contestBody, []byte(`"solved_count": 2`)) {
		t.Fatalf("expected contest export scoreboard solved_count to include multiple solved challenges, got %s", string(contestBody))
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", env.student.ID), map[string]any{
		"format": "json",
	}, otherTeacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", env.student.ID), map[string]any{
		"format": "json",
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var reviewArchive assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &reviewArchive)
	if reviewArchive.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected review archive export to start in processing state, got %+v", reviewArchive)
	}

	reviewReady := waitForReportStatus(t, env, reviewArchive.ReportID, teacherHeaders, assessmententity.ReportStatusReady, 5*time.Second)
	if reviewReady.DownloadURL == nil {
		t.Fatalf("expected review archive download url after ready, got %+v", reviewReady)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", reviewArchive.ReportID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type for review archive, got %q", contentType)
	}
	reviewBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read review archive body: %v", err)
	}
	if !bytes.Contains(reviewBody, []byte(`"student"`)) || !bytes.Contains(reviewBody, []byte(env.student.Username)) {
		t.Fatalf("expected review archive payload to contain student metadata, got %s", string(reviewBody))
	}
	if !bytes.Contains(reviewBody, []byte(`"teacher_observations"`)) {
		t.Fatalf("expected review archive payload to contain teacher observations, got %s", string(reviewBody))
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", env.student.ID), nil, otherTeacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if body, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("read review archive view body: %v", err)
	} else if !bytes.Contains(body, []byte(`"teacher_observations"`)) || !bytes.Contains(body, []byte(env.student.Username)) {
		t.Fatalf("expected review archive view payload to contain observations and student username, got %s", string(body))
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", env.student.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", env.student.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)
}

func TestFullRouter_ContestChallengeAndScoreboardStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))
	peerHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))
	otherHeaders := bearerHeaders(loginForToken(t, env.router, env.otherStudent.Username, "Password123"))

	challengeA := createRecommendationChallenge(t, env, "Contest Matrix A", taxonomy.DimensionWeb)
	challengeB := createRecommendationChallenge(t, env, "Contest Matrix B", taxonomy.DimensionWeb)
	editableContest := createFullRouterContest(t, env, "Editable Contest", contestcontracts.ContestStatusRegistration)

	resp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", editableContest.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	hidden := false
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), map[string]any{
		"challenge_id": challengeA.ID,
		"points":       220,
		"order":        1,
		"is_visible":   hidden,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var contestChallenge contesthttp.ContestChallengeResp
	decodeFullRouterData(t, resp, &contestChallenge)
	if contestChallenge.Points != 220 || contestChallenge.IsVisible {
		t.Fatalf("unexpected contest challenge: %+v", contestChallenge)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), map[string]any{
		"challenge_id": challengeA.ID,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	updatedVisible := true
	updatedPoints := 260
	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", editableContest.ID, challengeA.ID), map[string]any{
		"points":     updatedPoints,
		"order":      2,
		"is_visible": updatedVisible,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var adminChallenges []contesthttp.ContestChallengeResp
	decodeFullRouterData(t, resp, &adminChallenges)
	if len(adminChallenges) != 1 || adminChallenges[0].Points != updatedPoints || adminChallenges[0].Order != 2 || !adminChallenges[0].IsVisible {
		t.Fatalf("unexpected admin contest challenges: %+v", adminChallenges)
	}

	createContestRegistration(t, env, editableContest.ID, env.student.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, editableContest.ID, env.peerStudent.ID, contestcontracts.ContestRegistrationStatusPending, nil)

	setContestStatus(t, env, editableContest.ID, contestcontracts.ContestStatusRunning, nil)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", editableContest.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var visibleChallenges []contesthttp.ContestChallengeInfo
	decodeFullRouterData(t, resp, &visibleChallenges)
	if len(visibleChallenges) != 1 || visibleChallenges[0].ChallengeID != challengeA.ID || visibleChallenges[0].Points != updatedPoints {
		t.Fatalf("unexpected visible contest challenges: %+v", visibleChallenges)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", editableContest.ID, challengeA.ID), nil, otherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", editableContest.ID, challengeA.ID), nil, peerHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", editableContest.ID, challengeA.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var startedContestInstance instancecontracts.InstanceResp
	decodeFullRouterData(t, resp, &startedContestInstance)
	if startedContestInstance.ChallengeID != challengeA.ID {
		t.Fatalf("unexpected started contest instance: %+v", startedContestInstance)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", editableContest.ID), map[string]any{
		"challenge_id": challengeB.ID,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	conflictContest := createFullRouterContest(t, env, "Conflict Contest", contestcontracts.ContestStatusRegistration)
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", conflictContest.ID), map[string]any{
		"challenge_id": challengeB.ID,
		"points":       180,
		"order":        1,
		"is_visible":   true,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	createContestSubmission(t, env, conflictContest.ID, env.team.ID, env.student.ID, challengeB.ID, 180)
	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", conflictContest.ID, challengeB.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusConflict)

	if err := env.db.Where("contest_id = ? AND challenge_id = ?", conflictContest.ID, challengeB.ID).Delete(&contestcontracts.Submission{}).Error; err != nil {
		t.Fatalf("delete conflict contest submissions: %v", err)
	}
	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", conflictContest.ID, challengeB.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	scoreboardContest := createFullRouterContest(t, env, "Scoreboard Contest", contestcontracts.ContestStatusRunning)
	createContestRegistration(t, env, scoreboardContest.ID, env.student.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, scoreboardContest.ID, env.peerStudent.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	teamAlpha := createContestTeam(t, env, scoreboardContest.ID, env.student.ID, "Alpha", 4)
	teamBeta := createContestTeam(t, env, scoreboardContest.ID, env.peerStudent.ID, "Beta", 4)
	seedContestScore(t, env, scoreboardContest.ID, teamAlpha.ID, 100)
	seedContestScore(t, env, scoreboardContest.ID, teamBeta.ID, 80)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", scoreboardContest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publicScoreboard contesthttp.ScoreboardResp
	decodeFullRouterData(t, resp, &publicScoreboard)
	if publicScoreboard.Frozen || len(publicScoreboard.Scoreboard.List) != 2 || publicScoreboard.Scoreboard.List[0].TeamID != teamAlpha.ID {
		t.Fatalf("unexpected public scoreboard: %+v", publicScoreboard)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/freeze", scoreboardContest.ID), map[string]any{
		"minutes_before_end": 180,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	seedContestScore(t, env, scoreboardContest.ID, teamBeta.ID, 200)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", scoreboardContest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var frozenScoreboard contesthttp.ScoreboardResp
	decodeFullRouterData(t, resp, &frozenScoreboard)
	if !frozenScoreboard.Frozen || frozenScoreboard.Scoreboard.List[0].TeamID != teamAlpha.ID || frozenScoreboard.Scoreboard.List[0].Score != 100 {
		t.Fatalf("unexpected frozen scoreboard: %+v", frozenScoreboard)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/scoreboard/live", scoreboardContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var liveScoreboard contesthttp.ScoreboardResp
	decodeFullRouterData(t, resp, &liveScoreboard)
	if liveScoreboard.Scoreboard.List[0].TeamID != teamBeta.ID || liveScoreboard.Scoreboard.List[0].Score != 200 {
		t.Fatalf("unexpected live scoreboard: %+v", liveScoreboard)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/unfreeze", scoreboardContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", scoreboardContest.ID), nil, nil)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var unfrozenScoreboard contesthttp.ScoreboardResp
	decodeFullRouterData(t, resp, &unfrozenScoreboard)
	if unfrozenScoreboard.Frozen || unfrozenScoreboard.Scoreboard.List[0].TeamID != teamBeta.ID || unfrozenScoreboard.Scoreboard.List[0].Score != 200 {
		t.Fatalf("unexpected unfrozen scoreboard: %+v", unfrozenScoreboard)
	}

	notFrozenContest := createFullRouterContest(t, env, "Not Frozen Contest", contestcontracts.ContestStatusRunning)
	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/unfreeze", notFrozenContest.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)
}

func TestFullRouter_AdminContestListSupportsModeStatusesSortAndSummary(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
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

	resp := performFullRouterRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/admin/contests?mode=awd&statuses=registration,running,frozen,ended&sort_key=start_time&sort_order=desc&page=1&page_size=2",
		nil,
		adminHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusOK)

	envelope := decodeFullRouterJSON[fullRouterEnvelope](t, resp.Body.Bytes())
	rawPage := decodeFullRouterJSON[map[string]any](t, envelope.Data)
	rawSummary, ok := rawPage["summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected raw summary object, got %#v", rawPage["summary"])
	}
	if _, exists := rawSummary["registration_count"]; exists {
		t.Fatalf("unexpected legacy summary key registration_count in payload: %#v", rawSummary)
	}
	if registeringCount := int(rawSummary["registering_count"].(float64)); registeringCount != 1 {
		t.Fatalf("expected raw payload registering_count=1, got %#v", rawSummary["registering_count"])
	}

	var page contesthttp.ContestPageResp
	decodeFullRouterData(t, resp, &page)

	if page.Total != 4 {
		t.Fatalf("expected filtered total=4, got %d", page.Total)
	}
	if page.Page != 1 || page.PageSize != 2 {
		t.Fatalf("unexpected pagination payload: page=%d size=%d", page.Page, page.PageSize)
	}
	if len(page.List) != 2 {
		t.Fatalf("expected 2 contests on first page, got %d", len(page.List))
	}
	if got := []string{page.List[0].Title, page.List[1].Title}; got[0] != "AWD Running" || got[1] != "AWD Registration" {
		t.Fatalf("unexpected contest ordering: %v", got)
	}
	if page.Summary.RegisteringCount != 1 || page.Summary.RunningCount != 1 || page.Summary.FrozenCount != 1 || page.Summary.EndedCount != 1 {
		t.Fatalf("unexpected summary counts: %+v", page.Summary)
	}
	if page.Summary.DraftCount != 0 {
		t.Fatalf("expected draft count=0 under filtered statuses, got %d", page.Summary.DraftCount)
	}
}
