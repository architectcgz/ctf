package http

import (
	"context"

	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type awdCommandService interface {
	CreateRound(ctx context.Context, contestID int64, req contestcmd.CreateAWDRoundInput) (*dto.AWDRoundResp, error)
	RunCurrentRoundChecks(ctx context.Context, contestID int64, req contestcmd.RunCurrentRoundChecksInput) (*dto.AWDCheckerRunResp, error)
	RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error)
	PreviewChecker(ctx context.Context, contestID int64, req contestcmd.PreviewCheckerInput) (*dto.AWDCheckerPreviewResp, error)
	UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req contestcmd.UpsertServiceCheckInput) (*dto.AWDTeamServiceResp, error)
	CreateAttackLog(ctx context.Context, contestID, roundID int64, req contestcmd.CreateAttackLogInput) (*dto.AWDAttackLogResp, error)
	SubmitAttack(ctx context.Context, userID, contestID, serviceID int64, req contestcmd.SubmitAttackInput) (*dto.AWDAttackLogResp, error)
}

type awdServiceCommandService interface {
	CreateContestAWDService(ctx context.Context, contestID int64, req *dto.CreateContestAWDServiceReq) (*dto.ContestAWDServiceResp, error)
	UpdateContestAWDService(ctx context.Context, contestID, serviceID int64, req *dto.UpdateContestAWDServiceReq) error
	DeleteContestAWDService(ctx context.Context, contestID, serviceID int64) error
}

type awdQueryService interface {
	ListRounds(ctx context.Context, contestID int64) ([]contestqry.AWDRoundResult, error)
	ListServices(ctx context.Context, contestID, roundID int64) ([]contestqry.AWDTeamServiceResult, error)
	ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]contestqry.AWDAttackLogResult, error)
	GetRoundSummary(ctx context.Context, contestID, roundID int64) (*contestqry.AWDRoundSummaryResult, error)
	GetTrafficSummary(ctx context.Context, contestID, roundID int64) (*contestqry.AWDTrafficSummaryResult, error)
	ListTrafficEvents(ctx context.Context, contestID, roundID int64, req *contestqry.ListAWDTrafficEventsInput) (*contestqry.AWDTrafficEventPageResult, error)
	GetReadiness(ctx context.Context, contestID int64) (*contestqry.AWDReadinessResult, error)
	GetUserWorkspace(ctx context.Context, userID, contestID int64) (*contestqry.AWDWorkspaceResult, error)
}

type awdServiceQueryService interface {
	ListContestAWDServices(ctx context.Context, contestID int64) ([]contestqry.ContestAWDServiceResult, error)
}

type AWDHandler struct {
	commands        awdCommandService
	queries         awdQueryService
	serviceCommands awdServiceCommandService
	serviceQueries  awdServiceQueryService
}

func NewAWDHandler(
	commands awdCommandService,
	queries awdQueryService,
	serviceCommands awdServiceCommandService,
	serviceQueries awdServiceQueryService,
) *AWDHandler {
	return &AWDHandler{
		commands:        commands,
		queries:         queries,
		serviceCommands: serviceCommands,
		serviceQueries:  serviceQueries,
	}
}
