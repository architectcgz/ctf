package advice

import (
	"strings"
	"testing"
)

func TestEvaluateStudentOnlyMarksWeakDimensionsWhenEvidenceIsSufficient(t *testing.T) {
	t.Parallel()

	evaluation := EvaluateStudent(StudentFactSnapshot{
		UserID:                 7,
		MaxWrongStreak:         3,
		RecentEventCount7d:     4,
		CorrectSubmissionCount: 1,
		Dimensions: []DimensionFact{
			{Dimension: "web", ProfileScore: 0.22, AttemptCount: 4, SuccessCount: 0, EvidenceCount: 4},
			{Dimension: "crypto", ProfileScore: 0.18, AttemptCount: 1, SuccessCount: 0, EvidenceCount: 1},
		},
	})

	if len(evaluation.WeakDimensions) != 1 {
		t.Fatalf("expected exactly one weak dimension, got %+v", evaluation.WeakDimensions)
	}
	if evaluation.WeakDimensions[0].Dimension != "web" {
		t.Fatalf("expected web to be the weak dimension, got %+v", evaluation.WeakDimensions[0])
	}
	if evaluation.RecommendedDifficultyBand != DifficultyBandBeginner {
		t.Fatalf("expected beginner difficulty band, got %s", evaluation.RecommendedDifficultyBand)
	}
	if evaluation.WeakDimensions[0].Severity != SeverityDanger {
		t.Fatalf("expected web severity danger, got %s", evaluation.WeakDimensions[0].Severity)
	}
	if evaluation.Dimensions[1].IsWeak {
		t.Fatalf("expected low-evidence crypto dimension not to be treated as explicit weak dimension")
	}
}

func TestBuildReviewArchiveObservationsKeepsProcessSignalsAndDimensionFocusSeparate(t *testing.T) {
	t.Parallel()

	snapshot := StudentFactSnapshot{
		UserID:                 11,
		ActiveDays7d:           4,
		RecentEventCount7d:     6,
		CorrectSubmissionCount: 2,
		WrongSubmissionCount:   5,
		MaxWrongStreak:         4,
		WriteupCount:           0,
		ApprovedReviewCount:    0,
		HandsOnEventCount:      3,
		Dimensions: []DimensionFact{
			{Dimension: "pwn", ProfileScore: 0.28, AttemptCount: 5, SuccessCount: 0, EvidenceCount: 5},
		},
	}
	evaluation := EvaluateStudent(snapshot)
	items := BuildReviewArchiveObservations(snapshot, evaluation)

	if len(items) < 4 {
		t.Fatalf("expected multiple review observations, got %+v", items)
	}

	var hasClosure, hasStability, hasHandsOn, hasDimension bool
	for _, item := range items {
		switch item.Code {
		case "training_closure":
			hasClosure = true
			if item.Severity != SeverityAttention {
				t.Fatalf("expected closure severity attention, got %+v", item)
			}
		case "submission_stability":
			hasStability = true
			if item.Severity != SeverityDanger {
				t.Fatalf("expected submission stability severity danger, got %+v", item)
			}
		case "hands_on_depth":
			hasHandsOn = true
		case "dimension_focus":
			hasDimension = true
			if item.Dimension == nil || *item.Dimension != "pwn" {
				t.Fatalf("expected dimension focus to point at pwn, got %+v", item)
			}
		}
	}

	if !hasClosure || !hasStability || !hasHandsOn || !hasDimension {
		t.Fatalf("expected closure/stability/hands-on/dimension observations, got %+v", items)
	}
}

func TestBuildReviewArchiveObservationsAddsLowActivitySignal(t *testing.T) {
	t.Parallel()

	snapshot := StudentFactSnapshot{
		UserID:                 13,
		ActiveDays7d:           1,
		RecentEventCount7d:     1,
		CorrectSubmissionCount: 0,
	}

	items := BuildReviewArchiveObservations(snapshot, EvaluateStudent(snapshot))
	for _, item := range items {
		if item.Code != "low_activity" {
			continue
		}
		if item.Severity != SeverityWarning {
			t.Fatalf("expected low_activity warning, got %+v", item)
		}
		return
	}
	t.Fatalf("expected low_activity observation, got %+v", items)
}

func TestBuildReviewArchiveObservationsDoesNotWarnForSingleWrongSubmission(t *testing.T) {
	t.Parallel()

	snapshot := StudentFactSnapshot{
		UserID:               14,
		ActiveDays7d:         3,
		RecentEventCount7d:   4,
		WrongSubmissionCount: 1,
		MaxWrongStreak:       1,
	}

	items := BuildReviewArchiveObservations(snapshot, EvaluateStudent(snapshot))
	for _, item := range items {
		if item.Code == "submission_stability" {
			t.Fatalf("expected no submission stability warning for single wrong submission, got %+v", item)
		}
	}
}

