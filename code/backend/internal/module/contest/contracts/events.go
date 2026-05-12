package contracts

import (
	"fmt"
	"time"
)

const (
	EventAnnouncementCreated = "contest.announcement_created"
	EventAnnouncementDeleted = "contest.announcement_deleted"
	EventAWDAttackAccepted   = "contest.awd.attack_accepted"
	EventAWDPreviewProgress  = "contest.awd_preview_progress"
	EventScoreboardUpdated   = "contest.scoreboard_updated"
)

func AnnouncementChannel(contestID int64) string {
	return fmt.Sprintf("contest:%d:announcements", contestID)
}

func ScoreboardChannel(contestID int64) string {
	return fmt.Sprintf("contest:%d:scoreboard", contestID)
}

type AnnouncementCreatedEvent struct {
	ContestID      int64
	AnnouncementID int64
	Title          string
	Content        string
	CreatedAt      time.Time
	OccurredAt     time.Time
}

type AnnouncementDeletedEvent struct {
	ContestID      int64
	AnnouncementID int64
	OccurredAt     time.Time
}

type AWDAttackAcceptedEvent struct {
	UserID         int64
	ContestID      int64
	AWDChallengeID int64
	Dimension      string
	OccurredAt     time.Time
}

type AWDPreviewProgressEvent struct {
	UserID           int64
	ContestID        int64
	PreviewRequestID string
	PhaseKey         string
	PhaseLabel       string
	Detail           string
	Attempt          int
	TotalAttempts    int
	Status           string
	Error            string
	OccurredAt       time.Time
}

type ScoreboardUpdatedEvent struct {
	ContestID  int64
	OccurredAt time.Time
}
