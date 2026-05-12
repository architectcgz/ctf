package contracts

const (
	EventPublishCheckFinished = "challenge.publish_check_finished"
)

type PublishCheckFinishedEvent struct {
	UserID         int64
	ChallengeID    int64
	ChallengeTitle string
	Passed         bool
	FailureSummary string
}
