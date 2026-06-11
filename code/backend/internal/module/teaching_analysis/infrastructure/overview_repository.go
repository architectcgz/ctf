package infrastructure

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	queryports "ctf-platform/internal/module/teaching_analysis/ports"
)

type OverviewRepository struct {
	db *gorm.DB
}

func NewOverviewRepository(db *gorm.DB) *OverviewRepository {
	return &OverviewRepository{db: db}
}

func (r *OverviewRepository) ListStudentsByClasses(
	ctx context.Context,
	classNames []string,
	keyword, studentNo string,
	since time.Time,
) ([]queryports.StudentItem, error) {
	normalized := normalizeClassScope(classNames)
	if len(normalized) == 0 {
		return []queryports.StudentItem{}, nil
	}

	rows := make([]studentItemRow, 0)
	query := applyStudentFilters(
		applyStudentScopeFilter(listStudentsBaseQuery(r.db, ctx, since), normalized),
		"",
		keyword,
		studentNo,
	).Order(resolveStudentOrder("solved_count", "desc"))
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list students by classes: %w", err)
	}
	return toStudentItems(rows), nil
}

func (r *OverviewRepository) GetOverviewTrend(
	ctx context.Context,
	classNames []string,
	since time.Time,
	days int,
) (*queryports.OverviewTrend, error) {
	normalized := normalizeClassScope(classNames)
	if days <= 0 {
		days = 7
	}
	if len(normalized) == 0 {
		return &queryports.OverviewTrend{Points: []queryports.OverviewTrendPoint{}}, nil
	}

	type eventRow struct {
		UserID     int64     `gorm:"column:user_id"`
		OccurredAt time.Time `gorm:"column:occurred_at"`
		IsSolve    bool      `gorm:"column:is_solve"`
	}

	rows := make([]eventRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT s.user_id, s.submitted_at AS occurred_at, s.is_correct AS is_solve
		FROM submissions s
		JOIN users u ON u.id = s.user_id
		WHERE u.role = ? AND u.class_name IN ? AND u.deleted_at IS NULL AND s.submitted_at >= ?
		UNION ALL
		SELECT i.user_id, i.created_at AS occurred_at, FALSE AS is_solve
		FROM instances i
		JOIN users u ON u.id = i.user_id
		WHERE u.role = ? AND u.class_name IN ? AND u.deleted_at IS NULL AND i.created_at >= ?
		UNION ALL
		SELECT i.user_id, i.updated_at AS occurred_at, FALSE AS is_solve
		FROM instances i
		JOIN users u ON u.id = i.user_id
		WHERE u.role = ? AND u.class_name IN ? AND u.deleted_at IS NULL
			AND i.status IN ('stopped', 'expired') AND i.updated_at >= ?
	`, identitycontracts.RoleStudent, normalized, since, identitycontracts.RoleStudent, normalized, since, identitycontracts.RoleStudent, normalized, since).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get overview trend: %w", err)
	}

	points := make([]queryports.OverviewTrendPoint, days)
	indexByDate := make(map[string]int, days)
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i).Format("2006-01-02")
		points[i] = queryports.OverviewTrendPoint{Date: date}
		indexByDate[date] = i
	}

	activeUsersByDate := make(map[string]map[int64]struct{}, days)
	for _, row := range rows {
		date := row.OccurredAt.Format("2006-01-02")
		idx, ok := indexByDate[date]
		if !ok {
			continue
		}
		points[idx].EventCount++
		if row.IsSolve {
			points[idx].SolveCount++
		}
		users := activeUsersByDate[date]
		if users == nil {
			users = make(map[int64]struct{})
			activeUsersByDate[date] = users
		}
		users[row.UserID] = struct{}{}
	}

	for i := range points {
		points[i].ActiveStudentCount = int64(len(activeUsersByDate[points[i].Date]))
	}

	return &queryports.OverviewTrend{Points: points}, nil
}
