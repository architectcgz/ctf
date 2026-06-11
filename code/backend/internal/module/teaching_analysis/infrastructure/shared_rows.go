package infrastructure

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	queryports "ctf-platform/internal/module/teaching_analysis/ports"
)

type classItemRow struct {
	Name         string `gorm:"column:name"`
	StudentCount int64  `gorm:"column:student_count"`
}

type studentItemRow struct {
	ID               int64   `gorm:"column:id"`
	Username         string  `gorm:"column:username"`
	StudentNo        *string `gorm:"column:student_no"`
	Name             *string `gorm:"column:name"`
	ClassName        *string `gorm:"column:class_name"`
	SolvedCount      int     `gorm:"column:solved_count"`
	TotalScore       int     `gorm:"column:total_score"`
	RecentEventCount int     `gorm:"column:recent_event_count"`
	WeakDimension    *string `gorm:"column:weak_dimension"`
}

type progressRow struct {
	Key    string `gorm:"column:key"`
	Total  int    `gorm:"column:total"`
	Solved int    `gorm:"column:solved"`
}

func toClassItems(rows []classItemRow) []queryports.ClassItem {
	items := make([]queryports.ClassItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, queryports.ClassItem{
			Name:         row.Name,
			StudentCount: row.StudentCount,
		})
	}
	return items
}

func toStudentItems(rows []studentItemRow) []queryports.StudentItem {
	items := make([]queryports.StudentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, queryports.StudentItem{
			ID:               row.ID,
			Username:         row.Username,
			StudentNo:        row.StudentNo,
			Name:             row.Name,
			ClassName:        row.ClassName,
			SolvedCount:      row.SolvedCount,
			TotalScore:       row.TotalScore,
			RecentEventCount: row.RecentEventCount,
			WeakDimension:    row.WeakDimension,
		})
	}
	return items
}

func toProgressRows(rows []progressRow) []queryports.ProgressRow {
	items := make([]queryports.ProgressRow, 0, len(rows))
	for _, row := range rows {
		items = append(items, queryports.ProgressRow{
			Key:    row.Key,
			Total:  row.Total,
			Solved: row.Solved,
		})
	}
	return items
}

func listStudentsBaseQuery(db *gorm.DB, ctx context.Context, since time.Time) *gorm.DB {
	return db.WithContext(ctx).Table("users AS u").
		Select(`
			u.id,
			u.username,
			NULLIF(u.name, '') AS name,
			NULLIF(u.student_no, '') AS student_no,
			NULLIF(u.class_name, '') AS class_name,
			COALESCE((
				SELECT COUNT(DISTINCT s.challenge_id)
				FROM submissions s
				JOIN challenges c ON c.id = s.challenge_id
				WHERE s.user_id = u.id AND s.is_correct = TRUE AND c.status = ?
			), 0) AS solved_count,
			COALESCE((
				SELECT SUM(c.points)
				FROM submissions s
				JOIN challenges c ON c.id = s.challenge_id
				WHERE s.user_id = u.id AND s.is_correct = TRUE AND c.status = ?
			), 0) AS total_score,
			COALESCE((
				SELECT COUNT(*)
				FROM (
					SELECT s.id
					FROM submissions s
					WHERE s.user_id = u.id AND s.submitted_at >= ?
					UNION ALL
					SELECT i.id
					FROM instances i
					WHERE i.user_id = u.id AND i.created_at >= ?
					UNION ALL
					SELECT i.id
					FROM instances i
					WHERE i.user_id = u.id AND i.status IN ('stopped', 'expired') AND i.updated_at >= ?
				) recent_events
			), 0) AS recent_event_count,
			(
				SELECT sp.dimension
				FROM skill_profiles sp
				WHERE sp.user_id = u.id
				ORDER BY sp.score ASC, sp.updated_at DESC
				LIMIT 1
			) AS weak_dimension
		`, challengecontracts.ChallengeStatusPublished, challengecontracts.ChallengeStatusPublished, since, since, since).
		Where("u.role = ? AND u.deleted_at IS NULL", identitycontracts.RoleStudent)
}

func applyStudentFilters(query *gorm.DB, className, keyword, studentNo string) *gorm.DB {
	if className != "" {
		query = query.Where("u.class_name = ?", className)
	}
	if keyword != "" {
		likeKeyword := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("(LOWER(u.username) LIKE ? OR LOWER(u.name) LIKE ?)", likeKeyword, likeKeyword)
	}
	if studentNo != "" {
		query = query.Where("u.student_no = ?", studentNo)
	}
	return query
}

func normalizeClassScope(classNames []string) []string {
	normalized := make([]string, 0, len(classNames))
	seen := make(map[string]struct{}, len(classNames))
	for _, className := range classNames {
		name := strings.TrimSpace(className)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	sort.Strings(normalized)
	return normalized
}

func applyStudentScopeFilter(query *gorm.DB, classNames []string) *gorm.DB {
	normalized := normalizeClassScope(classNames)
	if len(normalized) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("u.class_name IN ?", normalized)
}

func resolveStudentOrder(sortKey, sortOrder string) string {
	direction := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		direction = "ASC"
	}

	switch sortKey {
	case "name":
		return fmt.Sprintf("COALESCE(NULLIF(u.name, ''), u.username) %s, u.username ASC", direction)
	case "student_no":
		return fmt.Sprintf("COALESCE(NULLIF(u.student_no, ''), u.username) %s, u.username ASC", direction)
	case "total_score":
		return fmt.Sprintf("total_score %s, solved_count DESC, COALESCE(NULLIF(u.student_no, ''), u.username) ASC, u.username ASC", direction)
	case "solved_count":
		fallthrough
	default:
		return fmt.Sprintf("solved_count %s, total_score DESC, COALESCE(NULLIF(u.student_no, ''), u.username) ASC, u.username ASC", direction)
	}
}