func TestEvaluateStudentTreatsStableRecentSuccessAsFoundationInsteadOfExplicitWeakness(t *testing.T) {
	t.Parallel()

	evaluation := EvaluateStudent(StudentFactSnapshot{
		UserID: 12,
		Dimensions: []DimensionFact{
			{Dimension: "pwn", ProfileScore: 0.26, AttemptCount: 2, SuccessCount: 2, EvidenceCount: 4},
		},
	})

	if len(evaluation.WeakDimensions) != 0 {
		t.Fatalf("expected no explicit weak dimensions for stable recent success, got %+v", evaluation.WeakDimensions)
	}
	if len(evaluation.RecommendationTargets) != 1 {
		t.Fatalf("expected one recommendation target, got %+v", evaluation.RecommendationTargets)
	}
	if evaluation.RecommendationTargets[0].Severity != SeverityAttention {
		t.Fatalf("expected attention severity for stable coverage gap, got %+v", evaluation.RecommendationTargets[0])
	}
	if evaluation.RecommendationTargets[0].Summary == "" {
		t.Fatalf("expected structured summary for stable coverage gap")
	}
}

func TestEvaluateStudentDoesNotRecommendHealthyDimensionWithOnlyGoodEvidence(t *testing.T) {
	t.Parallel()

	evaluation := EvaluateStudent(StudentFactSnapshot{
		UserID: 19,
		Dimensions: []DimensionFact{
			{Dimension: "web", ProfileScore: 0.82, AttemptCount: 3, SuccessCount: 2, EvidenceCount: 8},
		},
	})

	if len(evaluation.WeakDimensions) != 0 {
		t.Fatalf("expected no weak dimensions for healthy evidence-backed student, got %+v", evaluation.WeakDimensions)
	}
	if len(evaluation.RecommendationTargets) != 0 {
		t.Fatalf("expected no recommendation targets for healthy evidence-backed student, got %+v", evaluation.RecommendationTargets)
	}

	items := BuildReviewArchiveObservations(
		StudentFactSnapshot{
			UserID:                 19,
			ActiveDays7d:           4,
			RecentEventCount7d:     8,
			CorrectSubmissionCount: 2,
			WriteupCount:           1,
			HandsOnEventCount:      5,
			Dimensions: []DimensionFact{
				{Dimension: "web", ProfileScore: 0.82, AttemptCount: 3, SuccessCount: 2, EvidenceCount: 8},
			},
		},
		evaluation,
	)
	for _, item := range items {
		if item.Code == "dimension_focus" {
			t.Fatalf("expected no dimension_focus observation for healthy evidence-backed student, got %+v", item)
		}
	}
}

func TestBuildClassReviewAggregatesWeakDimensionAndRiskSignals(t *testing.T) {
	t.Parallel()

	students := []StudentFactSnapshot{
		{
			UserID:                 1,
			Username:               "alice",
			ActiveDays7d:           1,
			RecentEventCount7d:     1,
			CorrectSubmissionCount: 1,
			MaxWrongStreak:         4,
			Dimensions: []DimensionFact{
				{Dimension: "web", ProfileScore: 0.2, AttemptCount: 4, SuccessCount: 0, EvidenceCount: 4},
			},
		},
		{
			UserID:                 2,
			Username:               "bob",
			ActiveDays7d:           2,
			RecentEventCount7d:     3,
			CorrectSubmissionCount: 2,
			WriteupCount:           0,
			Dimensions: []DimensionFact{
				{Dimension: "web", ProfileScore: 0.3, AttemptCount: 3, SuccessCount: 1, EvidenceCount: 3},
			},
		},
		{
			UserID:                 3,
			Username:               "carol",
			ActiveDays7d:           5,
			RecentEventCount7d:     8,
			CorrectSubmissionCount: 3,
			WriteupCount:           2,
			ApprovedReviewCount:    1,
			Dimensions: []DimensionFact{
				{Dimension: "crypto", ProfileScore: 0.82, AttemptCount: 3, SuccessCount: 2, EvidenceCount: 3},
			},
		},
	}

	evaluations := make(map[int64]StudentEvaluation, len(students))
	for _, student := range students {
		evaluations[student.UserID] = EvaluateStudent(student)
	}

	items := BuildClassReview(
		"信安2401",
		ClassSummarySnapshot{ClassName: "信安2401", StudentCount: 3, ActiveRate: 55, RecentEventCount: 12},
		&ClassTrendSnapshot{EventDelta: -3, SolveDelta: -1},
		students,
		evaluations,
	)

	if len(items) < 4 {
		t.Fatalf("expected multiple class review items, got %+v", items)
	}

	var hasActivity, hasWeakCluster, hasClosureGap, hasRetryCost, hasTrend bool
	for _, item := range items {
		switch item.Code {
		case "activity_risk":
			hasActivity = true
			if item.Severity != SeverityWarning {
				t.Fatalf("expected activity risk warning, got %+v", item)
			}
		case "weak_dimension_cluster":
			hasWeakCluster = true
			if item.Dimension != "web" {
				t.Fatalf("expected weak dimension cluster web, got %+v", item)
			}
		case "training_closure_gap":
			hasClosureGap = true
		case "retry_cost_high":
			hasRetryCost = true
		case "trend_watch":
			hasTrend = true
		}
	}

	if !hasActivity || !hasWeakCluster || !hasClosureGap || !hasRetryCost || !hasTrend {
		t.Fatalf("expected activity/weak cluster/closure/retry/trend items, got %+v", items)
	}
}

