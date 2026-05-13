package ports_test

import (
	"context"
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ctxOnlyAWDCheckerPreviewTokenStore struct{}

func (ctxOnlyAWDCheckerPreviewTokenStore) StoreAWDCheckerPreviewToken(context.Context, contestports.AWDCheckerPreviewTokenRecord, time.Duration) (string, error) {
	return "", nil
}

func (ctxOnlyAWDCheckerPreviewTokenStore) LoadAWDCheckerPreviewToken(context.Context, int64, string) (*contestports.AWDCheckerPreviewTokenRecord, bool, error) {
	return nil, false, nil
}

func (ctxOnlyAWDCheckerPreviewTokenStore) DeleteAWDCheckerPreviewToken(context.Context, int64, string) error {
	return nil
}

var _ contestports.AWDCheckerPreviewTokenStore = (*ctxOnlyAWDCheckerPreviewTokenStore)(nil)
