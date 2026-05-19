package queries

import (
	contestports "ctf-platform/internal/module/contest/ports"
)

type ChallengeService struct {
	repo          contestports.ContestChallengeReadRepository
	challengeRepo contestports.ContestChallengeCatalog
	contestRepo   contestports.ContestLookupRepository
	awdRepo       contestports.ContestChallengeAWDServiceListRepository
}

func NewChallengeService(repo contestports.ContestChallengeReadRepository, challengeRepo contestports.ContestChallengeCatalog, contestRepo contestports.ContestLookupRepository, awdRepo contestports.ContestChallengeAWDServiceListRepository) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
		awdRepo:       awdRepo,
	}
}
