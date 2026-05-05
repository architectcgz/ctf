package commands

type CreateAWDChallengeInput struct {
	Name           string
	Slug           string
	Category       string
	Difficulty     string
	Description    string
	ServiceType    string
	DeploymentMode string
}

type UpdateAWDChallengeInput struct {
	Name           string
	Slug           string
	Category       string
	Difficulty     string
	Description    string
	ServiceType    string
	DeploymentMode string
	Status         string
}
