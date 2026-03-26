package http

import (
	"context"

	"ctf-platform/internal/dto"
)

type awdCommandService interface {
	CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error)
	RunCurrentRoundChecks(ctx context.Context, contestID int64) (*dto.AWDCheckerRunResp, error)
	RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error)
	UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error)
	CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error)
	SubmitAttack(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error)
}

type awdQueryService interface {
	ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error)
	ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error)
	ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]*dto.AWDAttackLogResp, error)
	GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error)
}

type AWDHandler struct {
	commands awdCommandService
	queries  awdQueryService
}

func NewAWDHandler(commands awdCommandService, queries awdQueryService) *AWDHandler {
	return &AWDHandler{commands: commands, queries: queries}
}
