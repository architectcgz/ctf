package fullrouterconteststate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder

type RegistrationSnapshot struct {
	ID     int64
	Status string
}

type TeamSnapshot struct {
	ID        int64
	CaptainID int64
}

type ContestParticipationStateMatrixDriver struct {
	Request               RequestFunc
	AdminHeaders          map[string]string
	StudentHeaders        map[string]string
	PeerHeaders           map[string]string
	OtherHeaders          map[string]string
	RetryHeaders          map[string]string
	RegistrationContestID int64
	ExistingContestID     int64
	ExistingTeamID        int64
	ExistingChallengeID   int64
	StudentID             int64
	PeerStudentID         int64
	OtherStudentID        int64
	RetryStudentID        int64
	FillerStudentID       int64
	FindRegistration      func(t *testing.T, contestID, userID int64) RegistrationSnapshot
	CreateRegistration    func(t *testing.T, contestID, userID int64, status string)
	CreateTeam            func(t *testing.T, contestID, captainID int64, name string, maxMembers int) TeamSnapshot
	AddTeamMember         func(t *testing.T, contestID, teamID, userID int64)
	CreateSubmission      func(t *testing.T, contestID, teamID, userID, challengeID int64, points int)
}

type ContestAndReviewArchiveExportStateMatrixDriver struct {
	Request             RequestFunc
	AdminHeaders        map[string]string
	TeacherHeaders      map[string]string
	OtherTeacherHeaders map[string]string
	StudentHeaders      map[string]string
	ContestID           int64
	StudentID           int64
	StudentUsername     string
	WaitForReportStatus func(t *testing.T, reportID int64, headers map[string]string) *assessmentcmd.ReportExportData
}

type ContestChallengeAndScoreboardStateMatrixDriver struct {
	Request                  RequestFunc
	AdminHeaders             map[string]string
	StudentHeaders           map[string]string
	PeerHeaders              map[string]string
	OtherHeaders             map[string]string
	EditableContestID        int64
	ConflictContestID        int64
	ScoreboardContestID      int64
	NotFrozenContestID       int64
	ChallengeAID             int64
	ChallengeBID             int64
	StudentID                int64
	PeerStudentID            int64
	TeamAlphaID              int64
	TeamBetaID               int64
	ConflictSubmissionTeamID int64
	CreateRegistration       func(t *testing.T, contestID, userID int64, status string)
	SetContestStatus         func(t *testing.T, contestID int64, status string)
	CreateSubmission         func(t *testing.T, contestID, teamID, userID, challengeID int64, points int)
	DeleteSubmissions        func(t *testing.T, contestID, challengeID int64)
	SeedScore                func(t *testing.T, contestID, teamID int64, score int)
}

