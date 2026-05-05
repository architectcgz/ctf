package queries

type ListAWDChallengesInput struct {
	Keyword     string
	ServiceType string
	Status      string
	Page        int
	Size        int
}
