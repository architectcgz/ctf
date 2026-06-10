package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (r *Repository) Create(ctx context.Context, contest *contestentity.Contest) error {
	return r.dbWithContext(ctx).Create(contest).Error
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*contestentity.Contest, error) {
	var contest contestentity.Contest
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&contest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contestdomain.ErrContestNotFound
		}
		return nil, err
	}
	return &contest, nil
}

func (r *Repository) Update(ctx context.Context, contest *contestentity.Contest) error {
	if contest == nil {
		return contestdomain.ErrContestNotFound
	}
	updatedAt := time.Now().UTC()
	contest.UpdatedAt = updatedAt

	updateResult := r.dbWithContext(ctx).
		Model(&contestentity.Contest{}).
		Where("id = ? AND deleted_at IS NULL", contest.ID).
		Updates(contestUpdateFields(contest, updatedAt))
	if updateResult.Error != nil {
		return updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return contestdomain.ErrContestNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context, filter contestports.ContestListFilter, offset, limit int) ([]*contestentity.Contest, int64, error) {
	var contests []*contestentity.Contest
	var total int64

	query := applyContestListFilter(r.dbWithContext(ctx).Model(&contestentity.Contest{}), filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortColumn := contestListSortColumn(contestports.ContestListFilterSort(filter))
	sortDirection := contestListSortDirection(contestports.ContestListFilterSort(filter))
	err := query.
		Order(fmt.Sprintf("%s %s", sortColumn, sortDirection)).
		Order(fmt.Sprintf("id %s", sortDirection)).
		Offset(offset).
		Limit(limit).
		Find(&contests).Error
	return contests, total, err
}

func (r *Repository) Summarize(ctx context.Context, filter contestports.ContestListFilter) (contestports.ContestListSummary, error) {
	type contestStatusCountRow struct {
		Status string
		Count  int64
	}

	var rows []contestStatusCountRow
	query := applyContestListFilter(r.dbWithContext(ctx).Model(&contestentity.Contest{}), filter)
	if err := query.Select("status, COUNT(*) AS count").Group("status").Scan(&rows).Error; err != nil {
		return contestports.ContestListSummary{}, err
	}

	summary := contestports.ContestListSummary{}
	for _, row := range rows {
		switch row.Status {
		case contestentity.ContestStatusDraft:
			summary.DraftCount = row.Count
		case contestentity.ContestStatusRegistration:
			summary.RegistrationCount = row.Count
		case contestentity.ContestStatusRunning:
			summary.RunningCount = row.Count
		case contestentity.ContestStatusFrozen:
			summary.FrozenCount = row.Count
		case contestentity.ContestStatusEnded:
			summary.EndedCount = row.Count
		}
	}
	return summary, nil
}

func contestListSortColumn(sort contestports.ContestListSort) string {
	if contestports.ContestListSortIsStartTime(sort) {
		return "start_time"
	}
	return "created_at"
}

func contestListSortDirection(sort contestports.ContestListSort) string {
	if contestports.ContestListSortIsAsc(sort) {
		return "ASC"
	}
	return "DESC"
}

func applyContestListFilter(query *gorm.DB, filter contestports.ContestListFilter) *gorm.DB {
	if statuses := contestports.ContestListFilterStatuses(filter); len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if mode := contestports.ContestListFilterMode(filter); mode != nil {
		query = query.Where("mode = ?", *mode)
	}
	return query
}
