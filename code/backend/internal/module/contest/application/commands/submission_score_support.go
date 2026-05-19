package commands

import (
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func (s *SubmissionService) calculateContestScore(contestChallenge contestentity.ContestChallenge, challengeRecord contestentity.Challenge, solveCount int64) int {
	baseScore := s.resolveContestBaseScore(contestChallenge, challengeRecord)
	if baseScore <= 0 {
		baseScore = s.cfg.Contest.BaseScore
	}
	return contestdomain.CalculateDynamicScore(baseScore, s.cfg.Contest.MinScore, s.cfg.Contest.Decay, solveCount)
}

func (s *SubmissionService) resolveContestBaseScore(contestChallenge contestentity.ContestChallenge, challengeRecord contestentity.Challenge) float64 {
	switch {
	case contestChallenge.ContestScore != nil && *contestChallenge.ContestScore > 0:
		return float64(*contestChallenge.ContestScore)
	case contestChallenge.Points > 0:
		return float64(contestChallenge.Points)
	case challengeRecord.Points > 0:
		return float64(challengeRecord.Points)
	default:
		return s.cfg.Contest.BaseScore
	}
}

type contestSubmissionScoreUpdate struct {
	SubmissionID int64
	TeamID       *int64
	OldScore     int
	NewScore     int
}

func buildContestSubmissionScoreUpdates(submissions []contestentity.Submission, firstBloodBy *int64, recalculatedScore, firstBloodBonus int, currentSubmissionID int64) ([]contestSubmissionScoreUpdate, int) {
	firstBloodSubmissionID := int64(0)
	if firstBloodBy != nil {
		for _, solvedSubmission := range submissions {
			if solvedSubmission.TeamID != nil && *solvedSubmission.TeamID == *firstBloodBy {
				firstBloodSubmissionID = solvedSubmission.ID
				break
			}
		}
	}

	updates := make([]contestSubmissionScoreUpdate, 0, len(submissions))
	currentScore := 0
	for _, solvedSubmission := range submissions {
		newScore := recalculatedScore
		if firstBloodSubmissionID > 0 && solvedSubmission.ID == firstBloodSubmissionID {
			newScore += firstBloodBonus
		}
		updates = append(updates, contestSubmissionScoreUpdate{
			SubmissionID: solvedSubmission.ID,
			TeamID:       solvedSubmission.TeamID,
			OldScore:     solvedSubmission.Score,
			NewScore:     newScore,
		})
		if solvedSubmission.ID == currentSubmissionID {
			currentScore = newScore
		}
	}
	return updates, currentScore
}
