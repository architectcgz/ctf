package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

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

func (r *ParticipationRepository) ListRegistrations(ctx context.Context, contestID int64, status *string, offset, limit int) ([]*contestports.ContestParticipationRegistrationRow, int64, error) {
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

	var rows []*contestports.ContestParticipationRegistrationRow
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
