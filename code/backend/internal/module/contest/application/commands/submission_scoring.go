package commands

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) handleCorrectSubmission(ctx context.Context, submission *contestentity.Submission, contestChallenge *contestentity.ContestChallenge, teamID *int64) (int, error) {
	challengeRecord, err := s.repo.FindChallengeByID(ctx, submission.ChallengeID)
	if err != nil {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	scoringResult, err := s.applyCorrectSubmissionScoring(ctx, submission, challengeRecord, teamID)
	if err != nil {
		return 0, mapSubmissionError(err)
	}

	if err := s.syncCorrectSubmissionScoreboard(ctx, submission.ContestID, scoringResult.teamScoreDeltas); err != nil {
		return 0, err
	}

	return scoringResult.finalScore, nil
}
