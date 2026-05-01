package http

import (
	"context"

	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type participationCommandService interface {
	RegisterContest(ctx context.Context, contestID, userID int64) error
	ReviewRegistration(ctx context.Context, contestID, registrationID, reviewerID int64, req *dto.ReviewContestRegistrationReq) (*dto.ContestRegistrationResp, error)
	CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req contestcmd.CreateAnnouncementInput) (*dto.ContestAnnouncementResp, error)
	DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) error
}

type participationQueryService interface {
	ListRegistrations(ctx context.Context, contestID int64, query contestqry.ContestRegistrationQueryInput) (*contestqry.RegistrationPageResult[*contestqry.ContestRegistrationResult], error)
	ListAnnouncements(ctx context.Context, contestID int64) ([]*contestqry.ContestAnnouncementResult, error)
	GetMyProgress(ctx context.Context, contestID, userID int64) (*contestqry.ParticipationProgressResult, error)
}

type ParticipationHandler struct {
	commands participationCommandService
	queries  participationQueryService
}

func NewParticipationHandler(commands participationCommandService, queries participationQueryService) *ParticipationHandler {
	return &ParticipationHandler{commands: commands, queries: queries}
}
