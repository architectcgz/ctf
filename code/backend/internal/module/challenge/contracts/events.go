package contracts

const (
	EventPublishCheckFinished           = "challenge.publish_check_finished"
	EventPublishedCatalogChanged        = "challenge.published_catalog_changed"
	ChallengeCatalogChangeTypePublished = "published"
	ChallengeCatalogChangeTypeUpdated   = "updated"
	ChallengeCatalogChangeTypeDeleted   = "deleted"
	ChallengeCatalogChangeTypeImported  = "imported"
)

type PublishCheckFinishedEvent struct {
	UserID         int64
	ChallengeID    int64
	ChallengeTitle string
	Passed         bool
	FailureSummary string
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
