package contracts

import "time"

const (
	SubmissionStatusCorrect       = "correct"
	SubmissionStatusIncorrect     = "incorrect"
	SubmissionStatusPendingReview = "pending_review"
)

type SubmissionResp struct {
	IsCorrect          bool       `json:"is_correct"`
	Status             string     `json:"status"`
	Message            string     `json:"message,omitempty"`
	Points             int        `json:"points,omitempty"`
	SubmittedAt        time.Time  `json:"submitted_at"`
	InstanceShutdownAt *time.Time `json:"instance_shutdown_at,omitempty"`
}

type ChallengeSubmissionRecordResp struct {
	ID          int64     `json:"id"`
	Status      string    `json:"status"`
	Answer      string    `json:"answer,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
}
