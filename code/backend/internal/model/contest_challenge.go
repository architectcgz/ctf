package model

import contestcontracts "ctf-platform/internal/module/contest/contracts"

type AWDCheckerType = contestcontracts.AWDCheckerType

const (
	AWDCheckerTypeLegacyProbe  = contestcontracts.AWDCheckerTypeLegacyProbe
	AWDCheckerTypeHTTPStandard = contestcontracts.AWDCheckerTypeHTTPStandard
	AWDCheckerTypeTCPStandard  = contestcontracts.AWDCheckerTypeTCPStandard
	AWDCheckerTypeScript       = contestcontracts.AWDCheckerTypeScript
)

type AWDCheckerValidationState = contestcontracts.AWDCheckerValidationState

const (
	AWDCheckerValidationStatePending = contestcontracts.AWDCheckerValidationStatePending
	AWDCheckerValidationStatePassed  = contestcontracts.AWDCheckerValidationStatePassed
	AWDCheckerValidationStateFailed  = contestcontracts.AWDCheckerValidationStateFailed
	AWDCheckerValidationStateStale   = contestcontracts.AWDCheckerValidationStateStale
)

type ContestChallenge = contestcontracts.ContestChallenge
