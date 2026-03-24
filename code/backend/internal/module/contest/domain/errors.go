package domain

import "errors"

var (
	ErrContestNotFound      = errors.New("contest not found")
	ErrTeamFull             = errors.New("team is full")
	ErrAlreadyJoinedContest = errors.New("user already joined another team in contest")
)
