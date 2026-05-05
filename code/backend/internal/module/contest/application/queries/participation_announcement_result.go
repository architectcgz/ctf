package queries

import "time"

type ContestAnnouncementResult struct {
	ID        int64
	Title     string
	Content   string
	CreatedAt time.Time
}
