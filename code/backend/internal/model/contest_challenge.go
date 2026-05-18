package model

import contestentity "ctf-platform/internal/module/contest/entity"

type AWDCheckerType = contestentity.AWDCheckerType

const (
	AWDCheckerTypeLegacyProbe  = contestentity.AWDCheckerTypeLegacyProbe
	AWDCheckerTypeHTTPStandard = contestentity.AWDCheckerTypeHTTPStandard
	AWDCheckerTypeTCPStandard  = contestentity.AWDCheckerTypeTCPStandard
	AWDCheckerTypeScript       = contestentity.AWDCheckerTypeScript
)

type AWDCheckerValidationState = contestentity.AWDCheckerValidationState

const (
	AWDCheckerValidationStatePending = contestentity.AWDCheckerValidationStatePending
	AWDCheckerValidationStatePassed  = contestentity.AWDCheckerValidationStatePassed
	AWDCheckerValidationStateFailed  = contestentity.AWDCheckerValidationStateFailed
	AWDCheckerValidationStateStale   = contestentity.AWDCheckerValidationStateStale
)

type ContestChallenge = contestentity.ContestChallenge
