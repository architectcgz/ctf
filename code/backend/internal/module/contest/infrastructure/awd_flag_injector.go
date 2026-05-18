package infrastructure

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type noopAWDFlagInjector struct {
	log *zap.Logger
}

func NewNoopAWDFlagInjector(log *zap.Logger) contestports.AWDFlagInjector {
	if log == nil {
		log = zap.NewNop()
	}
	return &noopAWDFlagInjector{log: log}
}

func (i *noopAWDFlagInjector) InjectRoundFlags(_ context.Context, contest *contestentity.Contest, round *contestentity.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	i.log.Debug("skip_awd_flag_injection",
		zap.Int64("contest_id", contest.ID),
		zap.Int64("round_id", round.ID),
		zap.Int("assignment_count", len(assignments)),
	)
	return nil
}

type dockerAWDFlagInjector struct {
	db           *gorm.DB
	writer       contestports.AWDContainerFileWriter
	flagFilePath string
	log          *zap.Logger
}

func NewDockerAWDFlagInjector(db *gorm.DB, writer contestports.AWDContainerFileWriter, log *zap.Logger) contestports.AWDFlagInjector {
	if writer == nil {
		return NewNoopAWDFlagInjector(log)
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &dockerAWDFlagInjector{
		db:           db,
		writer:       writer,
		flagFilePath: "/flag",
		log:          log,
	}
}
