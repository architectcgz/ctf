package domain

import (
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func RegistrationStatusError(status string) error {
	switch status {
	case "", contestentity.ContestRegistrationStatusApproved:
		return nil
	case contestentity.ContestRegistrationStatusPending:
		return contestcontracts.ErrContestRegistrationPending
	case contestentity.ContestRegistrationStatusRejected:
		return contestcontracts.ErrRegistrationNotApproved
	default:
		return contestcontracts.ErrRegistrationNotApproved
	}
}
