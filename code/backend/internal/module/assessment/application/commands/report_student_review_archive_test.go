package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

func TestValidateStudentReviewArchiveAccess(t *testing.T) {
	t.Parallel()

	teacher := &assessmentdomain.ReportUser{ID: 1, Role: identitycontracts.RoleTeacher, ClassName: "class-a"}
	admin := &assessmentdomain.ReportUser{ID: 2, Role: identitycontracts.RoleAdmin}
	student := &assessmentdomain.ReportUser{ID: 3, Role: identitycontracts.RoleStudent, ClassName: "class-a"}
	otherStudent := &assessmentdomain.ReportUser{ID: 4, Role: identitycontracts.RoleStudent, ClassName: "class-b"}

	if err := validateStudentReviewArchiveAccess(teacher, student); err != nil {
		t.Fatalf("expected same-class teacher access, got %v", err)
	}
	if err := validateStudentReviewArchiveAccess(admin, otherStudent); err != nil {
		t.Fatalf("expected admin access, got %v", err)
	}

	err := validateStudentReviewArchiveAccess(teacher, otherStudent)
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != apperror.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %#v", err)
	}
}

func TestBuildStudentReviewArchiveDataIncludesTeachingObservations(t *testing.T) {
	t.Parallel()

	submittedAt := time.Date(2026, 4, 1, 9, 12, 0, 0, time.UTC)
	reviewedAt := submittedAt.Add(8 * time.Minute)
	lastEventAt := time.Date(2026, 4, 1, 9, 20, 0, 0, time.UTC)
	wrong := false
	correct := true

	repo := &testReportRepository{
		users: map[int64]*assessmentdomain.ReportUser{
			7: {
				ID:        7,
				Username:  "alice",
				Name:      "Alice",
				ClassName: "class-a",
				Role:      identitycontracts.RoleStudent,
			},
		},
		personalStats: &assessmentdomain.PersonalReportStats{
			TotalScore:    100,
			TotalSolved:   1,
			TotalAttempts: 4,
			Rank:          2,
		},
		totalChallenges: 5,
		timeline: []assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:        "hint_unlock",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   submittedAt,
				Detail:      "解锁第 1 级提示",
			},
			{
				Type:        "flag_submit",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   submittedAt.Add(3 * time.Minute),
				IsCorrect:   &wrong,
				Detail:      "提交未命中 Flag",
			},
			{
				Type:        "flag_submit",
				ChallengeID: 11,
				Title:       "web-1",
				Timestamp:   lastEventAt,
				IsCorrect:   &correct,
				Points:      intPtr(100),
				Detail:      "提交命中 Flag",
			},
		},
		evidence: []assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:        "instance_access",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   submittedAt.Add(1 * time.Minute),
				Detail:      "访问攻击目标",
				Meta:        map[string]any{"event_stage": "access"},
			},
			{
				Type:        "instance_proxy_request",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   submittedAt.Add(2 * time.Minute),
				Detail:      "经平台代理发起 POST /login",
				Meta:        map[string]any{"event_stage": "exploit", "method": "POST"},
			},
			{
				Type:        "challenge_hint_unlock",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   submittedAt,
				Detail:      "解锁第 1 级提示",
				Meta:        map[string]any{"event_stage": "analysis"},
			},
			{
				Type:        "challenge_submission",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   lastEventAt,
				Detail:      "提交命中 Flag",
				Meta:        map[string]any{"event_stage": "submit", "is_correct": true, "points": 100},
			},
		},
		writeups: []assessmentdomain.ReviewArchiveWriteupItem{
			{
				ID:               1,
				ChallengeID:      11,
				Category:         "web",
				ChallengeTitle:   "web-1",
				Title:            "从回显到 flag",
				SubmissionStatus: "published",
				VisibilityStatus: "visible",
				IsRecommended:    true,
				PublishedAt:      &submittedAt,
				UpdatedAt:        reviewedAt,
			},
		},
		manualReviews: []assessmentdomain.ReviewArchiveManualReviewItem{
			{
				ID:             2,
				ChallengeID:    12,
				Category:       "misc",
				ChallengeTitle: "misc-essay",
				Answer:         "完整答案正文",
				ReviewStatus:   "approved",
				SubmittedAt:    submittedAt,
				ReviewedAt:     &reviewedAt,
				ReviewComment:  "通过",
				Score:          100,
				ReviewerName:   "teacher-a",
			},
		},
	}

	service := NewReportService(
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		&testAssessmentProfileReader{
			resp: &assessmentcontracts.SkillProfile{
				UserID: 7,
				Dimensions: []*assessmentcontracts.SkillDimension{
					{Dimension: "web", Score: 0.8},
				},
				UpdatedAt: submittedAt.Format(time.RFC3339),
			},
		},
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	data, err := service.buildStudentReviewArchiveData(context.Background(), 7)
	if err != nil {
		t.Fatalf("buildStudentReviewArchiveData() error = %v", err)
	}

	summaryPayload, err := json.Marshal(data.Summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if bytes.Contains(summaryPayload, []byte(`"hint_unlock_count"`)) {
		t.Fatalf("expected summary payload to omit hint_unlock_count, got %s", string(summaryPayload))
	}
	if data.Summary.CorrectSubmissionCount != 1 {
		t.Fatalf("expected 1 correct submission, got %d", data.Summary.CorrectSubmissionCount)
	}
	if data.Summary.WriteupCount != 1 {
		t.Fatalf("expected 1 writeup, got %d", data.Summary.WriteupCount)
	}
	if data.Summary.LastActivityAt == nil || !data.Summary.LastActivityAt.Equal(lastEventAt) {
		t.Fatalf("expected last activity at %s, got %#v", lastEventAt, data.Summary.LastActivityAt)
	}

	if len(data.TeacherObservations.Items) == 0 {
		t.Fatal("expected teaching observations to be generated")
	}

	reflection := findObservation(data.TeacherObservations.Items, "reflection_status")
	if reflection == nil || reflection.Severity != "good" {
		t.Fatalf("expected reflection_status observation, got %#v", reflection)
	}

	awdSummary := findObservation(data.TeacherObservations.Items, "awd_summary")
	if awdSummary != nil {
		t.Fatalf("expected no awd_summary without awd evidence, got %#v", awdSummary)
	}
}

func TestBuildReviewArchiveSummaryCountsAWDAttackEvents(t *testing.T) {
	t.Parallel()

	success := true
	now := time.Date(2026, 4, 13, 15, 0, 0, 0, time.UTC)

	summary := buildReviewArchiveSummary(
		6,
		&assessmentdomain.PersonalReportStats{
			TotalScore:    150,
			TotalSolved:   1,
			TotalAttempts: 3,
			Rank:          1,
		},
		[]assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:        "awd_attack_submit",
				ChallengeID: 91,
				Title:       "awd-web",
				Timestamp:   now,
				IsCorrect:   &success,
				Points:      intPtr(150),
				Detail:      "AWD 攻击命中 blue-team，得分 150",
			},
		},
		nil,
		nil,
		nil,
	)

	if summary.CorrectSubmissionCount != 1 {
		t.Fatalf("expected AWD timeline event counted as correct submission, got %+v", summary)
	}
	if summary.LastActivityAt == nil || !summary.LastActivityAt.Equal(now) {
		t.Fatalf("expected last activity at %s, got %+v", now, summary.LastActivityAt)
	}
}

