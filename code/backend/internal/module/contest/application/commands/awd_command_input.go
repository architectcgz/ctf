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

type SubmitAttackInput struct {
	VictimTeamID int64
	Flag         string
}

type PreviewCheckerInput struct {
	AWDChallengeID   int64
	ServiceID        int64
	CheckerType      string
	CheckerConfig    map[string]any
	AccessURL        string
	PreviewFlag      string
	PreviewRequestID string
}

type CreateContestAWDServiceInput struct {
	AWDChallengeID         int64
	Points                 int
	DisplayName            string
	Order                  int
	IsVisible              *bool
	CheckerType            *string
	CheckerConfig          map[string]any
	AWDSLAScore            *int
	AWDDefenseScore        *int
	AWDCheckerPreviewToken *string
}

type UpdateContestAWDServiceInput struct {
	AWDChallengeID         *int64
	Points                 *int
	DisplayName            *string
	Order                  *int
	IsVisible              *bool
	CheckerType            *string
	CheckerConfig          map[string]any
	AWDSLAScore            *int
	AWDDefenseScore        *int
	AWDCheckerPreviewToken *string
}
