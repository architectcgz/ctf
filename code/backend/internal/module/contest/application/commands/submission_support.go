package commands

import (
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

type validatedContestSubmission struct {
	contestChallenge *contestentity.ContestChallenge
	teamID           *int64
	submittedAt      time.Time
	isCorrect        bool
}

func buildContestSubmission(userID, contestID, challengeID int64, flag string, teamID *int64, submittedAt time.Time) *contestentity.Submission {
	return &contestentity.Submission{
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
