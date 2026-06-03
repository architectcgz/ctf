package app

import (
	"bytes"
	"strings"
	"testing"

	practicecmd "ctf-platform/internal/module/practice/application/commands"
)

func TestPracticeFlow_PublishedChallengeLifecycleAndAccess(t *testing.T) {
	result := runPublishedPracticeFlowScenario(t)

	if len(result.listBeforeItems) != 1 {
		t.Fatalf("expected 1 published challenge, got %+v", result.listBeforeItems)
	}
	if result.listBeforeItems[0].IsSolved {
		t.Fatalf("expected challenge to be unsolved before submission")
	}

	if bytes.Contains(result.detailBody, []byte(`"is_unlocked"`)) {
		t.Fatalf("expected challenge detail payload to omit is_unlocked, got %s", string(result.detailBody))
	}
	if bytes.Contains(result.detailBody, []byte(`"cost_points"`)) {
		t.Fatalf("expected challenge detail payload to omit cost_points, got %s", string(result.detailBody))
	}
	if result.detail.IsSolved {
		t.Fatalf("expected challenge detail to be unsolved before submission")
	}
	if result.detail.AttachmentURL != "https://example.com/files/web-sqli-101.zip" {
		t.Fatalf("unexpected attachment_url: %s", result.detail.AttachmentURL)
	}
	if !result.detail.NeedTarget {
		t.Fatalf("expected need_target=true, got false")
	}
	if len(result.detail.Hints) != 1 || result.detail.Hints[0].Content == "" {
		t.Fatalf("expected hint content available in challenge detail, got %+v", result.detail.Hints)
	}

	if result.instance.ID <= 0 || result.instance.AccessURL == "" {
		t.Fatalf("expected instance to expose access url, got %+v", result.instance)
	}
	if !strings.Contains(result.proxyAccess.AccessURL, "/api/v1/instances/") || !strings.Contains(result.proxyAccess.AccessURL, "/proxy/") {
		t.Fatalf("expected proxied instance access url, got %+v", result.proxyAccess)
	}
	if result.proxyLocation == "" || strings.Contains(result.proxyLocation, "ticket=") {
		t.Fatalf("expected sanitized proxy redirect location, got %q", result.proxyLocation)
	}
	if len(result.proxyCookies) == 0 {
		t.Fatal("expected proxy bootstrap to issue cookie")
	}
}

func TestPracticeFlow_PublishedChallengeSubmissionsAndProgress(t *testing.T) {
	result := runPublishedPracticeFlowScenario(t)

	if result.wrongSubmission.IsCorrect {
		t.Fatalf("expected wrong flag submission to be incorrect")
	}
	if result.wrongSubmission.Message != "" {
		t.Fatalf("expected wrong submission message to be omitted, got %+v", result.wrongSubmission)
	}

	if !result.correctSubmission.IsCorrect {
		t.Fatalf("expected correct flag submission to succeed")
	}
	if result.correctSubmission.Points != 100 {
		t.Fatalf("expected 100 points, got %d", result.correctSubmission.Points)
	}
	if result.correctSubmission.Message != "" {
		t.Fatalf("expected correct submission message to be omitted, got %+v", result.correctSubmission)
	}

	if len(result.submissionHistory) != 2 {
		t.Fatalf("expected 2 submission history records, got %d", len(result.submissionHistory))
	}
	if result.submissionHistory[0].Status != practicecmd.SubmissionStatusCorrect {
		t.Fatalf("unexpected latest submission record: %+v", result.submissionHistory[0])
	}
	if result.submissionHistory[0].Message != "" {
		t.Fatalf("expected latest submission record message to be omitted, got %+v", result.submissionHistory[0])
	}
	if result.submissionHistory[1].Status != practicecmd.SubmissionStatusIncorrect {
		t.Fatalf("unexpected previous submission record: %+v", result.submissionHistory[1])
	}
	if result.submissionHistory[1].Message != "" {
		t.Fatalf("expected previous submission record message to be omitted, got %+v", result.submissionHistory[1])
	}

	if !result.repeatSubmission.IsCorrect {
		t.Fatalf("expected repeated correct submission to stay correct, got %+v", result.repeatSubmission)
	}
	if result.repeatSubmission.Points != 0 {
		t.Fatalf("expected repeated correct submission not to award points, got %+v", result.repeatSubmission)
	}

	if len(result.listAfterItems) != 1 {
		t.Fatalf("expected 1 challenge after submit, got %+v", result.listAfterItems)
	}
	if !result.listAfterItems[0].IsSolved {
		t.Fatalf("expected challenge to be solved after correct submission")
	}
	if result.listAfterItems[0].SolvedCount != 1 {
		t.Fatalf("expected solved_count 1, got %d", result.listAfterItems[0].SolvedCount)
	}
	if result.listAfterItems[0].TotalAttempts != 2 {
		t.Fatalf("expected total_attempts 2, got %d", result.listAfterItems[0].TotalAttempts)
	}

	if result.progress.TotalSolved != 1 {
		t.Fatalf("expected total_solved 1, got %d", result.progress.TotalSolved)
	}
	if result.progress.TotalScore != 100 {
		t.Fatalf("expected total_score 100, got %d", result.progress.TotalScore)
	}
	if result.progress.Rank != 1 {
		t.Fatalf("expected rank 1, got %d", result.progress.Rank)
	}

	if len(result.submissions) != 2 {
		t.Fatalf("expected 2 submission records, got %d", len(result.submissions))
	}
	if result.submissions[0].IsCorrect {
		t.Fatalf("expected first submission to be incorrect")
	}
	if !result.submissions[1].IsCorrect {
		t.Fatalf("expected second submission to be correct")
	}
}
