package contest

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type ParticipationService struct {
	contestRepo Repository
	repo        *ParticipationRepository
	teamRepo    *TeamRepository
}

func NewParticipationService(contestRepo Repository, repo *ParticipationRepository, teamRepo *TeamRepository) *ParticipationService {
	return &ParticipationService{
		contestRepo: contestRepo,
		repo:        repo,
		teamRepo:    teamRepo,
	}
}

func (s *ParticipationService) RegisterContest(ctx context.Context, contestID, userID int64) error {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if contest.Status != model.ContestStatusRegistration {
		return errcode.ErrContestRegistrationClosed
	}

	var teamID *int64
	team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
	if err == nil && team != nil && team.ID > 0 {
		teamID = &team.ID
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrInternal.WithCause(err)
		}
		registration = &model.ContestRegistration{
			ContestID: contestID,
			UserID:    userID,
			TeamID:    teamID,
			Status:    model.ContestRegistrationStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.repo.CreateRegistration(ctx, registration); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		return nil
	}

	if registration.Status != model.ContestRegistrationStatusApproved {
		registration.Status = model.ContestRegistrationStatusPending
		registration.ReviewedBy = nil
		registration.ReviewedAt = nil
	}
	registration.TeamID = teamID
	registration.UpdatedAt = now
	if err := s.repo.SaveRegistration(ctx, registration); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ParticipationService) ListRegistrations(ctx context.Context, contestID int64, query *dto.ContestRegistrationQuery) (*dto.PageResult, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, ErrContestNotFound) {
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

func (s *ParticipationService) ReviewRegistration(ctx context.Context, contestID, registrationID, reviewerID int64, req *dto.ReviewContestRegistrationReq) (*dto.ContestRegistrationResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	registration, err := s.repo.FindRegistrationByID(ctx, contestID, registrationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrContestRegistrationNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if registration.Status != model.ContestRegistrationStatusPending {
		return nil, errcode.ErrInvalidStatusTransition
	}

	now := time.Now()
	registration.Status = req.Status
	registration.ReviewedBy = &reviewerID
	registration.ReviewedAt = &now
	registration.UpdatedAt = now
	if req.Status == model.ContestRegistrationStatusRejected {
		registration.TeamID = nil
	}
	if err := s.repo.SaveRegistration(ctx, registration); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	user, err := s.repo.FindUserByID(ctx, registration.UserID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.ContestRegistrationResp{
		ID:         registration.ID,
		ContestID:  registration.ContestID,
		UserID:     registration.UserID,
		Username:   user.Username,
		TeamID:     registration.TeamID,
		Status:     registration.Status,
		ReviewedBy: registration.ReviewedBy,
		ReviewedAt: registration.ReviewedAt,
		CreatedAt:  registration.CreatedAt,
		UpdatedAt:  registration.UpdatedAt,
	}, nil
}

func (s *ParticipationService) ListAnnouncements(ctx context.Context, contestID int64) ([]*dto.ContestAnnouncementResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, ErrContestNotFound) {
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

func (s *ParticipationService) CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req *dto.CreateContestAnnouncementReq) (*dto.ContestAnnouncementResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	item := &model.ContestAnnouncement{
		ContestID: contestID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedBy: &actorUserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateAnnouncement(ctx, item); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.ContestAnnouncementResp{
		ID:        item.ID,
		Title:     item.Title,
		Content:   item.Content,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *ParticipationService) DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) error {
	deleted, err := s.repo.DeleteAnnouncement(ctx, contestID, announcementID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !deleted {
		return errcode.ErrContestAnnouncementNotFound
	}
	return nil
}

func (s *ParticipationService) GetMyProgress(ctx context.Context, contestID, userID int64) (*dto.ContestMyProgressResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var teamID *int64
	if registration, err := s.repo.FindRegistration(ctx, contestID, userID); err == nil {
		teamID = registration.TeamID
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if teamID == nil {
		team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
		if err == nil && team != nil && team.ID > 0 {
			teamID = &team.ID
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
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
