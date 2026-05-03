package domain

import "time"

const (
	ContestStatusTransitionReasonTimeWindow   = "time_window"
	ContestStatusTransitionReasonManualUpdate = "manual_update"

	ContestStatusTransitionSideEffectPending   = "pending"
	ContestStatusTransitionSideEffectSucceeded = "succeeded"
	ContestStatusTransitionSideEffectFailed    = "failed"
)

type ContestStatusTransition struct {
	ContestID         int64
	FromStatus        string
	ToStatus          string
	FromStatusVersion int64
	Reason            string
	OccurredAt        time.Time
	AppliedBy         string
}

type ContestStatusTransitionResult struct {
	Transition    ContestStatusTransition
	Applied       bool
	StatusVersion int64
	RecordID      int64
}

func ValidateStatusTransition(from, to string) error {
	if from == "" || to == "" || from == to || !IsValidTransition(from, to) {
		return ErrInvalidStatusTransition
	}
	return nil
}
