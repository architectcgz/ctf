package infrastructure

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

const (
	persistedContestModeAWD = "awd"

	persistedContestRegistrationStatusApproved = "approved"

	awdScopeControlScopeTeam        = "team"
	awdScopeControlScopeTeamService = "team_service"

	awdScopeControlTypeRetired         = "retired"
	awdScopeControlTypeServiceDisabled = "service_disabled"
)

type contestAWDServiceSnapshotReadModel struct {
	Name       string         `json:"name"`
	Category   string         `json:"category"`
	Difficulty string         `json:"difficulty"`
	FlagConfig map[string]any `json:"flag_config,omitempty"`
}

func decodeContestAWDServiceSnapshotReadModel(raw string) (contestAWDServiceSnapshotReadModel, error) {
	if raw == "" {
		return contestAWDServiceSnapshotReadModel{}, nil
	}
	var snapshot contestAWDServiceSnapshotReadModel
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return contestAWDServiceSnapshotReadModel{}, err
	}
	if snapshot.FlagConfig == nil {
		snapshot.FlagConfig = map[string]any{}
	}
	return snapshot, nil
}

func joinAWDActiveScopeControls(query *gorm.DB, contestExpr, teamExpr, serviceExpr, retiredAlias, disabledAlias string) *gorm.DB {
	retiredJoin := fmt.Sprintf(
		"LEFT JOIN awd_scope_controls AS %[1]s ON %[1]s.contest_id = %[2]s AND %[1]s.team_id = %[3]s AND %[1]s.scope_type = ? AND %[1]s.service_id = 0 AND %[1]s.control_type = ?",
		retiredAlias, contestExpr, teamExpr,
	)
	disabledJoin := fmt.Sprintf(
		"LEFT JOIN awd_scope_controls AS %[1]s ON %[1]s.contest_id = %[2]s AND %[1]s.team_id = %[3]s AND %[1]s.scope_type = ? AND %[1]s.service_id = %[4]s AND %[1]s.control_type = ?",
		disabledAlias, contestExpr, teamExpr, serviceExpr,
	)
	return query.
		Joins(retiredJoin, awdScopeControlScopeTeam, awdScopeControlTypeRetired).
		Joins(disabledJoin, awdScopeControlScopeTeamService, awdScopeControlTypeServiceDisabled)
}

func applyAWDActiveScopeFilter(query *gorm.DB, serviceExpr, retiredAlias, disabledAlias string) *gorm.DB {
	return query.Where(
		fmt.Sprintf("(%s IS NULL OR (%s.id IS NULL AND %s.id IS NULL))", serviceExpr, retiredAlias, disabledAlias),
	)
}
