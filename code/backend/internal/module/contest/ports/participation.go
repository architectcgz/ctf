package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type ContestParticipationRepository interface {
	FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	FindRegistrationByID(ctx context.Context, contestID, registrationID int64) (*model.ContestRegistration, error)
	CreateRegistration(ctx context.Context, registration *model.ContestRegistration) error
	SaveRegistration(ctx context.Context, registration *model.ContestRegistration) error
	ListRegistrations(ctx context.Context, contestID int64, status *string, offset, limit int) ([]*ContestParticipationRegistrationRow, int64, error)
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	ListAnnouncements(ctx context.Context, contestID int64) ([]*model.ContestAnnouncement, error)
	CreateAnnouncement(ctx context.Context, announcement *model.ContestAnnouncement) error
	DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) (bool, error)
	ListSolvedProgress(ctx context.Context, contestID, userID int64) ([]*ContestParticipationSolvedProgressRow, error)
}

type ContestParticipationRegistrationRow struct {
	ID         int64
	ContestID  int64
	UserID     int64
	Username   string
	TeamID     *int64
	Status     string
	ReviewedBy *int64
	ReviewedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ContestParticipationSolvedProgressRow struct {
	ContestChallengeID int64
	SolvedAt           time.Time
	PointsEarned       int
}
