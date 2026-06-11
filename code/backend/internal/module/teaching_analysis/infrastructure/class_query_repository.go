package infrastructure

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	queryports "ctf-platform/internal/module/teaching_analysis/ports"
)

type ClassQueryRepository struct {
	db *gorm.DB
}

func NewClassQueryRepository(db *gorm.DB) *ClassQueryRepository {
	return &ClassQueryRepository{db: db}
}

func (r *ClassQueryRepository) CountStudentsByClass(ctx context.Context, className string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&identitycontracts.User{}).
		Where("class_name = ? AND role = ? AND deleted_at IS NULL", className, identitycontracts.RoleStudent).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count students by class: %w", err)
	}
	return count, nil
}

func (r *ClassQueryRepository) CountClasses(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&identitycontracts.User{}).
		Distinct("class_name").
		Where("role = ? AND class_name <> '' AND deleted_at IS NULL", identitycontracts.RoleStudent).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count classes: %w", err)
	}
	return count, nil
}

func (r *ClassQueryRepository) ListClasses(ctx context.Context, offset, limit int) ([]queryports.ClassItem, error) {
	rows := make([]classItemRow, 0)
	query := r.db.WithContext(ctx).Model(&identitycontracts.User{}).
		Select("class_name AS name, COUNT(*) AS student_count").
		Where("role = ? AND class_name <> '' AND deleted_at IS NULL", identitycontracts.RoleStudent).
		Group("class_name").
		Order("class_name ASC")
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	return toClassItems(rows), nil
}
