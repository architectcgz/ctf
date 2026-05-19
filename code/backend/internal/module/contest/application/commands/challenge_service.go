package commands

import (
	contestports "ctf-platform/internal/module/contest/ports"
)

type ChallengeService struct {
	repo          contestChallengeCommandRepository
	challengeRepo contestports.ContestChallengeCatalog
	contestRepo   contestports.ContestLookupRepository
	awdRepo       contestports.ContestChallengeAWDServiceListRepository
}

type contestChallengeCommandRepository interface {
	contestports.ContestChallengeWriteRepository
}

func NewChallengeService(
	repo contestChallengeCommandRepository,
	challengeRepo contestports.ContestChallengeCatalog,
	contestRepo contestports.ContestLookupRepository,
	awdRepo contestports.ContestChallengeAWDServiceListRepository,
) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
		awdRepo:       awdRepo,
	}
}
