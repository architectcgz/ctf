package queries

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) ListRegistrations(ctx context.Context, contestID int64, query ContestRegistrationQueryInput) (*RegistrationPageResult[*ContestRegistrationResult], error) {
	if err := s.ensureContestExists(ctx, contestID); err != nil {
		return nil, err
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	size := query.Size
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	rows, total, err := s.repo.ListRegistrations(ctx, contestID, query.Status, (page-1)*size, size)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make([]*ContestRegistrationResult, 0, len(rows))
	for _, row := range rows {
		items = append(items, &ContestRegistrationResult{
			ID:         row.ID,
			ContestID:  row.ContestID,
			UserID:     row.UserID,
			Username:   row.Username,
			TeamID:     row.TeamID,
			Status:     row.Status,
			ReviewedBy: row.ReviewedBy,
			ReviewedAt: row.ReviewedAt,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}

	return &RegistrationPageResult[*ContestRegistrationResult]{
		List:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}
