package app

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	"ctf-platform/internal/shared/taxonomy"
	fullrouterteacherstate "ctf-platform/tests/system/http/fullrouterteacherstate"
)

func TestFullRouter_TeacherAWDReviewExportStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))

	now := time.Now()
	reviewContest := &contestcontracts.Contest{
		Title:       "Teacher AWD Review Matrix",
		Description: "teacher awd review export matrix",
		Mode:        contestcontracts.ContestModeAWD,
		StartTime:   now.Add(-2 * time.Hour),
		EndTime:     now.Add(time.Hour),
		Status:      contestcontracts.ContestStatusRunning,
	}
	if err := env.db.Create(reviewContest).Error; err != nil {
		t.Fatalf("create teacher awd review contest: %v", err)
	}

	if err := env.db.Create(&contestcontracts.ContestChallenge{
		ContestID:   reviewContest.ID,
		ChallengeID: env.challenge.ID,
		Points:      100,
		Order:       1,
		IsVisible:   true,
	}).Error; err != nil {
		t.Fatalf("create teacher awd review contest challenge: %v", err)
	}

	createContestRegistration(t, env, reviewContest.ID, env.student.ID, contestcontracts.ContestRegistrationStatusApproved, nil)
	createContestRegistration(t, env, reviewContest.ID, env.peerStudent.ID, contestcontracts.ContestRegistrationStatusApproved, nil)

	blueTeam := createContestTeam(t, env, reviewContest.ID, env.student.ID, "AWD Review Blue", 4)
	redTeam := createContestTeam(t, env, reviewContest.ID, env.peerStudent.ID, "AWD Review Red", 4)

	if err := env.db.Model(&contestcontracts.Team{}).Where("id = ?", blueTeam.ID).Updates(map[string]any{
		"total_score":   240,
		"last_solve_at": now.Add(-10 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update blue team score: %v", err)
	}
	if err := env.db.Model(&contestcontracts.Team{}).Where("id = ?", redTeam.ID).Updates(map[string]any{
		"total_score":   180,
		"last_solve_at": now.Add(-8 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("update red team score: %v", err)
	}

	round1Start := now.Add(-90 * time.Minute)
	round1End := now.Add(-60 * time.Minute)
	round1 := &contestcontracts.AWDRound{
		ContestID:    reviewContest.ID,
		RoundNumber:  1,
		Status:       contestcontracts.AWDRoundStatusFinished,
		StartedAt:    &round1Start,
		EndedAt:      &round1End,
		AttackScore:  80,
		DefenseScore: 60,
	}
	if err := env.db.Create(round1).Error; err != nil {
		t.Fatalf("create teacher awd review round1: %v", err)
	}

	round2Start := now.Add(-30 * time.Minute)
	round2 := &contestcontracts.AWDRound{
		ContestID:    reviewContest.ID,
		RoundNumber:  2,
		Status:       contestcontracts.AWDRoundStatusRunning,
		StartedAt:    &round2Start,
		AttackScore:  120,
		DefenseScore: 90,
	}
	if err := env.db.Create(round2).Error; err != nil {
		t.Fatalf("create teacher awd review round2: %v", err)
	}

	reviewServiceID := contesttestsupport.DefaultAWDContestServiceID(reviewContest.ID, env.challenge.ID)
	contesttestsupport.SyncAWDContestServiceFixture(
		t,
		env.db,
		reviewContest.ID,
		env.challenge.ID,
		"review-service",
		contestcontracts.AWDCheckerTypeHTTPStandard,
		`{"method":"GET","path":"/health"}`,
		100,
		60,
		40,
		now,
	)

	serviceSeeds := []*contestcontracts.AWDTeamService{
		{RoundID: round1.ID, TeamID: blueTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, ServiceStatus: contestcontracts.AWDServiceStatusUp, AttackReceived: 1, SLAScore: 30, DefenseScore: 40, AttackScore: 20, UpdatedAt: now.Add(-70 * time.Minute)},
		{RoundID: round1.ID, TeamID: redTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, ServiceStatus: contestcontracts.AWDServiceStatusCompromised, AttackReceived: 2, SLAScore: 20, DefenseScore: 30, AttackScore: 15, UpdatedAt: now.Add(-68 * time.Minute)},
		{RoundID: round2.ID, TeamID: blueTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, ServiceStatus: contestcontracts.AWDServiceStatusUp, AttackReceived: 1, SLAScore: 40, DefenseScore: 50, AttackScore: 35, UpdatedAt: now.Add(-12 * time.Minute)},
		{RoundID: round2.ID, TeamID: redTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, ServiceStatus: contestcontracts.AWDServiceStatusDown, AttackReceived: 3, SLAScore: 10, DefenseScore: 15, AttackScore: 10, UpdatedAt: now.Add(-11 * time.Minute)},
	}
	for _, item := range serviceSeeds {
		if err := env.db.Create(item).Error; err != nil {
			t.Fatalf("create teacher awd review team service: %v", err)
		}
	}

	attackSeeds := []*contestcontracts.AWDAttackLog{
		{RoundID: round1.ID, AttackerTeamID: blueTeam.ID, VictimTeamID: redTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 30, CreatedAt: now.Add(-65 * time.Minute)},
		{RoundID: round2.ID, AttackerTeamID: redTeam.ID, VictimTeamID: blueTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceManual, IsSuccess: false, ScoreGained: 0, CreatedAt: now.Add(-10 * time.Minute)},
	}
	for _, item := range attackSeeds {
		if err := env.db.Create(item).Error; err != nil {
			t.Fatalf("create teacher awd review attack log: %v", err)
		}
	}

	trafficSeeds := []*contestcontracts.AWDTrafficEvent{
		{ContestID: reviewContest.ID, RoundID: round1.ID, AttackerTeamID: blueTeam.ID, VictimTeamID: redTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, Method: http.MethodGet, Path: "/health", StatusCode: http.StatusOK, Source: contestcontracts.AWDAttackSourceSubmission, CreatedAt: now.Add(-64 * time.Minute)},
		{ContestID: reviewContest.ID, RoundID: round2.ID, AttackerTeamID: redTeam.ID, VictimTeamID: blueTeam.ID, ServiceID: reviewServiceID, AWDChallengeID: env.challenge.ID, Method: http.MethodPost, Path: "/exploit", StatusCode: http.StatusForbidden, Source: contestcontracts.AWDAttackSourceManual, CreatedAt: now.Add(-9 * time.Minute)},
	}
	for _, item := range trafficSeeds {
		if err := env.db.Create(item).Error; err != nil {
			t.Fatalf("create teacher awd review traffic event: %v", err)
		}
	}

	resp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/awd/reviews", nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var reviewList assessmentqry.TeacherAWDReviewContestPageResp
	decodeFullRouterData(t, resp, &reviewList)
	if reviewList.Page != 1 || reviewList.PageSize != 20 {
		t.Fatalf("expected review list default pagination page=1 page_size=20, got %+v", reviewList)
	}
	foundContest := false
	for _, contest := range reviewList.List {
		if contest.ID != reviewContest.ID {
			continue
		}
		foundContest = true
		if contest.CurrentRound == nil || *contest.CurrentRound != 2 {
			t.Fatalf("expected current round 2, got %+v", contest)
		}
		if contest.ExportReady {
			t.Fatalf("expected running review contest export_ready=false, got %+v", contest)
		}
	}
	if !foundContest {
		t.Fatalf("expected review contest %d in list, got %+v", reviewContest.ID, reviewList)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/awd/reviews/%d?round=2", reviewContest.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var reviewDetail assessmentqry.TeacherAWDReviewArchiveResp
	decodeFullRouterData(t, resp, &reviewDetail)
	if reviewDetail.Scope.SnapshotType != "live" {
		t.Fatalf("expected live snapshot, got %+v", reviewDetail.Scope)
	}
	if reviewDetail.SelectedRound == nil || reviewDetail.SelectedRound.Round.RoundNumber != 2 {
		t.Fatalf("expected selected round 2, got %+v", reviewDetail.SelectedRound)
	}
	if len(reviewDetail.SelectedRound.Teams) != 2 || len(reviewDetail.SelectedRound.Services) == 0 || len(reviewDetail.SelectedRound.Traffic) == 0 {
		t.Fatalf("expected populated selected round payload, got %+v", reviewDetail.SelectedRound)
	}
	if reviewDetail.SelectedRound.Traffic[0].ServiceID != reviewServiceID {
		t.Fatalf("expected selected round traffic service_id=%d, got %+v", reviewServiceID, reviewDetail.SelectedRound.Traffic[0])
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/awd/reviews/%d/export/archive", reviewContest.ID), map[string]any{
		"round_number": 2,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var archiveExport assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &archiveExport)
	if archiveExport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected archive export processing status, got %+v", archiveExport)
	}

	archiveReady := waitForReportStatus(t, env, archiveExport.ReportID, teacherHeaders, assessmententity.ReportStatusReady, 5*time.Second)
	if archiveReady.DownloadURL == nil {
		t.Fatalf("expected archive export download url, got %+v", archiveReady)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", archiveExport.ReportID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/zip" {
		t.Fatalf("expected zip content-type, got %q", contentType)
	}
	if disposition := resp.Header().Get("Content-Disposition"); !strings.Contains(disposition, ".zip") {
		t.Fatalf("expected zip content disposition, got %q", disposition)
	}

	archiveReader, err := zip.NewReader(bytes.NewReader(resp.Body.Bytes()), int64(resp.Body.Len()))
	if err != nil {
		t.Fatalf("open awd review archive zip: %v", err)
	}
	archiveEntries := make(map[string]struct{}, len(archiveReader.File))
	for _, file := range archiveReader.File {
		archiveEntries[file.Name] = struct{}{}
	}
	for _, required := range []string{"manifest.json", "overview.json", "rounds.json", "teams.json", "selected-round.json"} {
		if _, ok := archiveEntries[required]; !ok {
			t.Fatalf("expected zip entry %s, got %+v", required, archiveEntries)
		}
	}
	manifestJSON := readFullRouterZIPEntry(t, archiveReader, "manifest.json")
	manifest := decodeFullRouterJSON[map[string]any](t, manifestJSON)
	if manifest["snapshot_type"] != "live" {
		t.Fatalf("expected manifest snapshot_type=live, got %+v", manifest)
	}
	if selectedRound, ok := manifest["selected_round"].(float64); !ok || int(selectedRound) != 2 {
		t.Fatalf("expected manifest selected_round=2, got %+v", manifest)
	}

	selectedRoundJSON := readFullRouterZIPEntry(t, archiveReader, "selected-round.json")
	if !bytes.Contains(selectedRoundJSON, []byte(`"round_number"`)) || !bytes.Contains(selectedRoundJSON, []byte(`"team_id"`)) {
		t.Fatalf("expected selected-round.json field names to be preserved, got %s", selectedRoundJSON)
	}
	selectedRoundPayload := decodeFullRouterJSON[assessmentqry.TeacherAWDSelectedRoundResp](t, selectedRoundJSON)
	if selectedRoundPayload.Round.RoundNumber != 2 {
		t.Fatalf("expected selected-round.json round_number=2, got %+v", selectedRoundPayload.Round)
	}
	if len(selectedRoundPayload.Teams) != 2 || len(selectedRoundPayload.Services) == 0 {
		t.Fatalf("expected selected-round.json teams/services payload, got %+v", selectedRoundPayload)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/awd/reviews/%d/export/report", reviewContest.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusBadRequest)

	round2End := now.Add(-2 * time.Minute)
	if err := env.db.Model(&contestcontracts.AWDRound{}).Where("id = ?", round2.ID).Updates(map[string]any{
		"status":     contestcontracts.AWDRoundStatusFinished,
		"ended_at":   round2End,
		"updated_at": time.Now(),
	}).Error; err != nil {
		t.Fatalf("end teacher awd review round2: %v", err)
	}
	setContestStatus(t, env, reviewContest.ID, contestcontracts.ContestStatusEnded, nil)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/awd/reviews/%d/export/report", reviewContest.ID), map[string]any{
		"round_number": 2,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var reportExport assessmentcmd.ReportExportData
	decodeFullRouterData(t, resp, &reportExport)
	if reportExport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected report export processing status, got %+v", reportExport)
	}

	reportReady := waitForReportStatus(t, env, reportExport.ReportID, teacherHeaders, assessmententity.ReportStatusReady, 5*time.Second)
	if reportReady.DownloadURL == nil {
		t.Fatalf("expected report export download url, got %+v", reportReady)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", reportExport.ReportID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/pdf" {
		t.Fatalf("expected pdf content-type, got %q", contentType)
	}
	if disposition := resp.Header().Get("Content-Disposition"); !strings.Contains(disposition, ".pdf") {
		t.Fatalf("expected pdf content disposition, got %q", disposition)
	}
	if !bytes.HasPrefix(resp.Body.Bytes(), []byte("%PDF")) {
		t.Fatalf("expected pdf body prefix, got %q", resp.Body.Bytes())
	}
	for _, token := range [][]byte{
		[]byte("Teacher AWD Review Report"),
		[]byte("Selected Round"),
	} {
		if !fullRouterPDFContainsText(resp.Body.Bytes(), string(token)) {
			t.Fatalf("expected pdf body to contain %q", token)
		}
	}
}

func TestFullRouter_TeacherAccessAndRecommendationStateMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)
	createRecommendationChallenge(t, env, "Matrix Weak Web 2", taxonomy.DimensionWeb)

	fullrouterteacherstate.VerifyTeacherAccessAndRecommendationStateMatrix(
		t,
		fullrouterteacherstate.TeacherAccessAndRecommendationStateMatrixDriver{
			Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterRequest(t, env.router, method, target, payload, headers)
			},
			AdminHeaders:   bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
			TeacherHeaders: bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd)),
			StudentHeaders: bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd)),
			ClassName:      env.className,
			OtherClassName: env.otherStudent.ClassName,
			StudentID:      env.student.ID,
			OtherStudentID: env.otherStudent.ID,
		},
	)
}

