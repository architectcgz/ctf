package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
)

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
