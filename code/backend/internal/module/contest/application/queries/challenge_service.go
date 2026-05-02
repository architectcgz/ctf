package queries

import (
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ChallengeService struct {
	repo          contestports.ContestChallengeReadRepository
	challengeRepo challengecontracts.ContestChallengeContract
	contestRepo   contestports.ContestLookupRepository
	awdRepo       contestports.ContestChallengeAWDServiceListRepository
}

func NewChallengeService(repo contestports.ContestChallengeReadRepository, challengeRepo challengecontracts.ContestChallengeContract, contestRepo contestports.ContestLookupRepository, awdRepo contestports.ContestChallengeAWDServiceListRepository) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
		awdRepo:       awdRepo,
	}
}
