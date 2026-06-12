package contracts

import "time"

const (
	EventPublishCheckFinished               = "challenge.publish_check_finished"
	EventPublishCheckFinishedPayloadVersion = 1
	EventPublishedCatalogChanged            = "challenge.published_catalog_changed"
	ChallengeCatalogChangeTypePublished     = "published"
	ChallengeCatalogChangeTypeUpdated       = "updated"
	ChallengeCatalogChangeTypeDeleted       = "deleted"
	ChallengeCatalogChangeTypeImported      = "imported"
)

type PublishCheckFinishedEvent struct {
	UserID         int64     `json:"user_id"`
	ChallengeID    int64     `json:"challenge_id"`
	ChallengeTitle string    `json:"challenge_title"`
	Passed         bool      `json:"passed"`
	FailureSummary string    `json:"failure_summary,omitempty"`
	OccurredAt     time.Time `json:"occurred_at"`
}

type PublishedCatalogChangedEvent struct {
	ChallengeID      int64
	ChangeType       string
	PreviousStatus   string
	CurrentStatus    string
	PreviousCategory string
	CurrentCategory  string
	PreviousPoints   int
	CurrentPoints    int
}