type AdminContestListSupportsModeStatusesSortAndSummaryDriver struct {
	Request      RequestFunc
	AdminHeaders map[string]string
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func VerifyContestParticipationStateMatrix(t *testing.T, driver ContestParticipationStateMatrixDriver) {
	t.Helper()

	resp := driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", driver.RegistrationContestID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusOK)

	peerRegistration := driver.FindRegistration(t, driver.RegistrationContestID, driver.PeerStudentID)
	if peerRegistration.Status != contestcontracts.ContestRegistrationStatusPending {
		t.Fatalf("expected pending registration, got %s", peerRegistration.Status)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", driver.RegistrationContestID), map[string]any{
		"name":        "PendingTeam",
		"max_members": 3,
	}, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/registrations?status=pending", driver.RegistrationContestID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)
	var registrationPage map[string]any
	decodeEnvelopeData(t, resp, &registrationPage)
	if total := int(registrationPage["total"].(float64)); total != 1 {
		t.Fatalf("expected 1 pending registration, got %d", total)
	}

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/admin/contests/%d/registrations/%d", driver.RegistrationContestID, peerRegistration.ID), map[string]any{
		"status": contestcontracts.ContestRegistrationStatusApproved,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", driver.ExistingContestID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	driver.CreateRegistration(t, driver.RegistrationContestID, driver.StudentID, contestcontracts.ContestRegistrationStatusApproved)
	driver.CreateRegistration(t, driver.RegistrationContestID, driver.OtherStudentID, contestcontracts.ContestRegistrationStatusApproved)
	driver.CreateRegistration(t, driver.RegistrationContestID, driver.FillerStudentID, contestcontracts.ContestRegistrationStatusApproved)
	driver.CreateRegistration(t, driver.RegistrationContestID, driver.RetryStudentID, contestcontracts.ContestRegistrationStatusRejected)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/register", driver.RegistrationContestID), nil, driver.RetryHeaders)
	assertStatus(t, resp, http.StatusOK)
	retryRegistration := driver.FindRegistration(t, driver.RegistrationContestID, driver.RetryStudentID)
	if retryRegistration.Status != contestcontracts.ContestRegistrationStatusPending {
		t.Fatalf("expected rejected registration to requeue as pending, got %s", retryRegistration.Status)
	}

	fullTeam := driver.CreateTeam(t, driver.RegistrationContestID, driver.OtherStudentID, "FullTeam", 2)
	driver.AddTeamMember(t, driver.RegistrationContestID, fullTeam.ID, driver.FillerStudentID)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams/%d/join", driver.RegistrationContestID, fullTeam.ID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", driver.RegistrationContestID), map[string]any{
		"name":        "AlphaTeam",
		"max_members": 4,
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var createdTeam contesthttp.TeamResp
	decodeEnvelopeData(t, resp, &createdTeam)
	if createdTeam.CaptainID != driver.StudentID {
		t.Fatalf("expected student captain id %d, got %d", driver.StudentID, createdTeam.CaptainID)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams", driver.RegistrationContestID), map[string]any{
		"name":        "DuplicateTeam",
		"max_members": 4,
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusConflict)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/teams/%d/join", driver.RegistrationContestID, createdTeam.ID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d", driver.RegistrationContestID, createdTeam.ID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d/leave", driver.RegistrationContestID, createdTeam.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d/leave", driver.RegistrationContestID, createdTeam.ID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/contests/%d/teams/%d", driver.RegistrationContestID, createdTeam.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	driver.CreateSubmission(t, driver.ExistingContestID, driver.ExistingTeamID, driver.StudentID, driver.ExistingChallengeID, 100)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/my-progress", driver.ExistingContestID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var progress contesthttp.ContestMyProgressResp
	decodeEnvelopeData(t, resp, &progress)
	if progress.ContestID != driver.ExistingContestID || len(progress.Solved) == 0 {
		t.Fatalf("expected existing contest progress, got %+v", progress)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/my-progress", driver.ExistingContestID), nil, driver.OtherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var emptyProgress contesthttp.ContestMyProgressResp
	decodeEnvelopeData(t, resp, &emptyProgress)
	if emptyProgress.TeamID != nil || len(emptyProgress.Solved) != 0 {
		t.Fatalf("expected empty progress for unregistered student, got %+v", emptyProgress)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/announcements", driver.ExistingContestID), nil, nil)
	assertStatus(t, resp, http.StatusOK)

	var announcements []contesthttp.ContestAnnouncementResp
	decodeEnvelopeData(t, resp, &announcements)
	if len(announcements) == 0 {
		t.Fatalf("expected seeded announcement")
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/announcements", driver.ExistingContestID), map[string]any{
		"title":   "新的公告",
		"content": "integration notice",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var createdAnnouncement contesthttp.ContestAnnouncementResp
	decodeEnvelopeData(t, resp, &createdAnnouncement)
	if createdAnnouncement.Title != "新的公告" {
		t.Fatalf("unexpected announcement title: %+v", createdAnnouncement)
	}

	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/announcements/%d", driver.ExistingContestID, createdAnnouncement.ID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)
}

func VerifyContestAndReviewArchiveExportStateMatrix(t *testing.T, driver ContestAndReviewArchiveExportStateMatrixDriver) {
	t.Helper()

	resp := driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/export", driver.ContestID), map[string]any{
		"format": "json",
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	missingContestResp := driver.Request(http.MethodPost, "/api/v1/admin/contests/999999/export", map[string]any{
		"format": "json",
	}, driver.AdminHeaders)
	assertStatus(t, missingContestResp, http.StatusNotFound)

	var contestExport assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &contestExport)
	if contestExport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected contest export to start in processing state, got %+v", contestExport)
	}

	contestReady := driver.WaitForReportStatus(t, contestExport.ReportID, driver.AdminHeaders)
	if contestReady == nil || contestReady.DownloadURL == nil {
		t.Fatalf("expected contest export download url after ready, got %+v", contestReady)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", contestExport.ReportID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type for contest export, got %q", contentType)
	}
	contestBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read contest export body: %v", err)
	}
	if !bytes.Contains(contestBody, []byte(`"contest"`)) {
		t.Fatalf("expected contest export payload to contain contest metadata, got %s", string(contestBody))
	}
	if !bytes.Contains(contestBody, []byte(`"solved_count": 2`)) {
		t.Fatalf("expected contest export scoreboard solved_count to include multiple solved challenges, got %s", string(contestBody))
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", driver.StudentID), map[string]any{
		"format": "json",
	}, driver.OtherTeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", driver.StudentID), map[string]any{
		"format": "json",
	}, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var reviewArchive assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &reviewArchive)
	if reviewArchive.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected review archive export to start in processing state, got %+v", reviewArchive)
	}

	reviewReady := driver.WaitForReportStatus(t, reviewArchive.ReportID, driver.TeacherHeaders)
	if reviewReady == nil || reviewReady.DownloadURL == nil {
		t.Fatalf("expected review archive download url after ready, got %+v", reviewReady)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", reviewArchive.ReportID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type for review archive, got %q", contentType)
	}
	reviewBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read review archive body: %v", err)
	}
	if !bytes.Contains(reviewBody, []byte(`"student"`)) || !bytes.Contains(reviewBody, []byte(driver.StudentUsername)) {
		t.Fatalf("expected review archive payload to contain student metadata, got %s", string(reviewBody))
	}
	if !bytes.Contains(reviewBody, []byte(`"teacher_observations"`)) {
		t.Fatalf("expected review archive payload to contain teacher observations, got %s", string(reviewBody))
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", driver.StudentID), nil, driver.OtherTeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", driver.StudentID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)
	if body, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("read review archive view body: %v", err)
	} else if !bytes.Contains(body, []byte(`"teacher_observations"`)) || !bytes.Contains(body, []byte(driver.StudentUsername)) {
		t.Fatalf("expected review archive view payload to contain observations and student username, got %s", string(body))
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", driver.StudentID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive", driver.StudentID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusForbidden)
}

