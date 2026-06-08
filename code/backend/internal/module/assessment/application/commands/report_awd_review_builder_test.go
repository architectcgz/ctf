package commands

import (
	"context"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"strings"
	"testing"
)

func TestTeacherAWDReviewExportBuilderSelectsFocusRoundWhenRoundMissing(t *testing.T) {
	t.Parallel()

	reader := &testAWDReviewArchiveReader{
		archives: []*assessmentqry.TeacherAWDReviewArchiveResp{
			{
				Contest: assessmentqry.TeacherAWDReviewContestMetaResp{
					ID:     21,
					Title:  "awd-review",
					Status: contestcontracts.ContestStatusEnded,
				},
				Rounds: []assessmentqry.TeacherAWDReviewRoundResp{
					{RoundNumber: 1, ServiceCount: 2, AttackCount: 0, TrafficCount: 1},
					{RoundNumber: 2, ServiceCount: 2, AttackCount: 3, TrafficCount: 2},
				},
			},
			{
				Contest: assessmentqry.TeacherAWDReviewContestMetaResp{
					ID:     21,
					Title:  "awd-review",
					Status: contestcontracts.ContestStatusEnded,
				},
				Rounds: []assessmentqry.TeacherAWDReviewRoundResp{
					{RoundNumber: 1, ServiceCount: 2, AttackCount: 0, TrafficCount: 1},
					{RoundNumber: 2, ServiceCount: 2, AttackCount: 3, TrafficCount: 2},
				},
				SelectedRound: &assessmentqry.TeacherAWDSelectedRoundResp{
					Round: assessmentqry.TeacherAWDReviewRoundResp{
						RoundNumber:  2,
						ServiceCount: 2,
						AttackCount:  3,
						TrafficCount: 2,
					},
				},
			},
		},
	}

	builder := NewAWDReviewExportBuilder(reader)
	archive, err := builder.BuildArchive(context.Background(), 11, 21, nil)
	if err != nil {
		t.Fatalf("BuildArchive() error = %v", err)
	}
	if archive == nil || archive.SelectedRound == nil || archive.SelectedRound.Round.RoundNumber != 2 {
		t.Fatalf("expected selected round 2, got %+v", archive)
	}
	if len(reader.inputs) != 2 {
		t.Fatalf("expected 2 archive reads, got %d", len(reader.inputs))
	}
	if reader.inputs[0].RoundNumber != nil {
		t.Fatalf("expected first archive read without round filter, got %+v", reader.inputs[0])
	}
	if reader.inputs[1].RoundNumber == nil || *reader.inputs[1].RoundNumber != 2 {
		t.Fatalf("expected second archive read with round 2, got %+v", reader.inputs[1])
	}
}

func TestHottestRoundPrefersAttackDenseRound(t *testing.T) {
	t.Parallel()

	round := hottestRound([]assessmentqry.TeacherAWDReviewRoundResp{
		{RoundNumber: 1, ServiceCount: 2, AttackCount: 0, TrafficCount: 4},
		{RoundNumber: 2, ServiceCount: 1, AttackCount: 2, TrafficCount: 1},
		{RoundNumber: 3, ServiceCount: 5, AttackCount: 0, TrafficCount: 0},
	})
	if round == nil || round.RoundNumber != 2 {
		t.Fatalf("expected hottest round 2, got %+v", round)
	}
}

func TestTopRiskyServicePrefersCompromisedService(t *testing.T) {
	t.Parallel()

	service := topRiskyService([]assessmentqry.TeacherAWDReviewServiceResp{
		{TeamName: "blue", AWDChallengeTitle: "web", ServiceStatus: contestcontracts.AWDServiceStatusUp, AttackReceived: 4},
		{TeamName: "red", AWDChallengeTitle: "api", ServiceStatus: contestcontracts.AWDServiceStatusCompromised, AttackReceived: 1},
	})
	if service == nil || service.TeamName != "red" {
		t.Fatalf("expected compromised red service to be top risk, got %+v", service)
	}
}

func TestBuildAWDReviewSuggestionsIncludesTrafficOnlyHint(t *testing.T) {
	t.Parallel()

	suggestions := buildAWDReviewSuggestions(
		[]assessmentqry.TeacherAWDReviewRoundResp{
			{RoundNumber: 4, AttackCount: 0, TrafficCount: 3, ServiceCount: 1},
		},
		&assessmentqry.TeacherAWDSelectedRoundResp{
			Round: assessmentqry.TeacherAWDReviewRoundResp{RoundNumber: 4},
			Services: []assessmentqry.TeacherAWDReviewServiceResp{
				{TeamName: "blue", AWDChallengeTitle: "web", ServiceStatus: contestcontracts.AWDServiceStatusUp, AttackReceived: 2},
			},
			Traffic: []assessmentqry.TeacherAWDReviewTrafficResp{
				{Path: "/health"},
			},
		},
	)

	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions, got empty")
	}
	joined := strings.Join(suggestions, "\n")
	if !strings.Contains(joined, "访问流量") {
		t.Fatalf("expected traffic-only hint, got %s", joined)
	}
	if !strings.Contains(joined, "第 4 轮") {
		t.Fatalf("expected key round hint, got %s", joined)
	}
}