func TestFullRouter_ChallengeWriteupsUseCommunitySemantics(t *testing.T) {
	env := newFullRouterTestEnv(t)

	fullrouterteacherstate.VerifyChallengeWriteupsUseCommunitySemantics(
		t,
		fullrouterteacherstate.ChallengeWriteupsUseCommunitySemanticsDriver{
			Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterRequest(t, env.router, method, target, payload, headers)
			},
			AdminHeaders:   bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd)),
			TeacherHeaders: bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd)),
			StudentHeaders: bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123")),
			CreateChallengePayload: map[string]any{
				"title":       "Community Writeup Challenge",
				"description": "community writeup semantics",
				"category":    taxonomy.DimensionWeb,
				"difficulty":  taxonomy.DifficultyEasy,
				"points":      80,
			},
			SetChallengePublished: func(t *testing.T, challengeID int64) {
				if err := env.db.Model(&appChallengeRow{}).
					Where("id = ?", challengeID).
					Update("status", challengecontracts.ChallengeStatusPublished).Error; err != nil {
					t.Fatalf("set challenge published: %v", err)
				}
			},
			CreatePracticeSubmission: func(t *testing.T, challengeID int64) {
				createPracticeSubmission(t, env, env.peerStudent.ID, challengeID, 80)
			},
		},
	)
}
