package infrastructure

import (
	"context"
	"errors"
	"fmt"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
	"gorm.io/gorm"
)

func (r *ParticipationRepository) ListAnnouncements(ctx context.Context, contestID int64) ([]*contestentity.ContestAnnouncement, error) {
	var announcements []*contestentity.ContestAnnouncement
	if err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("created_at DESC, id DESC").
		Find(&announcements).Error; err != nil {
		return nil, err
	}
	return announcements, nil
}

func (r *ParticipationRepository) ListAnnouncementSyncEvents(ctx context.Context, contestID, afterID int64, limit int) ([]*contestports.ContestAnnouncementSyncEventRow, error) {
	if limit <= 0 {
		limit = 100
	}

	var rows []contestentity.ContestRealtimeOutbox
	if err := r.dbWithContext(ctx).
		Where("channel = ? AND status = ? AND id > ?", contestcontracts.AnnouncementChannel(contestID), "sent", afterID).
		Where("event_name IN ?", []string{
			contestcontracts.EventAnnouncementCreated,
			contestcontracts.EventAnnouncementDeleted,
		}).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	events := make([]*contestports.ContestAnnouncementSyncEventRow, 0, len(rows))
	for _, row := range rows {
		relay, err := decodeRealtimeRelay(row)
		if err != nil {
			return nil, fmt.Errorf("decode announcement sync relay: %w", err)
		}

		event := &contestports.ContestAnnouncementSyncEventRow{
			Cursor:      row.ID,
			MessageType: relay.MessageType,
			OccurredAt:  relay.Timestamp,
		}

		switch payload := relay.Payload.(type) {
		case contestcontracts.AnnouncementCreatedRelayPayload:
			announcement := payload.Announcement
			event.Announcement = &announcement
		case contestcontracts.AnnouncementDeletedRelayPayload:
			announcementID := payload.AnnouncementID
			event.AnnouncementID = &announcementID
		default:
			return nil, fmt.Errorf("unsupported announcement sync payload type: %T", relay.Payload)
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *ParticipationRepository) LatestAnnouncementSyncCursor(ctx context.Context, contestID int64) (int64, error) {
	var row struct {
		ID int64 `gorm:"column:id"`
	}

	err := r.dbWithContext(ctx).
		Model(&contestentity.ContestRealtimeOutbox{}).
		Select("id").
		Where("channel = ? AND status = ?", contestcontracts.AnnouncementChannel(contestID), "sent").
		Where("event_name IN ?", []string{
			contestcontracts.EventAnnouncementCreated,
			contestcontracts.EventAnnouncementDeleted,
		}).
		Order("id DESC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.ID, nil
}

func (r *ParticipationRepository) CreateAnnouncement(ctx context.Context, announcement *contestentity.ContestAnnouncement) error {
	return r.dbWithContext(ctx).Create(announcement).Error
}

func (r *ParticipationRepository) DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) (bool, error) {
	result := r.dbWithContext(ctx).
		Where("id = ? AND contest_id = ?", announcementID, contestID).
		Delete(&contestentity.ContestAnnouncement{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
