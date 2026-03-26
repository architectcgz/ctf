package commands

import (
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
)

type AWDService struct {
	repo         contestports.AWDRepository
	roundManager contestports.AWDRoundManager
	redis        *redislib.Client
	contestRepo  contestports.ContestLookupRepository
	flagSecret   string
	awdConfig    config.ContestAWDConfig
	log          *zap.Logger
}

func NewAWDService(
	repo contestports.AWDRepository,
	contestRepo contestports.ContestLookupRepository,
	redis *redislib.Client,
	flagSecret string,
	awdConfig config.ContestAWDConfig,
	log *zap.Logger,
	roundManager contestports.AWDRoundManager,
) *AWDService {
	if log == nil {
		log = zap.NewNop()
	}
	return &AWDService{
		repo:         repo,
		roundManager: roundManager,
		redis:        redis,
		contestRepo:  contestRepo,
		flagSecret:   flagSecret,
		awdConfig:    awdConfig,
		log:          log,
	}
}
