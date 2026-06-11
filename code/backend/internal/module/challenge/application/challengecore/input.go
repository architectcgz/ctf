package challengecore

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
	ImageID         *int64
	AttachmentURL   string
	InstanceSharing string
	Hints           []ChallengeHintInput
}

type OptionalImageIDInput struct {
	Set   bool
	Value *int64
}

type UpdateChallengeInput struct {
	Title           string
	Description     string
	Category        string
	Difficulty      string
	Points          int
	ImageID         OptionalImageIDInput
	AttachmentURL   *string
	InstanceSharing string
	Hints           []ChallengeHintInput
}