func TestBuildReviewArchiveSummaryCombinesTimelineAndEvidenceSuccessesWhenSourcesAreSplit(t *testing.T) {
	t.Parallel()

	success := true
	now := time.Date(2026, 4, 13, 15, 0, 0, 0, time.UTC)

	summary := buildReviewArchiveSummary(
		6,
		&assessmentdomain.PersonalReportStats{
			TotalScore:    150,
			TotalSolved:   1,
			TotalAttempts: 2,
			Rank:          1,
		},
		[]assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:        "flag_submit",
				ChallengeID: 11,
				Title:       "web-flag",
				Timestamp:   now,
				IsCorrect:   &success,
			},
		},
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:      "awd_attack_submission",
				Category:  "web",
				Timestamp: now.Add(time.Minute),
				Meta:      map[string]any{"is_success": true},
			},
		},
		nil,
		nil,
	)

	if summary.CorrectSubmissionCount != 2 {
		t.Fatalf("expected timeline and evidence successes to both count, got %+v", summary)
	}
}

func TestBuildReviewArchiveObservationsTreatsAWDAttacksAsHandsOnEvidence(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 13, 15, 30, 0, 0, time.UTC)
	evidence := []assessmentdomain.ReviewArchiveEvidenceEvent{
		{
			Type:        "awd_attack_submission",
			ChallengeID: 101,
			Category:    "pwn",
			Title:       "awd-pwn",
			Timestamp:   now.Add(-2 * time.Minute),
			Detail:      "AWD 攻击未命中 red-team",
			Meta:        map[string]any{"is_success": false, "event_stage": "exploit"},
		},
		{
			Type:        "awd_attack_submission",
			ChallengeID: 101,
			Category:    "pwn",
			Title:       "awd-pwn",
			Timestamp:   now.Add(-1 * time.Minute),
			Detail:      "AWD 攻击未命中 blue-team",
			Meta:        map[string]any{"is_success": false, "event_stage": "exploit"},
		},
	}

	observations := buildReviewArchiveObservations(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          2,
			CorrectSubmissionCount: 0,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "pwn", Score: 0.3},
		},
		nil,
		evidence,
		nil,
		nil,
	)

	if observation := findObservation(observations.Items, "weak_direction"); observation == nil || observation.Dimension == nil || *observation.Dimension != "pwn" {
		t.Fatalf("expected weak_direction observation from repeated AWD failures, got %+v", observations.Items)
	}
	if observation := findObservation(observations.Items, "awd_summary"); observation == nil || observation.Severity != "warning" {
		t.Fatalf("expected awd_summary warning observation from AWD exploit evidence, got %+v", observations.Items)
	}
}

