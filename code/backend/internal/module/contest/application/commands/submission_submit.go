package commands

import (
	"context"

	"ctf-platform/internal/dto"
)

func (s *SubmissionService) SubmitFlagInContest(ctx context.Context, userID, contestID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	attempt, err := s.validateContestSubmission(ctx, userID, contestID, challengeID, flag)
	if err != nil {
		return nil, err
	}

	submission := buildContestSubmission(userID, contestID, challengeID, flag, attempt.teamID, attempt.submittedAt)
	if !attempt.isCorrect {
		return s.handleIncorrectSubmission(ctx, submission)
	}

	finalScore, err := s.handleCorrectSubmission(ctx, submission, attempt.contestChallenge, attempt.teamID)
	if err != nil {
		return nil, err
	}

	return contestResponseMapperInst.ToSubmissionRespPtr(submissionRespSource{
		IsCorrect:   true,
		Points:      finalScore,
		SubmittedAt: submission.SubmittedAt,
	}), nil
}
