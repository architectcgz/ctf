package commands

type CreateAWDRoundInput struct {
	RoundNumber    int
	Status         *string
	AttackScore    *int
	DefenseScore   *int
	ForceOverride  *bool
	OverrideReason *string
}

type UpsertServiceCheckInput struct {
	TeamID        int64
	ServiceID     int64
	ServiceStatus string
	CheckResult   map[string]any
}
