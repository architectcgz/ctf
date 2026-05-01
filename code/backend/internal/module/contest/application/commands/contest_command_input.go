package commands

import "time"

type CreateContestInput struct {
	Title       string
	Description string
	Mode        string
	StartTime   time.Time
	EndTime     time.Time
}

type UpdateContestInput struct {
	Title          *string
	Description    *string
	Mode           *string
	StartTime      *time.Time
	EndTime        *time.Time
	Status         *string
	ForceOverride  *bool
	OverrideReason *string
}
