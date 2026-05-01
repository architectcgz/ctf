package http

import (
	"context"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type challengeCommandService interface {
	AddChallengeToContest(ctx context.Context, contestID int64, req *dto.AddContestChallengeReq) (*dto.ContestChallengeResp, error)
	RemoveChallengeFromContest(ctx context.Context, contestID, challengeID int64) error
	UpdateChallenge(ctx context.Context, contestID, challengeID int64, req *dto.UpdateContestChallengeReq) error
}

type challengeQueryService interface {
	GetContestChallenges(ctx context.Context, userID, contestID int64) ([]*contestqry.ContestChallengeInfoResult, error)
	ListAdminChallenges(ctx context.Context, contestID int64) ([]*contestqry.ContestChallengeResult, error)
}

type ChallengeHandler struct {
	commands challengeCommandService
	queries  challengeQueryService
}

func NewChallengeHandler(commands challengeCommandService, queries challengeQueryService) *ChallengeHandler {
	return &ChallengeHandler{commands: commands, queries: queries}
}
