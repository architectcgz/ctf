package model

import contestentity "ctf-platform/internal/module/contest/entity"

const (
	ContestModeJeopardy = contestentity.ContestModeJeopardy
	ContestModeAWD      = contestentity.ContestModeAWD

	ContestStatusDraft        = contestentity.ContestStatusDraft
	ContestStatusRegistration = contestentity.ContestStatusRegistration
	ContestStatusRunning      = contestentity.ContestStatusRunning
	ContestStatusFrozen       = contestentity.ContestStatusFrozen
	ContestStatusEnded        = contestentity.ContestStatusEnded
)

type Contest = contestentity.Contest
