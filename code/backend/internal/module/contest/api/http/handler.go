package http

import (
	"context"

	"ctf-platform/internal/dto"
)

type contestService interface {
	CreateContest(ctx context.Context, req *dto.CreateContestReq) (*dto.ContestResp, error)
	UpdateContest(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error)
}

type contestQueryService interface {
	GetContest(ctx context.Context, id int64) (*dto.ContestResp, error)
	ListContests(ctx context.Context, req *dto.ListContestsReq) ([]*dto.ContestResp, int64, error)
}

type scoreboardQueryService interface {
	GetScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error)
	GetLiveScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error)
}

type scoreboardCommandService interface {
	FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error
	UnfreezeScoreboard(ctx context.Context, contestID int64) error
}

type Handler struct {
	commands          contestService
	queries           contestQueryService
	scoreboardQueries scoreboardQueryService
	scoreboardCommand scoreboardCommandService
}

func NewHandler(commands contestService, queries contestQueryService, scoreboardQueries scoreboardQueryService, scoreboardCommand scoreboardCommandService) *Handler {
	return &Handler{
		commands:          commands,
		queries:           queries,
		scoreboardQueries: scoreboardQueries,
		scoreboardCommand: scoreboardCommand,
	}
}
