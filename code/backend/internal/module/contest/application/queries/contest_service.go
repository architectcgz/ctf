package queries

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type ContestService struct {
	repo contestports.Repository
	log  *zap.Logger
}

func NewContestService(repo contestports.Repository, log *zap.Logger) *ContestService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ContestService{repo: repo, log: log}
}

func (s *ContestService) GetContest(ctx context.Context, id int64) (*dto.ContestResp, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.ContestRespFromModel(contest), nil
}

func (s *ContestService) ListContests(ctx context.Context, req *dto.ListContestsReq) ([]*dto.ContestResp, int64, error) {
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

	resp := make([]*dto.ContestResp, len(contests))
	for i, c := range contests {
		resp[i] = domain.ContestRespFromModel(c)
	}
	return resp, total, nil
}
