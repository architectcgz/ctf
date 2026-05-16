package infrastructure

import (
	"fmt"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

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
		Joins(retiredJoin, model.AWDScopeControlScopeTeam, model.AWDScopeControlTypeRetired).
		Joins(disabledJoin, model.AWDScopeControlScopeTeamService, model.AWDScopeControlTypeServiceDisabled)
}

func applyAWDActiveScopeFilter(query *gorm.DB, serviceExpr, retiredAlias, disabledAlias string) *gorm.DB {
	return query.Where(
		fmt.Sprintf("(%s IS NULL OR (%s.id IS NULL AND %s.id IS NULL))", serviceExpr, retiredAlias, disabledAlias),
	)
}
