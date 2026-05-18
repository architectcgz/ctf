package ports

import (
	"context"
	"errors"
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

var ErrContestParticipationRegistrationNotFound = errors.New("contest participation registration not found")

type ContestParticipationRegistrationLookupRepository interface {
	FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error)
	FindRegistrationByID(ctx context.Context, contestID, registrationID int64) (*contestentity.ContestRegistration, error)
}

type ContestParticipationRegistrationWriteRepository interface {
	CreateRegistration(ctx context.Context, registration *contestentity.ContestRegistration) error
	SaveRegistration(ctx context.Context, registration *contestentity.ContestRegistration) error
}

type ContestParticipationRegistrationListRepository interface {
	ListRegistrations(ctx context.Context, contestID int64, status *string, offset, limit int) ([]*ContestParticipationRegistrationRow, int64, error)
}

type ContestParticipationUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error)
}

type ContestParticipationAnnouncementReadRepository interface {
	ListAnnouncements(ctx context.Context, contestID int64) ([]*contestentity.ContestAnnouncement, error)
}

type ContestParticipationAnnouncementWriteRepository interface {
	CreateAnnouncement(ctx context.Context, announcement *contestentity.ContestAnnouncement) error
	DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) (bool, error)
}

type ContestParticipationProgressRepository interface {
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
