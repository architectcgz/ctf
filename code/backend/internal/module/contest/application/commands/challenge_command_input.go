package commands

type AddContestChallengeInput struct {
	ChallengeID int64
	Points      int
	Order       int
	IsVisible   *bool
}

type UpdateContestChallengeInput struct {
	Points    *int
	Order     *int
	IsVisible *bool
}
