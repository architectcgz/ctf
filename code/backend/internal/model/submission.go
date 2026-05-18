package model

import contestcontracts "ctf-platform/internal/module/contest/contracts"

const (
	SubmissionReviewStatusNotRequired = contestcontracts.SubmissionReviewStatusNotRequired
	SubmissionReviewStatusPending     = contestcontracts.SubmissionReviewStatusPending
	SubmissionReviewStatusApproved    = contestcontracts.SubmissionReviewStatusApproved
	SubmissionReviewStatusRejected    = contestcontracts.SubmissionReviewStatusRejected
)

type Submission = contestcontracts.Submission
