package domain

import (
	contestentity "ctf-platform/internal/module/contest/entity"
	"ctf-platform/pkg/errcode"
)

func RegistrationStatusError(status string) error {
	switch status {
	case "", contestentity.ContestRegistrationStatusApproved:
		return nil
	case contestentity.ContestRegistrationStatusPending:
		return errcode.ErrContestRegistrationPending
	case contestentity.ContestRegistrationStatusRejected:
		return errcode.ErrRegistrationNotApproved
	default:
		return errcode.ErrRegistrationNotApproved
	}
}
