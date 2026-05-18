package contracts

import contestentity "ctf-platform/internal/module/contest/entity"

type Submission = contestentity.Submission

const (
	SubmissionReviewStatusNotRequired = contestentity.SubmissionReviewStatusNotRequired
	SubmissionReviewStatusPending     = contestentity.SubmissionReviewStatusPending
	SubmissionReviewStatusApproved    = contestentity.SubmissionReviewStatusApproved
	SubmissionReviewStatusRejected    = contestentity.SubmissionReviewStatusRejected
)

type ContestRegistration = contestentity.ContestRegistration

const (
	ContestRegistrationStatusPending  = contestentity.ContestRegistrationStatusPending
	ContestRegistrationStatusApproved = contestentity.ContestRegistrationStatusApproved
	ContestRegistrationStatusRejected = contestentity.ContestRegistrationStatusRejected
)

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
