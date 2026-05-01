package queries

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/pkg/errcode"
)

func (s *ContestService) ListContests(ctx context.Context, req ListContestsInput) ([]*ContestResult, int64, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size
	contests, total, err := s.repo.List(ctx, req.Status, offset, size)
	if err != nil {
		s.log.Error("list_contests_failed", zap.Error(err))
		return nil, 0, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*ContestResult, len(contests))
	for i, c := range contests {
		resp[i] = contestResultFromModel(c)
	}
	return resp, total, nil
}