func TestBuildReviewArchiveObservationsSkipsAWDSummaryWithoutAWDAttempts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)
	observations := buildReviewArchiveObservations(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          1,
			CorrectSubmissionCount: 0,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.34},
		},
		nil,
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:        "instance_access",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   now.Add(-2 * time.Minute),
				Detail:      "进入实例访问页面",
			},
			{
				Type:        "instance_proxy_request",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   now.Add(-time.Minute),
				Detail:      "经平台代理发起 GET /debug",
			},
		},
		nil,
		nil,
	)

	if observation := findObservation(observations.Items, "awd_summary"); observation != nil {
		t.Fatalf("expected no awd_summary without awd attempts, got %+v", observation)
	}
}

func TestBuildReviewArchiveObservationsMarksSingleAWDProbeAsAttention(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 13, 16, 5, 0, 0, time.UTC)
	observations := buildReviewArchiveObservations(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          1,
			CorrectSubmissionCount: 0,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.36},
		},
		nil,
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:        "awd_attack_submission",
				ChallengeID: 101,
				Category:    "web",
				Title:       "awd-web",
				Timestamp:   now,
				Detail:      "AWD 单次试探未命中",
				Meta:        map[string]any{"is_success": false, "event_stage": "probe"},
			},
			{
				Type:        "awd_traffic",
				ChallengeID: 101,
				Category:    "web",
				Title:       "awd-web",
				Timestamp:   now.Add(30 * time.Second),
				Detail:      "GET /status",
				Meta:        map[string]any{"event_stage": "exploit"},
			},
			{
				Type:        "awd_traffic",
				ChallengeID: 101,
				Category:    "web",
				Title:       "awd-web",
				Timestamp:   now.Add(90 * time.Second),
				Detail:      "POST /probe",
				Meta:        map[string]any{"event_stage": "exploit"},
			},
		},
		nil,
		nil,
	)

	if observation := findObservation(observations.Items, "awd_summary"); observation == nil || observation.Severity != "attention" {
		t.Fatalf("expected awd_summary attention for single awd probe, got %+v", observations.Items)
	}
}

func TestBuildReviewArchiveTeachingFactSnapshotOnlyMarksDimensionWithRealEvidence(t *testing.T) {
	t.Parallel()

	snapshot := buildReviewArchiveTeachingFactSnapshot(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          2,
			CorrectSubmissionCount: 0,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.28},
			{Dimension: "pwn", Score: 0.24},
		},
		nil,
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:        "challenge_submission",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   time.Date(2026, 4, 13, 15, 0, 0, 0, time.UTC),
				Detail:      "提交未命中 Flag",
				Meta:        map[string]any{"is_correct": false},
			},
			{
				Type:        "instance_proxy_request",
				ChallengeID: 11,
				Category:    "web",
				Title:       "web-1",
				Timestamp:   time.Date(2026, 4, 13, 15, 2, 0, 0, time.UTC),
				Detail:      "经平台代理发起 POST /login",
				Meta:        map[string]any{"event_stage": "exploit"},
			},
		},
		nil,
		nil,
	)

	evaluation := teachingadvice.EvaluateStudent(snapshot)
	if len(evaluation.WeakDimensions) != 0 {
		t.Fatalf("expected no explicit weak dimensions for sparse archive evidence, got %+v", evaluation.WeakDimensions)
	}
	if len(evaluation.RecommendationTargets) == 0 || evaluation.RecommendationTargets[0].Dimension != "web" {
		t.Fatalf("expected web to remain the primary recommendation target, got %+v", evaluation.RecommendationTargets)
	}
	for _, item := range evaluation.RecommendationTargets {
		if item.Dimension == "pwn" {
			t.Fatalf("expected pwn to stay out of recommendation targets without archive evidence, got %+v", evaluation.RecommendationTargets)
		}
	}
}

