package commands

type CreateAWDRoundInput struct {
	RoundNumber    int
	Status         *string
	AttackScore    *int
	DefenseScore   *int
	ForceOverride  *bool
	OverrideReason *string
}
