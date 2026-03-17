package contest

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

type ParticipationRepository struct {
	db *gorm.DB
}

type participationRegistrationRow struct {
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

type participationSolvedProgressRow struct {
	ContestChallengeID int64
	SolvedAt           time.Time
	PointsEarned       int
}

func NewParticipationRepository(db *gorm.DB) *ParticipationRepository {
	return &ParticipationRepository{db: db}
}

func (r *ParticipationRepository) dbWithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx)
}

func (r *ParticipationRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *ParticipationRepository) FindRegistrationByID(ctx context.Context, contestID, registrationID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("id = ? AND contest_id = ?", registrationID, contestID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *ParticipationRepository) CreateRegistration(ctx context.Context, registration *model.ContestRegistration) error {
	return r.dbWithContext(ctx).Create(registration).Error
}

func (r *ParticipationRepository) SaveRegistration(ctx context.Context, registration *model.ContestRegistration) error {
	return r.dbWithContext(ctx).Save(registration).Error
}

func (r *ParticipationRepository) ListRegistrations(ctx context.Context, contestID int64, status *string, offset, limit int) ([]*participationRegistrationRow, int64, error) {
	baseQuery := r.dbWithContext(ctx).
		Table("contest_registrations AS cr").
		Joins("JOIN users u ON u.id = cr.user_id").
		Where("cr.contest_id = ?", contestID)
	if status != nil {
		baseQuery = baseQuery.Where("cr.status = ?", *status)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []*participationRegistrationRow
	if err := baseQuery.
		Select("cr.id, cr.contest_id, cr.user_id, u.username, cr.team_id, cr.status, cr.reviewed_by, cr.reviewed_at, cr.created_at, cr.updated_at").
		Order("cr.created_at ASC, cr.id ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *ParticipationRepository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.dbWithContext(ctx).Select("id, username").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ParticipationRepository) ListAnnouncements(ctx context.Context, contestID int64) ([]*model.ContestAnnouncement, error) {
	var announcements []*model.ContestAnnouncement
	if err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("created_at DESC, id DESC").
		Find(&announcements).Error; err != nil {
		return nil, err
	}
	return announcements, nil
}

func (r *ParticipationRepository) CreateAnnouncement(ctx context.Context, announcement *model.ContestAnnouncement) error {
	return r.dbWithContext(ctx).Create(announcement).Error
}

func (r *ParticipationRepository) DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) (bool, error) {
	result := r.dbWithContext(ctx).
		Where("id = ? AND contest_id = ?", announcementID, contestID).
		Delete(&model.ContestAnnouncement{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ParticipationRepository) ListSolvedProgress(ctx context.Context, contestID, userID int64) ([]*participationSolvedProgressRow, error) {
	var rows []*participationSolvedProgressRow
	if err := r.dbWithContext(ctx).
		Table("submissions AS s").
		Select("cc.id AS contest_challenge_id, s.submitted_at AS solved_at, s.score AS points_earned").
		Joins("JOIN contest_challenges cc ON cc.contest_id = s.contest_id AND cc.challenge_id = s.challenge_id").
		Where("s.contest_id = ? AND s.user_id = ? AND s.is_correct = ?", contestID, userID, true).
		Order("s.submitted_at ASC, s.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
