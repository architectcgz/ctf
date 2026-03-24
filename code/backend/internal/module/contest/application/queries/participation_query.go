package queries

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) ListRegistrations(ctx context.Context, contestID int64, query *dto.ContestRegistrationQuery) (*dto.PageResult, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
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

	items := make([]*dto.ContestRegistrationResp, 0, len(rows))
	for _, row := range rows {
		items = append(items, &dto.ContestRegistrationResp{
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

	return &dto.PageResult{
		List:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *ParticipationService) ListAnnouncements(ctx context.Context, contestID int64) ([]*dto.ContestAnnouncementResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	announcements, err := s.repo.ListAnnouncements(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.ContestAnnouncementResp, 0, len(announcements))
	for _, item := range announcements {
		result = append(result, &dto.ContestAnnouncementResp{
			ID:        item.ID,
			Title:     item.Title,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
		})
	}
	return result, nil
}

func (s *ParticipationService) GetMyProgress(ctx context.Context, contestID, userID int64) (*dto.ContestMyProgressResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamID, err := s.resolveUserTeamID(ctx, contestID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListSolvedProgress(ctx, contestID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := &dto.ContestMyProgressResp{
		ContestID: contestID,
		TeamID:    teamID,
		Solved:    make([]*dto.ContestSolvedProgressItem, 0, len(rows)),
	}
	for _, row := range rows {
		result.Solved = append(result.Solved, &dto.ContestSolvedProgressItem{
			ContestChallengeID: row.ContestChallengeID,
			SolvedAt:           row.SolvedAt,
			PointsEarned:       row.PointsEarned,
		})
	}
	return result, nil
}

func (s *ParticipationService) resolveUserTeamID(ctx context.Context, contestID, userID int64) (*int64, error) {
	if registration, err := s.repo.FindRegistration(ctx, contestID, userID); err == nil {
		return registration.TeamID, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
	if err == nil && team != nil && team.ID > 0 {
		return &team.ID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return nil, nil
}
