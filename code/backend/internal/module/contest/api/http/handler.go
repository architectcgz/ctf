package http

import (
	"context"

	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type contestService interface {
	CreateContest(ctx context.Context, req contestcmd.CreateContestInput) (*dto.ContestResp, error)
	UpdateContest(ctx context.Context, id int64, req contestcmd.UpdateContestInput) (*dto.ContestResp, error)
}

type contestQueryService interface {
	GetContest(ctx context.Context, id int64) (*contestqry.ContestResult, error)
	ListContests(ctx context.Context, req contestqry.ListContestsInput) ([]*contestqry.ContestResult, int64, error)
}

type scoreboardQueryService interface {
	GetScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*contestqry.ScoreboardResult, error)
	GetLiveScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*contestqry.ScoreboardResult, error)
}

type scoreboardCommandService interface {
	FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error
	UnfreezeScoreboard(ctx context.Context, contestID int64) error
}

type Handler struct {
	commands          contestService
	queries           contestQueryService
	readinessQueries  awdReadinessQueryService
	scoreboardQueries scoreboardQueryService
	scoreboardCommand scoreboardCommandService
}

func NewHandler(commands contestService, queries contestQueryService, readinessQueries awdReadinessQueryService, scoreboardQueries scoreboardQueryService, scoreboardCommand scoreboardCommandService) *Handler {
	return &Handler{
		commands:          commands,
		queries:           queries,
		readinessQueries:  readinessQueries,
		scoreboardQueries: scoreboardQueries,
		scoreboardCommand: scoreboardCommand,
	}
}
