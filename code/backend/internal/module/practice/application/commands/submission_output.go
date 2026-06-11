package commands

import practicecontracts "ctf-platform/internal/module/practice/contracts"

const (
	SubmissionStatusCorrect       = practicecontracts.SubmissionStatusCorrect
	SubmissionStatusIncorrect     = practicecontracts.SubmissionStatusIncorrect
	SubmissionStatusPendingReview = practicecontracts.SubmissionStatusPendingReview
)

type SubmissionResp = practicecontracts.SubmissionResp
type ChallengeSubmissionRecordResp = practicecontracts.ChallengeSubmissionRecordResp
