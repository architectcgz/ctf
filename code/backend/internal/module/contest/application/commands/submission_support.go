package commands

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) calculateContestScore(contestChallenge model.ContestChallenge, challengeRecord model.Challenge, solveCount int64) int {
	baseScore := s.resolveContestBaseScore(contestChallenge, challengeRecord)
	if baseScore <= 0 {
		baseScore = s.cfg.Contest.BaseScore
	}
	return contestdomain.CalculateDynamicScore(baseScore, s.cfg.Contest.MinScore, s.cfg.Contest.Decay, solveCount)
}

func (s *SubmissionService) resolveContestBaseScore(contestChallenge model.ContestChallenge, challengeRecord model.Challenge) float64 {
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

func mapSubmissionError(err error) error {
	if appErr, ok := err.(*errcode.AppError); ok {
		return appErr
	}
	return errcode.ErrInternal.WithCause(err)
}

func isContestSubmissionUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	return pgErr.ConstraintName == "uk_submissions_contest_team_challenge_correct" ||
		pgErr.ConstraintName == "uk_submissions_contest_user_challenge_correct"
}

type contestSubmissionScoreUpdate struct {
	SubmissionID int64
	TeamID       *int64
	OldScore     int
	NewScore     int
}

func buildContestSubmissionScoreUpdates(submissions []model.Submission, firstBloodBy *int64, recalculatedScore, firstBloodBonus int, currentSubmissionID int64) ([]contestSubmissionScoreUpdate, int) {
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
