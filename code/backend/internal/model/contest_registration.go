package model

import contestcontracts "ctf-platform/internal/module/contest/contracts"

const (
	ContestRegistrationStatusPending  = contestcontracts.ContestRegistrationStatusPending
	ContestRegistrationStatusApproved = contestcontracts.ContestRegistrationStatusApproved
	ContestRegistrationStatusRejected = contestcontracts.ContestRegistrationStatusRejected
)

type ContestRegistration = contestcontracts.ContestRegistration