func TestBuildClassReviewKeepsHealthyClassAsAttentionWhenOnlyFewStudentsSlowDown(t *testing.T) {
	t.Parallel()

	students := []StudentFactSnapshot{
		{
			UserID:             1,
			Username:           "alice",
			ActiveDays7d:       1,
			RecentEventCount7d: 1,
		},
		{
			UserID:                 2,
			Username:               "bob",
			ActiveDays7d:           4,
			RecentEventCount7d:     6,
			CorrectSubmissionCount: 1,
		},
		{
			UserID:                 3,
			Username:               "carol",
			ActiveDays7d:           5,
			RecentEventCount7d:     8,
			CorrectSubmissionCount: 2,
		},
		{
			UserID:                 4,
			Username:               "dave",
			ActiveDays7d:           4,
			RecentEventCount7d:     5,
			CorrectSubmissionCount: 1,
		},
		{
			UserID:                 5,
			Username:               "eve",
			ActiveDays7d:           3,
			RecentEventCount7d:     4,
			CorrectSubmissionCount: 1,
		},
		{
			UserID:                 6,
			Username:               "frank",
			ActiveDays7d:           5,
			RecentEventCount7d:     7,
			CorrectSubmissionCount: 2,
		},
	}

	evaluations := make(map[int64]StudentEvaluation, len(students))
	for _, student := range students {
		evaluations[student.UserID] = EvaluateStudent(student)
	}

	items := BuildClassReview(
		"信安2401",
		ClassSummarySnapshot{ClassName: "信安2401", StudentCount: len(students), ActiveRate: 96, RecentEventCount: 60},
		&ClassTrendSnapshot{EventDelta: 3, SolveDelta: 1},
		students,
		evaluations,
	)

	for _, item := range items {
		if item.Code != "activity_risk" {
			continue
		}
		if item.Severity != SeverityAttention {
			t.Fatalf("expected activity risk attention for mostly healthy class, got %+v", item)
		}
		return
	}
	t.Fatalf("expected activity_risk item, got %+v", items)
}

func TestBuildRecommendationPlanExplainsCandidateDifficultyFallbackInReason(t *testing.T) {
	t.Parallel()

	snapshot := StudentFactSnapshot{
		UserID: 18,
		Dimensions: []DimensionFact{
			{Dimension: "reverse", ProfileScore: 0.34, AttemptCount: 3, SuccessCount: 0, EvidenceCount: 3},
		},
	}
	evaluation := EvaluateStudent(snapshot)
	plan := BuildRecommendationPlan(snapshot, evaluation, []ChallengeCandidate{
		{ID: 51, Title: "rev-lab", Category: "reverse", Difficulty: "easy", Points: 120},
	})

	if len(plan.Reasons) != 1 {
		t.Fatalf("expected one recommendation reason, got %+v", plan.Reasons)
	}
	if plan.Reasons[0].Dimension != "reverse" {
		t.Fatalf("expected reverse recommendation reason, got %+v", plan.Reasons[0])
	}
	if plan.Reasons[0].Evidence == "" || plan.Reasons[0].Summary == "" {
		t.Fatalf("expected structured recommendation reason text, got %+v", plan.Reasons[0])
	}
	if !strings.Contains(plan.Reasons[0].Summary, "easy 难度题") {
		t.Fatalf("expected summary to mention actual challenge difficulty, got %+v", plan.Reasons[0])
	}
	if !strings.Contains(plan.Reasons[0].Summary, "beginner 难度") {
		t.Fatalf("expected summary to retain preferred training band, got %+v", plan.Reasons[0])
	}
}
