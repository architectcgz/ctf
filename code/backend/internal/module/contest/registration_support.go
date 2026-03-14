package contest

import (
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func registrationStatusError(status string) error {
	switch status {
	case "", model.ContestRegistrationStatusApproved:
		return nil
	case model.ContestRegistrationStatusPending:
		return errcode.ErrContestRegistrationPending
	case model.ContestRegistrationStatusRejected:
		return errcode.ErrRegistrationNotApproved
	default:
		return errcode.ErrRegistrationNotApproved
	}
}
