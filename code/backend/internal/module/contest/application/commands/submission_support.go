package commands

import (
	"errors"
	"fmt"
	"time"

	"ctf-platform/internal/model"
)

var errContestSharedProofInvalid = errors.New("contest shared proof invalid")

type validatedContestSubmission struct {
	contestChallenge *model.ContestChallenge
	teamID           *int64
	rateLimitKey     string
	submittedAt      time.Time
	isCorrect        bool
	sharedProofHash  string
}

func buildContestSubmission(userID, contestID, challengeID int64, flag string, teamID *int64, submittedAt time.Time) *model.Submission {
	return &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		ContestID:   &contestID,
		TeamID:      teamID,
		Flag:        flag,
		IsCorrect:   false,
		Score:       0,
		SubmittedAt: submittedAt,
	}
}

func contestSubmissionRateLimitKey(userID, contestID, challengeID int64) string {
	return fmt.Sprintf("contest:submit:rate:%d:%d:%d", userID, contestID, challengeID)
}
