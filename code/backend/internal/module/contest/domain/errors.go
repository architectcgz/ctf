package domain

import "errors"

var (
	ErrContestNotFound         = errors.New("contest not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrTeamFull                = errors.New("team is full")
	ErrAlreadyJoinedContest    = errors.New("user already joined another team in contest")
)
