package http

import (
	"context"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type awdCommandService interface {
	CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error)
	RunCurrentRoundChecks(ctx context.Context, contestID int64, req *dto.RunCurrentAWDCheckerReq) (*dto.AWDCheckerRunResp, error)
	RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error)
	PreviewChecker(ctx context.Context, contestID int64, req *dto.PreviewAWDCheckerReq) (*dto.AWDCheckerPreviewResp, error)
	UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error)
	CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error)
	SubmitAttack(ctx context.Context, userID, contestID, serviceID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error)
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
	GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error)
	GetTrafficSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDTrafficSummaryResp, error)
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
