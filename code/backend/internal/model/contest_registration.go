package model

import contestentity "ctf-platform/internal/module/contest/entity"

const (
	ContestRegistrationStatusPending  = contestentity.ContestRegistrationStatusPending
	ContestRegistrationStatusApproved = contestentity.ContestRegistrationStatusApproved
	ContestRegistrationStatusRejected = contestentity.ContestRegistrationStatusRejected
)

type ContestRegistration = contestentity.ContestRegistration
