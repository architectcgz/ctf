package http

import (
	"context"

	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type participationCommandService interface {
	RegisterContest(ctx context.Context, contestID, userID int64) error
	ReviewRegistration(ctx context.Context, contestID, registrationID, reviewerID int64, req contestcmd.ReviewRegistrationInput) (*contestcmd.ContestRegistrationResp, error)
	CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req contestcmd.CreateAnnouncementInput) (*contestcmd.ContestAnnouncementResp, error)
	DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) error
}

type participationQueryService interface {
	ListRegistrations(ctx context.Context, contestID int64, query contestqry.ContestRegistrationQueryInput) (*contestqry.RegistrationPageResult[*contestqry.ContestRegistrationResult], error)
	ListAnnouncements(ctx context.Context, contestID int64) ([]*contestqry.ContestAnnouncementResult, error)
	SyncAnnouncements(ctx context.Context, contestID int64, afterID *int64) (*contestqry.ContestAnnouncementSyncResult, error)
	GetMyProgress(ctx context.Context, contestID, userID int64) (*contestqry.ParticipationProgressResult, error)
}

type ParticipationHandler struct {
	commands participationCommandService
	queries  participationQueryService
}

func NewParticipationHandler(commands participationCommandService, queries participationQueryService) *ParticipationHandler {
	return &ParticipationHandler{commands: commands, queries: queries}
}
