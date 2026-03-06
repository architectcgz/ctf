package contest

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type Service interface {
	CreateContest(ctx context.Context, req *dto.CreateContestReq) (*dto.ContestResp, error)
	UpdateContest(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error)
	GetContest(ctx context.Context, id int64) (*dto.ContestResp, error)
	ListContests(ctx context.Context, req *dto.ListContestsReq) ([]*dto.ContestResp, int64, error)
}

type service struct {
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &service{repo: repo, log: log}
}

func (s *service) CreateContest(ctx context.Context, req *dto.CreateContestReq) (*dto.ContestResp, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, errcode.ErrInvalidTimeRange
	}

	contest := &model.Contest{
		Title:       req.Title,
		Description: req.Description,
		Mode:        req.Mode,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      model.ContestStatusDraft,
	}

	if err := s.repo.Create(ctx, contest); err != nil {
		s.log.Error("create_contest_failed", zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_created", zap.Int64("contest_id", contest.ID))
	return toContestResp(contest), nil
}

func (s *service) UpdateContest(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if req.Title != nil {
		contest.Title = *req.Title
	}
	if req.Description != nil {
		contest.Description = *req.Description
	}
	if req.Mode != nil {
		contest.Mode = *req.Mode
	}
	if req.StartTime != nil {
		contest.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		contest.EndTime = *req.EndTime
	}
	if req.Status != nil {
		contest.Status = *req.Status
	}

	if !contest.EndTime.After(contest.StartTime) {
		return nil, errcode.ErrInvalidTimeRange
	}

	if err := s.repo.Update(ctx, contest); err != nil {
		s.log.Error("update_contest_failed", zap.Int64("contest_id", id), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_updated", zap.Int64("contest_id", id))
	return toContestResp(contest), nil
}

func (s *service) GetContest(ctx context.Context, id int64) (*dto.ContestResp, error) {
	contest, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toContestResp(contest), nil
}

func (s *service) ListContests(ctx context.Context, req *dto.ListContestsReq) ([]*dto.ContestResp, int64, error) {
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
		resp[i] = toContestResp(c)
	}
	return resp, total, nil
}

func toContestResp(contest *model.Contest) *dto.ContestResp {
	return &dto.ContestResp{
		ID:          contest.ID,
		Title:       contest.Title,
		Description: contest.Description,
		Mode:        contest.Mode,
		StartTime:   contest.StartTime,
		EndTime:     contest.EndTime,
		Status:      contest.Status,
		CreatedAt:   contest.CreatedAt,
		UpdatedAt:   contest.UpdatedAt,
	}
}