func TestBuildReviewArchiveTeachingFactSnapshotUsesExplicitTrackedSubmissionCounts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	snapshot := buildReviewArchiveTeachingFactSnapshot(
		assessmentdomain.ReviewArchiveSummary{
			TotalAttempts:          2,
			CorrectSubmissionCount: 2,
			LastActivityAt:         &now,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.42},
		},
		nil,
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:      "challenge_submission",
				Category:  "web",
				Timestamp: now.Add(-4 * time.Minute),
				Meta:      map[string]any{"is_correct": false},
			},
			{
				Type:      "challenge_submission",
				Category:  "web",
				Timestamp: now.Add(-3 * time.Minute),
				Meta:      map[string]any{"is_correct": true},
			},
			{
				Type:      "awd_attack_submission",
				Category:  "web",
				Timestamp: now.Add(-2 * time.Minute),
				Meta:      map[string]any{"is_success": false},
			},
			{
				Type:      "awd_attack_submission",
				Category:  "web",
				Timestamp: now.Add(-1 * time.Minute),
				Meta:      map[string]any{"is_success": true},
			},
		},
		nil,
		nil,
	)

	if snapshot.CorrectSubmissionCount != 2 {
		t.Fatalf("expected 2 successful archive events, got %+v", snapshot)
	}
	if snapshot.WrongSubmissionCount != 2 {
		t.Fatalf("expected explicit tracked failures to be counted, got %+v", snapshot)
	}
	if snapshot.ChallengeSuccessCount != 1 {
		t.Fatalf("expected 1 challenge success, got %+v", snapshot)
	}
	if snapshot.SubmissionSuccessCount != 2 || snapshot.SubmissionFailureCount != 2 {
		t.Fatalf("expected explicit success/failure breakdown, got %+v", snapshot)
	}
	if snapshot.AWDSuccessCount != 1 {
		t.Fatalf("expected 1 awd success, got %+v", snapshot)
	}
}

func TestRecentReviewArchiveActivityStatsUsesEvidenceAndWriteupsWithinSevenDays(t *testing.T) {
	t.Parallel()

	referenceTime := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	publishedAt := referenceTime.Add(-2 * time.Hour)

	recentEventCount, activeDays := recentReviewArchiveActivityStats(
		referenceTime,
		[]assessmentdomain.ReviewArchiveTimelineEvent{
			{
				Type:      "challenge_submission",
				Timestamp: referenceTime.Add(-24 * time.Hour),
			},
		},
		[]assessmentdomain.ReviewArchiveEvidenceEvent{
			{
				Type:      "instance_proxy_request",
				Timestamp: referenceTime.Add(-3 * time.Hour),
			},
			{
				Type:      "challenge_submission",
				Timestamp: referenceTime.AddDate(0, 0, -10),
			},
		},
		[]assessmentdomain.ReviewArchiveWriteupItem{
			{
				PublishedAt: &publishedAt,
			},
		},
		nil,
	)

	if recentEventCount != 2 {
		t.Fatalf("expected 2 recent events from evidence/writeup within 7 days, got %d", recentEventCount)
	}
	if activeDays != 1 {
		t.Fatalf("expected 1 active day, got %d", activeDays)
	}
}

func TestBuildReviewArchiveTeachingFactSnapshotCountsRecentManualReviewsAsActivity(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	manualReviews := []assessmentdomain.ReviewArchiveManualReviewItem{
		{SubmittedAt: now.Add(-12 * time.Hour), Category: "web"},
		{SubmittedAt: now.Add(-48 * time.Hour), Category: "web"},
		{SubmittedAt: now.Add(-96 * time.Hour), Category: "web"},
	}

	snapshot := buildReviewArchiveTeachingFactSnapshot(
		assessmentdomain.ReviewArchiveSummary{
			LastActivityAt: &now,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.42},
		},
		nil,
		nil,
		nil,
		manualReviews,
	)

	if snapshot.RecentEventCount7d != 3 {
		t.Fatalf("expected recent manual reviews to count as 3 recent events, got %+v", snapshot)
	}
	if snapshot.ActiveDays7d != 3 {
		t.Fatalf("expected recent manual reviews to span 3 active days, got %+v", snapshot)
	}

	observations := buildReviewArchiveObservations(
		assessmentdomain.ReviewArchiveSummary{
			LastActivityAt: &now,
		},
		[]*assessmentcontracts.SkillDimension{
			{Dimension: "web", Score: 0.42},
		},
		nil,
		nil,
		nil,
		manualReviews,
	)
	if observation := findObservation(observations.Items, "activity_status"); observation == nil || observation.Severity != "good" {
		t.Fatalf("expected activity_status good when recent manual reviews keep the student active, got %+v", observation)
	}
}
