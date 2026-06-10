package queries

import "time"

type ContestAnnouncementResult struct {
	ID        int64
	Title     string
	Content   string
	CreatedAt time.Time
}

type ContestAnnouncementSyncEventResult struct {
	Cursor         int64
	Type           string
	Announcement   *ContestAnnouncementResult
	AnnouncementID *int64
	OccurredAt     time.Time
}

type ContestAnnouncementSyncResult struct {
	Events     []*ContestAnnouncementSyncEventResult
	NextCursor int64
	HasMore    bool
}
