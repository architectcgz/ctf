package composition

import (
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/platform/events"
)

type Root struct {
	Events events.Bus

	cfg   *config.Config
	log   *zap.Logger
	db    *gorm.DB
	cache *redislib.Client
}

func BuildRoot(cfg *config.Config, log *zap.Logger, db *gorm.DB, cache *redislib.Client) (*Root, error) {
	return &Root{
		Events: events.NewBus(),
		cfg:    cfg,
		log:    log,
		db:     db,
		cache:  cache,
	}, nil
}

func (r *Root) Config() *config.Config {
	return r.cfg
}

func (r *Root) Logger() *zap.Logger {
	return r.log
}

func (r *Root) DB() *gorm.DB {
	return r.db
}

func (r *Root) Cache() *redislib.Client {
	return r.cache
}
