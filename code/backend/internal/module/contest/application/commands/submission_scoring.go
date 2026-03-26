package commands

import (
	"context"
	"math"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) handleCorrectSubmission(ctx context.Context, submission *model.Submission, contestChallenge *model.ContestChallenge, teamID *int64) (int, error) {
	challengeRecord, err := s.repo.FindChallengeByID(ctx, submission.ChallengeID)
	if err != nil {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	finalScore := 0
	teamScoreDeltas := make(map[int64]int)

	err = s.repo.WithinTransaction(ctx, func(txRepo contestports.ContestSubmissionRepository) error {
		lockedChallenge, err := txRepo.LockContestChallenge(ctx, *submission.ContestID, submission.ChallengeID)
		if err != nil {
			return err
		}

		count, err := txRepo.CountCorrectSubmissions(ctx, *submission.ContestID, submission.ChallengeID, teamID, submission.UserID)
		if err != nil {
			return err
		}
		if count > 0 {
			return errcode.ErrContestChallengeSolved
		}

		if teamID != nil && lockedChallenge.FirstBloodBy == nil {
			if err := txRepo.UpdateFirstBlood(ctx, *submission.ContestID, submission.ChallengeID, *teamID); err != nil {
				return err
			}
			lockedChallenge.FirstBloodBy = teamID
		}

		submission.IsCorrect = true
		submission.Score = 0
		if err := txRepo.CreateSubmission(ctx, submission); err != nil {
			if isContestSubmissionUniqueViolation(err) {
				return errcode.ErrContestChallengeSolved
			}
			return err
		}

		solvedSubmissions, err := txRepo.ListCorrectSubmissions(ctx, *submission.ContestID, submission.ChallengeID)
		if err != nil {
			return err
		}

		recalculatedScore := s.calculateContestScore(*lockedChallenge, *challengeRecord, int64(len(solvedSubmissions)))
		firstBloodBonus := int(math.Round(float64(recalculatedScore) * s.cfg.Contest.FirstBloodBonus))
		scoreUpdates, currentScore := buildContestSubmissionScoreUpdates(solvedSubmissions, lockedChallenge.FirstBloodBy, recalculatedScore, firstBloodBonus, submission.ID)
		for _, update := range scoreUpdates {
			if update.NewScore == update.OldScore {
				continue
			}
			if err := txRepo.UpdateSubmissionScore(ctx, update.SubmissionID, update.NewScore); err != nil {
				return err
			}
			if update.TeamID != nil {
				teamScoreDeltas[*update.TeamID] += update.NewScore - update.OldScore
			}
		}

		for affectedTeamID, delta := range teamScoreDeltas {
			if delta == 0 {
				continue
			}
			var lastSolveAt *time.Time
			if teamID != nil && affectedTeamID == *teamID {
				lastSolveAt = &submission.SubmittedAt
			}
			if err := txRepo.AddTeamScore(ctx, affectedTeamID, delta, lastSolveAt); err != nil {
				return err
			}
		}

		finalScore = currentScore
		return nil
	})
	if err != nil {
		return 0, mapSubmissionError(err)
	}

	if submission.ContestID != nil && s.scoreboardService != nil {
		for affectedTeamID, delta := range teamScoreDeltas {
			if delta == 0 {
				continue
			}
			if err := s.scoreboardService.UpdateScore(ctx, *submission.ContestID, affectedTeamID, float64(delta)); err != nil {
				if rebuildErr := s.scoreboardService.RebuildScoreboard(ctx, *submission.ContestID); rebuildErr != nil {
					return 0, rebuildErr
				}
				break
			}
		}
	}
	return finalScore, nil
}
