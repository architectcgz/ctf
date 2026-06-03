package app

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"
	"ctf-platform/internal/shared/taxonomy"
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

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.student.Username, env.studentPwd))

	resp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/classes", nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var teacherClasses struct {
		List     []teachingqueryqueries.TeacherClassItem `json:"list"`
		Total    int64                                   `json:"total"`
		Page     int                                     `json:"page"`
		PageSize int                                     `json:"page_size"`
	}
	decodeFullRouterData(t, resp, &teacherClasses)
	if teacherClasses.Total != 1 || len(teacherClasses.List) != 1 || teacherClasses.List[0].Name != env.className {
		t.Fatalf("expected only teacher class page, got %+v", teacherClasses)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/classes?page=1&page_size=1", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var adminClasses struct {
		List     []teachingqueryqueries.TeacherClassItem `json:"list"`
		Total    int64                                   `json:"total"`
		Page     int                                     `json:"page"`
		PageSize int                                     `json:"page_size"`
	}
	decodeFullRouterData(t, resp, &adminClasses)
	if adminClasses.Page != 1 || adminClasses.PageSize != 1 || len(adminClasses.List) != 1 || adminClasses.Total < 2 {
		t.Fatalf("expected admin class pagination, got %+v", adminClasses)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/teacher/students?page=1&page_size=1&sort_key=total_score&sort_order=desc", nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var studentDirectory struct {
		List []struct {
			ID         int64   `json:"id"`
			Username   string  `json:"username"`
			ClassName  *string `json:"class_name"`
			TotalScore int     `json:"total_score"`
		} `json:"list"`
		Total    int64 `json:"total"`
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
	}
	decodeFullRouterData(t, resp, &studentDirectory)
	if studentDirectory.Page != 1 || studentDirectory.PageSize != 1 || studentDirectory.Total < 2 || len(studentDirectory.List) != 1 {
		t.Fatalf("expected paged teacher student directory, got %+v", studentDirectory)
	}
	if studentDirectory.List[0].ClassName == nil || *studentDirectory.List[0].ClassName == "" {
		t.Fatalf("expected class_name in student directory item, got %+v", studentDirectory.List[0])
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/summary", env.className), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/summary", env.otherStudent.ClassName), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/trend", env.className), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/classes/%s/review", env.className), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/progress", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var progress teachingqueryqueries.TeacherProgressResp
	decodeFullRouterData(t, resp, &progress)
	if progress.SolvedChallenges == 0 {
		t.Fatalf("expected solved challenges in teacher progress, got %+v", progress)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/timeline", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var timeline teachingqueryqueries.TimelineResp
	decodeFullRouterData(t, resp, &timeline)
	if len(timeline.Events) == 0 {
		t.Fatalf("expected timeline events, got %+v", timeline)
	}
	firstTimelineEvent := timeline.Events[0]
	if firstTimelineEvent.ChallengeID == 0 || firstTimelineEvent.Timestamp.IsZero() {
		t.Fatalf("expected populated timeline event fields, got %+v", firstTimelineEvent)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/recommendations", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var teacherRecommendations teachingqueryqueries.TeacherRecommendationResp
	decodeFullRouterData(t, resp, &teacherRecommendations)
	if len(teacherRecommendations.Challenges) == 0 {
		t.Fatalf("expected teacher recommendations, got %+v", teacherRecommendations)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/skill-profile", env.student.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var skillProfile assessmenthttp.SkillProfileResp
	decodeFullRouterData(t, resp, &skillProfile)
	if skillProfile.UserID != env.student.ID {
		t.Fatalf("expected skill profile for student %d, got %+v", env.student.ID, skillProfile)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/teacher/students/%d/progress", env.otherStudent.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/users/%d/skill-profile", env.otherStudent.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusForbidden)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/users/me/recommendations", nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var selfRecommendations assessmenthttp.RecommendationResp
	decodeFullRouterData(t, resp, &selfRecommendations)
	if len(selfRecommendations.Challenges) == 0 {
		t.Fatalf("expected self recommendations, got %+v", selfRecommendations)
	}
}

func TestFullRouter_ChallengeWriteupsUseCommunitySemantics(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := bearerHeaders(loginForToken(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := bearerHeaders(loginForToken(t, env.router, env.teacher.Username, env.teacherPwd))
	studentHeaders := bearerHeaders(loginForToken(t, env.router, env.peerStudent.Username, "Password123"))

	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "Community Writeup Challenge",
		"description": "community writeup semantics",
		"category":    taxonomy.DimensionWeb,
		"difficulty":  taxonomy.DifficultyEasy,
		"points":      80,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var createdChallenge challengehttp.ChallengeResp
	decodeFullRouterData(t, resp, &createdChallenge)

	resp = performFullRouterRequest(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", createdChallenge.ID), map[string]any{
		"title":      "Official Solution",
		"content":    "official content",
		"visibility": challengeentity.WriteupVisibilityPublic,
	}, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	if err := env.db.Model(&appChallengeRow{}).
		Where("id = ?", createdChallenge.ID).
		Update("status", challengecontracts.ChallengeStatusPublished).Error; err != nil {
		t.Fatalf("set challenge published: %v", err)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/writeup", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var officialPayload map[string]any
	decodeFullRouterData(t, resp, &officialPayload)
	if _, ok := officialPayload["is_recommended"]; !ok {
		t.Fatalf("expected official writeup payload to expose is_recommended, got %+v", officialPayload)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "我的草稿",
		"content":           "先记入口，再写利用链。",
		"submission_status": challengeentity.SubmissionWriteupStatusDraft,
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var draftPayload map[string]any
	decodeFullRouterData(t, resp, &draftPayload)
	if _, ok := draftPayload["review_status"]; ok {
		t.Fatalf("expected community writeup payload to drop review_status, got %+v", draftPayload)
	}

	createPracticeSubmission(t, env, env.peerStudent.ID, createdChallenge.ID, 80)

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/challenges/%d/writeup-submissions", createdChallenge.ID), map[string]any{
		"title":             "我的题解",
		"content":           "1. 找入口 2. 构造 payload 3. 读取 flag",
		"submission_status": "published",
	}, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var publishedPayload map[string]any
	decodeFullRouterData(t, resp, &publishedPayload)
	if publishedPayload["submission_status"] != "published" {
		t.Fatalf("expected published submission status, got %+v", publishedPayload)
	}
	if _, ok := publishedPayload["published_at"]; !ok {
		t.Fatalf("expected published writeup payload to expose published_at, got %+v", publishedPayload)
	}
	submissionID := int64(publishedPayload["id"].(float64))

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup/recommend", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var recommendedOfficial challengecontracts.AdminChallengeWriteupResp
	decodeFullRouterData(t, resp, &recommendedOfficial)
	if !recommendedOfficial.IsRecommended {
		t.Fatalf("expected official writeup to become recommended, got %+v", recommendedOfficial)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/recommend", submissionID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var recommendedCommunity challengecontracts.SubmissionWriteupResp
	decodeFullRouterData(t, resp, &recommendedCommunity)
	if !recommendedCommunity.IsRecommended {
		t.Fatalf("expected community writeup to become recommended, got %+v", recommendedCommunity)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/hide", submissionID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var hiddenCommunity challengecontracts.SubmissionWriteupResp
	decodeFullRouterData(t, resp, &hiddenCommunity)
	if hiddenCommunity.VisibilityStatus != challengeentity.SubmissionWriteupVisibilityHidden {
		t.Fatalf("expected hidden community writeup, got %+v", hiddenCommunity)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/solutions/community", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var hiddenCommunityList struct {
		List []map[string]any `json:"list"`
	}
	decodeFullRouterData(t, resp, &hiddenCommunityList)
	if len(hiddenCommunityList.List) != 0 {
		t.Fatalf("expected hidden community writeup to disappear from community list, got %+v", hiddenCommunityList)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/restore", submissionID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var restoredCommunity challengecontracts.SubmissionWriteupResp
	decodeFullRouterData(t, resp, &restoredCommunity)
	if restoredCommunity.VisibilityStatus != challengeentity.SubmissionWriteupVisibilityVisible {
		t.Fatalf("expected restored community writeup, got %+v", restoredCommunity)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/recommend", submissionID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/solutions/recommended", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var recommendedList struct {
		List []map[string]any `json:"list"`
	}
	decodeFullRouterData(t, resp, &recommendedList)
	if len(recommendedList.List) != 2 {
		t.Fatalf("expected recommended solutions list, got %+v", recommendedList)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/challenges/%d/solutions/community", createdChallenge.ID), nil, studentHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var communityList struct {
		List []map[string]any `json:"list"`
	}
	decodeFullRouterData(t, resp, &communityList)
	if len(communityList.List) != 1 {
		t.Fatalf("expected exactly one community solution, got %+v", communityList)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/teacher/community-writeups/%d/recommend", submissionID), nil, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var unrecommendedCommunity challengecontracts.SubmissionWriteupResp
	decodeFullRouterData(t, resp, &unrecommendedCommunity)
	if unrecommendedCommunity.IsRecommended {
		t.Fatalf("expected community writeup recommendation to be cleared, got %+v", unrecommendedCommunity)
	}

	resp = performFullRouterRequest(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup/recommend", createdChallenge.ID), nil, adminHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var unrecommendedOfficial challengecontracts.AdminChallengeWriteupResp
	decodeFullRouterData(t, resp, &unrecommendedOfficial)
	if unrecommendedOfficial.IsRecommended {
		t.Fatalf("expected official writeup recommendation to be cleared, got %+v", unrecommendedOfficial)
	}
}
