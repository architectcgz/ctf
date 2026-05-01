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

type RunCurrentRoundChecksInput struct {
	ForceOverride  *bool
	OverrideReason *string
}

type CreateAttackLogInput struct {
	AttackerTeamID int64
	VictimTeamID   int64
	ServiceID      int64
	AttackType     string
	SubmittedFlag  string
	IsSuccess      bool
}
