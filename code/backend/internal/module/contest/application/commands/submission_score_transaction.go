package commands

import (
	"context"
	"math"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type correctSubmissionScoringResult struct {
	finalScore      int
	teamScoreDeltas map[int64]int
}

func (s *SubmissionService) applyCorrectSubmissionScoring(ctx context.Context, submission *model.Submission, challengeRecord *model.Challenge, teamID *int64) (correctSubmissionScoringResult, error) {
	result := correctSubmissionScoringResult{
		teamScoreDeltas: make(map[int64]int),
	}

	err := s.repo.WithinTransaction(ctx, func(txRepo contestports.ContestSubmissionRepository) error {
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
				result.teamScoreDeltas[*update.TeamID] += update.NewScore - update.OldScore
			}
		}

		for affectedTeamID, delta := range result.teamScoreDeltas {
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

		result.finalScore = currentScore
		return nil
	})

	return result, err
}
