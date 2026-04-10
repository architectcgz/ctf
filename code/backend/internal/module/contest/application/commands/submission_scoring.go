package commands

import (
	"context"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) handleCorrectSubmission(ctx context.Context, submission *model.Submission, contestChallenge *model.ContestChallenge, teamID *int64, sharedProofHash string) (int, error) {
	challengeRecord, err := s.repo.FindChallengeByID(ctx, submission.ChallengeID)
	if err != nil {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	scoringResult, err := s.applyCorrectSubmissionScoring(ctx, submission, challengeRecord, teamID, sharedProofHash)
	if err != nil {
		return 0, mapSubmissionError(err)
	}

	if err := s.syncCorrectSubmissionScoreboard(ctx, submission.ContestID, scoringResult.teamScoreDeltas); err != nil {
		return 0, err
	}

	return scoringResult.finalScore, nil
}
