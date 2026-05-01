package queries

type ContestChallengeInfoResult struct {
	ID             int64
	ChallengeID    int64
	AWDChallengeID *int64
	AWDServiceID   *int64
	Title          string
	Category       string
	Difficulty     string
	Points         int
	Order          int
	SolvedCount    int64
	IsSolved       bool
}
