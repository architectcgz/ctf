package queries

import (
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ChallengeService struct {
	repo          contestports.ContestChallengeRepository
	challengeRepo challengecontracts.ContestChallengeContract
	contestRepo   contestports.ContestLookupRepository
	awdRepo       contestports.AWDRepository
}

func NewChallengeService(repo contestports.ContestChallengeRepository, challengeRepo challengecontracts.ContestChallengeContract, contestRepo contestports.ContestLookupRepository, awdRepo contestports.AWDRepository) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
		awdRepo:       awdRepo,
	}
}
