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
	db          *gorm.DB
	contestRepo Repository
	teamRepo    *TeamRepository
}

func NewParticipationService(db *gorm.DB, contestRepo Repository, teamRepo *TeamRepository) *ParticipationService {
	return &ParticipationService{
		db:          db,
		contestRepo: contestRepo,
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
	var registration model.ContestRegistration
	if err := s.db.WithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrInternal.WithCause(err)
		}
		registration = model.ContestRegistration{
			ContestID: contestID,
			UserID:    userID,
			TeamID:    teamID,
			Status:    model.ContestRegistrationStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.db.WithContext(ctx).Create(&registration).Error; err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
		return nil
	}

	updates := map[string]any{
		"team_id":    teamID,
		"updated_at": now,
	}
	if registration.Status != model.ContestRegistrationStatusApproved {
		updates["status"] = model.ContestRegistrationStatusPending
		updates["reviewed_by"] = nil
		updates["reviewed_at"] = nil
	}
	if err := s.db.WithContext(ctx).Model(&registration).Updates(updates).Error; err != nil {
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

	type registrationRow struct {
		ID         int64
		ContestID  int64
		UserID     int64
		Username   string
		TeamID     *int64
		Status     string
		ReviewedBy *int64
		ReviewedAt *time.Time
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	baseQuery := s.db.WithContext(ctx).
		Table("contest_registrations AS cr").
		Joins("JOIN users u ON u.id = cr.user_id").
		Where("cr.contest_id = ?", contestID)
	if query.Status != nil {
		baseQuery = baseQuery.Where("cr.status = ?", *query.Status)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var rows []*registrationRow
	if err := baseQuery.
		Select("cr.id, cr.contest_id, cr.user_id, u.username, cr.team_id, cr.status, cr.reviewed_by, cr.reviewed_at, cr.created_at, cr.updated_at").
		Order("cr.created_at ASC, cr.id ASC").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error; err != nil {
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

	var registration model.ContestRegistration
	if err := s.db.WithContext(ctx).
		Where("id = ? AND contest_id = ?", registrationID, contestID).
		First(&registration).Error; err != nil {
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
	if err := s.db.WithContext(ctx).Model(&registration).Updates(map[string]any{
		"status":      registration.Status,
		"team_id":     registration.TeamID,
		"reviewed_by": registration.ReviewedBy,
		"reviewed_at": registration.ReviewedAt,
		"updated_at":  registration.UpdatedAt,
	}).Error; err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var user model.User
	if err := s.db.WithContext(ctx).Select("id, username").First(&user, registration.UserID).Error; err != nil {
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

	var announcements []*model.ContestAnnouncement
	if err := s.db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("created_at DESC, id DESC").
		Find(&announcements).Error; err != nil {
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
	if err := s.db.WithContext(ctx).Create(item).Error; err != nil {
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
	result := s.db.WithContext(ctx).
		Where("id = ? AND contest_id = ?", announcementID, contestID).
		Delete(&model.ContestAnnouncement{})
	if result.Error != nil {
		return errcode.ErrInternal.WithCause(result.Error)
	}
	if result.RowsAffected == 0 {
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

	var registration model.ContestRegistration
	var teamID *int64
	if err := s.db.WithContext(ctx).Where("contest_id = ? AND user_id = ?", contestID, userID).First(&registration).Error; err == nil {
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

	type solvedRow struct {
		ContestChallengeID int64
		SolvedAt           time.Time
		PointsEarned       int
	}
	var rows []*solvedRow
	if err := s.db.WithContext(ctx).
		Table("submissions AS s").
		Select("cc.id AS contest_challenge_id, s.submitted_at AS solved_at, s.score AS points_earned").
		Joins("JOIN contest_challenges cc ON cc.contest_id = s.contest_id AND cc.challenge_id = s.challenge_id").
		Where("s.contest_id = ? AND s.user_id = ? AND s.is_correct = ?", contestID, userID, true).
		Order("s.submitted_at ASC, s.id ASC").
		Scan(&rows).Error; err != nil {
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
