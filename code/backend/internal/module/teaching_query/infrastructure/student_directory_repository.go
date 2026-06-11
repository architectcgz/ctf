package infrastructure

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
)

type StudentDirectoryRepository struct {
	db *gorm.DB
}

func NewStudentDirectoryRepository(db *gorm.DB) *StudentDirectoryRepository {
	return &StudentDirectoryRepository{db: db}
}

func (r *StudentDirectoryRepository) ListStudents(
	ctx context.Context,
	className, keyword, studentNo, sortKey, sortOrder string,
	since time.Time,
	offset, limit int,
) ([]queryports.StudentItem, int64, error) {
	rows := make([]studentItemRow, 0)
	var total int64
	countQuery := applyStudentFilters(
		r.db.WithContext(ctx).Table("users AS u").Where("u.role = ? AND u.deleted_at IS NULL", identitycontracts.RoleStudent),
		className,
		keyword,
		studentNo,
	)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count students: %w", err)
	}
	if total == 0 {
		return []queryports.StudentItem{}, 0, nil
	}

	query := applyStudentFilters(listStudentsBaseQuery(r.db, ctx, since), className, keyword, studentNo).
		Order(resolveStudentOrder(sortKey, sortOrder))
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list students: %w", err)
	}
	return toStudentItems(rows), total, nil
}

func (r *StudentDirectoryRepository) ListStudentsByClass(ctx context.Context, className, keyword, studentNo string, since time.Time) ([]queryports.StudentItem, error) {
	items, _, err := r.ListStudents(ctx, className, keyword, studentNo, "solved_count", "desc", since, 0, 0)
	if err != nil {
		return nil, err
	}
	return items, nil
}
