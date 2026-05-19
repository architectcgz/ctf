package entity

const (
	ChallengeStatusPublished      = "published"
	ChallengeFlagTypeStatic       = "static"
	ChallengeFlagTypeDynamic      = "dynamic"
	ChallengeFlagTypeRegex        = "regex"
	ChallengeFlagTypeManualReview = "manual_review"
)

// Challenge is a contest-facing challenge projection.
// It contains only fields required by contest bounded-context use cases.
type Challenge struct {
	ID         int64
	Title      string
	Category   string
	Difficulty string
	Points     int
	Status     string
	FlagType   string
	FlagPrefix string
}
