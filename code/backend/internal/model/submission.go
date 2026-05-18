package model

import contestentity "ctf-platform/internal/module/contest/entity"

const (
	SubmissionReviewStatusNotRequired = contestentity.SubmissionReviewStatusNotRequired
	SubmissionReviewStatusPending     = contestentity.SubmissionReviewStatusPending
	SubmissionReviewStatusApproved    = contestentity.SubmissionReviewStatusApproved
	SubmissionReviewStatusRejected    = contestentity.SubmissionReviewStatusRejected
)

type Submission = contestentity.Submission
