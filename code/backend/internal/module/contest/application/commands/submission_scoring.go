package commands

import (
	"context"

	"ctf-platform/internal/apperror"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	platformevents "ctf-platform/internal/platform/events"
)

func (s *SubmissionService) handleCorrectSubmission(ctx context.Context, submission *contestentity.Submission, contestChallenge *contestentity.ContestChallenge, teamID *int64) (int, error) {
	challengeRecord, err := s.repo.FindChallengeByID(ctx, submission.ChallengeID)
	if err != nil {
		return 0, apperror.ErrInternal.WithCause(err)
	}

	scoringResult, err := s.applyCorrectSubmissionScoring(ctx, submission, challengeRecord, teamID)
	if err != nil {
		return 0, mapSubmissionError(err)
	}

	if err := s.syncCorrectSubmissionScoreboard(ctx, submission.ContestID, scoringResult.teamScoreDeltas); err != nil {
		return 0, err
	}

	if submission.ContestID != nil {
		s.publishWeakEvent(ctx, platformevents.Event{
			Name: contestcontracts.EventFlagAccepted,
			Payload: contestcontracts.FlagAcceptedEvent{
				UserID:      submission.UserID,
				ContestID:   *submission.ContestID,
				ChallengeID: submission.ChallengeID,
				Dimension:   challengeRecord.Category,
				OccurredAt:  submission.SubmittedAt,
			},
		})
	}

	return scoringResult.finalScore, nil
}
