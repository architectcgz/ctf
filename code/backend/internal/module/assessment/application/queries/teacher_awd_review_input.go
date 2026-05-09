package queries

type ListTeacherAWDReviewContestsInput struct {
	Status  string
	Keyword string
	Page    int
	Size    int
}

type GetTeacherAWDReviewArchiveInput struct {
	RoundNumber *int
	TeamID      *int64
}
