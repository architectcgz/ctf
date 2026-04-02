package commands

import (
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ScoreboardAdminService struct {
	repo        contestports.ContestScoreboardAdminRepository
	redis       *redislib.Client
	cfg         *config.ContestConfig
	broadcaster contestports.RealtimeBroadcaster
}

func NewScoreboardAdminService(repo contestports.ContestScoreboardAdminRepository, redis *redislib.Client, cfg *config.ContestConfig) *ScoreboardAdminService {
	return &ScoreboardAdminService{repo: repo, redis: redis, cfg: cfg}
}

func (s *ScoreboardAdminService) SetRealtimeBroadcaster(broadcaster contestports.RealtimeBroadcaster) {
	s.broadcaster = broadcaster
}
