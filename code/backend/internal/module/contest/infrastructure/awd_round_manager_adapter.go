package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type AWDRoundManagerAdapter struct {
	source contestports.AWDRoundManager
}

func NewAWDRoundManagerAdapter(source contestports.AWDRoundManager) *AWDRoundManagerAdapter {
	if source == nil {
		return nil
	}
	return &AWDRoundManagerAdapter{source: source}
}

func (r *AWDRoundManagerAdapter) RunRoundServiceChecks(ctx context.Context, contest *model.Contest, round *model.AWDRound, source string) error {
	return r.source.RunRoundServiceChecks(ctx, contest, round, source)
}

func (r *AWDRoundManagerAdapter) EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	err := r.source.EnsureActiveRoundMaterialized(ctx, contest, now)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return contestports.ErrContestAWDRoundNotFound
	}
	return err
}

func (r *AWDRoundManagerAdapter) PreviewServiceCheck(ctx context.Context, req contestports.AWDServicePreviewRequest) (*contestports.AWDServicePreviewResult, error) {
	return r.source.PreviewServiceCheck(ctx, req)
}

var _ contestports.AWDRoundManager = (*AWDRoundManagerAdapter)(nil)
