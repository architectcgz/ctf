package commands

import "ctf-platform/internal/model"

type ChallengeHintInput struct {
	Level   int
	Title   string
	Content string
}

type CreateChallengeInput struct {
	Title           string
	Description     string
	Category        string
	Difficulty      string
	Points          int
	ImageID         int64
	AttachmentURL   string
	InstanceSharing model.InstanceSharing
	Hints           []ChallengeHintInput
}

type UpdateChallengeInput struct {
	Title           string
	Description     string
	Category        string
	Difficulty      string
	Points          int
	ImageID         *int64
	AttachmentURL   *string
	InstanceSharing model.InstanceSharing
	Hints           []ChallengeHintInput
}
