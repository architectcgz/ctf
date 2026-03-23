package application

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/model"
)

var (
	ErrContestNotFound = errors.New("contest not found")
)

type Repository interface {
	Create(ctx context.Context, contest *model.Contest) error
	FindByID(ctx context.Context, id int64) (*model.Contest, error)
	Update(ctx context.Context, contest *model.Contest) error
	List(ctx context.Context, status *string, offset, limit int) ([]*model.Contest, int64, error)
	ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	FindTeamsByIDs(ctx context.Context, ids []int64) ([]*model.Team, error)
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
	FindScoreboardTeamStats(ctx context.Context, contestID int64, contestMode string, teamIDs []int64) (map[int64]ScoreboardTeamStats, error)
}

type ScoreboardTeamStats struct {
	SolvedCount      int
	LastSubmissionAt *time.Time
}

type ContestChallengeRepository interface {
	AddChallenge(ctx context.Context, cc *model.ContestChallenge) error
	RemoveChallenge(ctx context.Context, contestID, challengeID int64) error
	UpdateChallenge(ctx context.Context, contestID, challengeID int64, updates map[string]any) error
	ListChallenges(ctx context.Context, contestID int64, visibleOnly bool) ([]*model.ContestChallenge, error)
	Exists(ctx context.Context, contestID, challengeID int64) (bool, error)
	HasSubmissions(ctx context.Context, contestID, challengeID int64) (bool, error)
}

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

type ContestTeamFinder interface {
	FindUserTeamInContest(userID, contestID int64) (*model.Team, error)
}
