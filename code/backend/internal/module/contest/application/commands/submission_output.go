package commands

import "time"

const (
	SubmissionStatusCorrect   = "correct"
	SubmissionStatusIncorrect = "incorrect"
)

type SubmissionResp struct {
	IsCorrect          bool       `json:"is_correct"`
	Status             string     `json:"status"`
	Message            string     `json:"message,omitempty"`
	Points             int        `json:"points,omitempty"`
	SubmittedAt        time.Time  `json:"submitted_at"`
	InstanceShutdownAt *time.Time `json:"instance_shutdown_at,omitempty"`
}
