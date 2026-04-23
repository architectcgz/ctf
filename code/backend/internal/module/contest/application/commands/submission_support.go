package commands

import (
	"fmt"
	"strings"
	"time"

	"ctf-platform/internal/model"
)

const defaultContestSubmissionRateLimitPrefix = "ctf:ratelimit"

type validatedContestSubmission struct {
	contestChallenge *model.ContestChallenge
	teamID           *int64
	rateLimitKey     string
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

func contestSubmissionRateLimitKey(prefix string, userID, contestID, challengeID int64) string {
	trimmedPrefix := strings.TrimSpace(prefix)
	if trimmedPrefix == "" {
		trimmedPrefix = defaultContestSubmissionRateLimitPrefix
	}
	return fmt.Sprintf("%s:contest:submit:rate:%d:%d:%d", trimmedPrefix, userID, contestID, challengeID)
}