func VerifyContestChallengeAndScoreboardStateMatrix(t *testing.T, driver ContestChallengeAndScoreboardStateMatrixDriver) {
	t.Helper()

	resp := driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", driver.EditableContestID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", driver.EditableContestID), map[string]any{
		"challenge_id": driver.ChallengeAID,
		"points":       220,
		"order":        1,
		"is_visible":   false,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var contestChallenge contesthttp.ContestChallengeResp
	decodeEnvelopeData(t, resp, &contestChallenge)
	if contestChallenge.Points != 220 || contestChallenge.IsVisible {
		t.Fatalf("unexpected contest challenge: %+v", contestChallenge)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", driver.EditableContestID), map[string]any{
		"challenge_id": driver.ChallengeAID,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusConflict)

	resp = driver.Request(http.MethodPut, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", driver.EditableContestID, driver.ChallengeAID), map[string]any{
		"points":     260,
		"order":      2,
		"is_visible": true,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", driver.EditableContestID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var adminChallenges []contesthttp.ContestChallengeResp
	decodeEnvelopeData(t, resp, &adminChallenges)
	if len(adminChallenges) != 1 || adminChallenges[0].Points != 260 || adminChallenges[0].Order != 2 || !adminChallenges[0].IsVisible {
		t.Fatalf("unexpected admin contest challenges: %+v", adminChallenges)
	}

	driver.CreateRegistration(t, driver.EditableContestID, driver.StudentID, contestcontracts.ContestRegistrationStatusApproved)
	driver.CreateRegistration(t, driver.EditableContestID, driver.PeerStudentID, contestcontracts.ContestRegistrationStatusPending)
	driver.SetContestStatus(t, driver.EditableContestID, contestcontracts.ContestStatusRunning)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/challenges", driver.EditableContestID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var visibleChallenges []contesthttp.ContestChallengeInfo
	decodeEnvelopeData(t, resp, &visibleChallenges)
	if len(visibleChallenges) != 1 || visibleChallenges[0].ChallengeID != driver.ChallengeAID || visibleChallenges[0].Points != 260 {
		t.Fatalf("unexpected visible contest challenges: %+v", visibleChallenges)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", driver.EditableContestID, driver.ChallengeAID), nil, driver.OtherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", driver.EditableContestID, driver.ChallengeAID), nil, driver.PeerHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/contests/%d/challenges/%d/instances", driver.EditableContestID, driver.ChallengeAID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var startedContestInstance instancecontracts.InstanceResp
	decodeEnvelopeData(t, resp, &startedContestInstance)
	if startedContestInstance.ChallengeID != driver.ChallengeAID {
		t.Fatalf("unexpected started contest instance: %+v", startedContestInstance)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", driver.EditableContestID), map[string]any{
		"challenge_id": driver.ChallengeBID,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/challenges", driver.ConflictContestID), map[string]any{
		"challenge_id": driver.ChallengeBID,
		"points":       180,
		"order":        1,
		"is_visible":   true,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	driver.CreateSubmission(t, driver.ConflictContestID, driver.ConflictSubmissionTeamID, driver.StudentID, driver.ChallengeBID, 180)
	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", driver.ConflictContestID, driver.ChallengeBID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusConflict)

	driver.DeleteSubmissions(t, driver.ConflictContestID, driver.ChallengeBID)
	resp = driver.Request(http.MethodDelete, fmt.Sprintf("/api/v1/admin/contests/%d/challenges/%d", driver.ConflictContestID, driver.ChallengeBID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	driver.SeedScore(t, driver.ScoreboardContestID, driver.TeamAlphaID, 100)
	driver.SeedScore(t, driver.ScoreboardContestID, driver.TeamBetaID, 80)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", driver.ScoreboardContestID), nil, nil)
	assertStatus(t, resp, http.StatusOK)

	var publicScoreboard contesthttp.ScoreboardResp
	decodeEnvelopeData(t, resp, &publicScoreboard)
	if publicScoreboard.Frozen || len(publicScoreboard.Scoreboard.List) != 2 || publicScoreboard.Scoreboard.List[0].TeamID != driver.TeamAlphaID {
		t.Fatalf("unexpected public scoreboard: %+v", publicScoreboard)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/freeze", driver.ScoreboardContestID), map[string]any{
		"minutes_before_end": 180,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	driver.SeedScore(t, driver.ScoreboardContestID, driver.TeamBetaID, 200)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", driver.ScoreboardContestID), nil, nil)
	assertStatus(t, resp, http.StatusOK)

	var frozenScoreboard contesthttp.ScoreboardResp
	decodeEnvelopeData(t, resp, &frozenScoreboard)
	if !frozenScoreboard.Frozen || frozenScoreboard.Scoreboard.List[0].TeamID != driver.TeamAlphaID || frozenScoreboard.Scoreboard.List[0].Score != 100 {
		t.Fatalf("unexpected frozen scoreboard: %+v", frozenScoreboard)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/admin/contests/%d/scoreboard/live", driver.ScoreboardContestID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var liveScoreboard contesthttp.ScoreboardResp
	decodeEnvelopeData(t, resp, &liveScoreboard)
	if liveScoreboard.Scoreboard.List[0].TeamID != driver.TeamBetaID || liveScoreboard.Scoreboard.List[0].Score != 200 {
		t.Fatalf("unexpected live scoreboard: %+v", liveScoreboard)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/unfreeze", driver.ScoreboardContestID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/contests/%d/scoreboard", driver.ScoreboardContestID), nil, nil)
	assertStatus(t, resp, http.StatusOK)

	var unfrozenScoreboard contesthttp.ScoreboardResp
	decodeEnvelopeData(t, resp, &unfrozenScoreboard)
	if unfrozenScoreboard.Frozen || unfrozenScoreboard.Scoreboard.List[0].TeamID != driver.TeamBetaID || unfrozenScoreboard.Scoreboard.List[0].Score != 200 {
		t.Fatalf("unexpected unfrozen scoreboard: %+v", unfrozenScoreboard)
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/unfreeze", driver.NotFrozenContestID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusBadRequest)
}

func VerifyAdminContestListSupportsModeStatusesSortAndSummary(t *testing.T, driver AdminContestListSupportsModeStatusesSortAndSummaryDriver) {
	t.Helper()

	resp := driver.Request(
		http.MethodGet,
		"/api/v1/admin/contests?mode=awd&statuses=registration,running,frozen,ended&sort_key=start_time&sort_order=desc&page=1&page_size=2",
		nil,
		driver.AdminHeaders,
	)
	assertStatus(t, resp, http.StatusOK)

	var env envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, resp.Body.String())
	}
	var rawPage map[string]any
	if err := json.Unmarshal(env.Data, &rawPage); err != nil {
		t.Fatalf("decode raw contest page: %v body=%s", err, resp.Body.String())
	}
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
	if err := json.Unmarshal(env.Data, &page); err != nil {
		t.Fatalf("decode contest page: %v body=%s", err, resp.Body.String())
	}
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

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()

	if resp.Code != want {
		t.Fatalf("expected status %d, got %d body=%s", want, resp.Code, resp.Body.String())
	}
}

func decodeEnvelopeData(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	var body envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, resp.Body.String())
	}
	if len(body.Data) == 0 || string(body.Data) == "null" {
		t.Fatalf("expected response data, got empty body=%s", resp.Body.String())
	}
	if err := json.Unmarshal(body.Data, target); err != nil {
		t.Fatalf("decode response data: %v body=%s", err, resp.Body.String())
	}
}
