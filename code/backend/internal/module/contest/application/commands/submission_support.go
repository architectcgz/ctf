package commands

import (
	"time"

	"ctf-platform/internal/model"
)

type validatedContestSubmission struct {
	contestChallenge *model.ContestChallenge
	teamID           *int64
	submittedAt      time.Time
	isCorrect        bool
}

func buildContestSubmission(userID, contestID, challengeID int64, flag string, teamID *int64, submittedAt time.Time) *model.Submission {
	return &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		ContestID:   &contestID,
		TeamID:      teamID,
		Flag:        "",
		IsCorrect:   false,
		Score:       0,
		SubmittedAt: submittedAt,
	}
}
