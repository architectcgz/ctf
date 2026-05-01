package commands

import (
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"

	redislib "github.com/redis/go-redis/v9"
)

type ChallengeService struct {
	repo          contestports.ContestChallengeRepository
	challengeRepo challengecontracts.ContestChallengeContract
	contestRepo   contestports.ContestLookupRepository
	awdRepo       contestports.AWDServiceStore
	redis         *redislib.Client
}

func NewChallengeService(
	repo contestports.ContestChallengeRepository,
	challengeRepo challengecontracts.ContestChallengeContract,
	contestRepo contestports.ContestLookupRepository,
	awdRepo contestports.AWDServiceStore,
	redis *redislib.Client,
) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
		awdRepo:       awdRepo,
		redis:         redis,
	}
}
