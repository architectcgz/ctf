package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type WriteupService struct {
	repo challengeports.ChallengeWriteupRepository
}

func NewWriteupService(repo challengeports.ChallengeWriteupRepository) *WriteupService {
	return &WriteupService{repo: repo}
}

func (s *WriteupService) Upsert(ctx context.Context, challengeID, actorUserID int64, req UpsertOfficialWriteupInput) (*dto.AdminChallengeWriteupResp, error) {
	if _, err := s.repo.FindByID(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	existing, err := s.repo.FindWriteupByChallengeID(ctx, challengeID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	writeup := &model.ChallengeWriteup{
		ChallengeID: challengeID,
		Title:       strings.TrimSpace(req.Title),
		Content:     strings.TrimSpace(req.Content),
		Visibility:  req.Visibility,
		CreatedBy:   &actorUserID,
		UpdatedAt:   time.Now(),
	}
	if existing != nil {
		writeup.ID = existing.ID
		writeup.CreatedAt = existing.CreatedAt
		writeup.IsRecommended = existing.IsRecommended
		writeup.RecommendedAt = existing.RecommendedAt
		writeup.RecommendedBy = existing.RecommendedBy
	}
	if err := s.repo.UpsertWriteup(ctx, writeup); err != nil {
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToAdminChallengeWriteupRespPtr(item), nil
}

func (s *WriteupService) Delete(ctx context.Context, challengeID int64) error {
	if _, err := s.repo.FindByID(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}
	return s.repo.DeleteWriteupByChallengeID(ctx, challengeID)
}

func (s *WriteupService) UpsertSubmission(ctx context.Context, challengeID, actorUserID int64, req UpsertSubmissionWriteupInput) (*dto.SubmissionWriteupResp, error) {
	challengeItem, err := s.repo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	now := time.Now()
	submissionStatus := req.SubmissionStatus
	var publishedAt *time.Time

	existing, err := s.repo.FindSubmissionWriteupByUserChallenge(ctx, actorUserID, challengeID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if submissionStatus == model.SubmissionWriteupStatusPublished {
		isSolved, solveErr := s.repo.GetSolvedStatus(ctx, actorUserID, challengeID)
		if solveErr != nil {
			return nil, solveErr
		}
		if !isSolved {
			return nil, errcode.ErrForbidden
		}
		publishedAt = &now
	}
	if existing != nil && existing.PublishedAt != nil && submissionStatus == model.SubmissionWriteupStatusPublished {
		publishedAt = existing.PublishedAt
	}

	writeup := &model.SubmissionWriteup{
		UserID:           actorUserID,
		ChallengeID:      challengeID,
		Title:            strings.TrimSpace(req.Title),
		Content:          strings.TrimSpace(req.Content),
		SubmissionStatus: submissionStatus,
		VisibilityStatus: model.SubmissionWriteupVisibilityVisible,
		PublishedAt:      publishedAt,
		UpdatedAt:        now,
	}
	if existing != nil {
		writeup.ID = existing.ID
		writeup.CreatedAt = existing.CreatedAt
		writeup.VisibilityStatus = existing.VisibilityStatus
		writeup.IsRecommended = existing.IsRecommended
		writeup.RecommendedAt = existing.RecommendedAt
		writeup.RecommendedBy = existing.RecommendedBy
	} else {
		writeup.CreatedAt = now
	}
	if submissionStatus == model.SubmissionWriteupStatusDraft {
		writeup.PublishedAt = nil
	}

	if err := s.repo.UpsertSubmissionWriteup(ctx, writeup); err != nil {
		return nil, err
	}
	item, err := s.repo.FindSubmissionWriteupByUserChallenge(ctx, actorUserID, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToSubmissionWriteupRespPtr(item), nil
}

func (s *WriteupService) RecommendOfficial(ctx context.Context, challengeID, actorUserID int64) (*dto.AdminChallengeWriteupResp, error) {
	item, err := s.loadOfficialWriteupForModeration(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	item.IsRecommended = true
	item.RecommendedAt = &now
	item.RecommendedBy = &actorUserID
	item.UpdatedAt = now

	if err := s.repo.UpsertWriteup(ctx, item); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindWriteupByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToAdminChallengeWriteupRespPtr(updated), nil
}

func (s *WriteupService) UnrecommendOfficial(ctx context.Context, challengeID, _ int64) (*dto.AdminChallengeWriteupResp, error) {
	item, err := s.loadOfficialWriteupForModeration(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	item.IsRecommended = false
	item.RecommendedAt = nil
	item.RecommendedBy = nil
	item.UpdatedAt = time.Now()

	if err := s.repo.UpsertWriteup(ctx, item); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindWriteupByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToAdminChallengeWriteupRespPtr(updated), nil
}

func (s *WriteupService) RecommendCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(ctx, submissionID, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}
	if record.VisibilityStatus == model.SubmissionWriteupVisibilityHidden {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("隐藏题解不能设为推荐"))
	}

	now := time.Now()
	record.IsRecommended = true
	record.RecommendedAt = &now
	record.RecommendedBy = &requesterID
	record.UpdatedAt = now

	if err := s.repo.UpsertSubmissionWriteup(ctx, record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToSubmissionWriteupRespPtr(updated), nil
}

func (s *WriteupService) UnrecommendCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(ctx, submissionID, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}

	record.IsRecommended = false
	record.RecommendedAt = nil
	record.RecommendedBy = nil
	record.UpdatedAt = time.Now()

	if err := s.repo.UpsertSubmissionWriteup(ctx, record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToSubmissionWriteupRespPtr(updated), nil
}

func (s *WriteupService) HideCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(ctx, submissionID, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}

	record.VisibilityStatus = model.SubmissionWriteupVisibilityHidden
	record.IsRecommended = false
	record.RecommendedAt = nil
	record.RecommendedBy = nil
	record.UpdatedAt = time.Now()

	if err := s.repo.UpsertSubmissionWriteup(ctx, record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToSubmissionWriteupRespPtr(updated), nil
}

func (s *WriteupService) RestoreCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(ctx, submissionID, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}

	record.VisibilityStatus = model.SubmissionWriteupVisibilityVisible
	record.UpdatedAt = time.Now()

	if err := s.repo.UpsertSubmissionWriteup(ctx, record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	return domain.ResponseMapper().ToSubmissionWriteupRespPtr(updated), nil
}

func (s *WriteupService) loadOfficialWriteupForModeration(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
	if _, err := s.repo.FindByID(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *WriteupService) loadCommunityWriteupForModeration(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*model.SubmissionWriteup, error) {
	record, err := s.repo.GetTeacherSubmissionWriteupByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanModerateCommunityWriteup(ctx, s.repo, requesterID, requesterRole, record); err != nil {
		return nil, err
	}

	item, err := s.repo.FindSubmissionWriteupByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func ensureTeacherCanModerateCommunityWriteup(
	ctx context.Context,
	repo challengeports.ChallengeWriteupRepository,
	requesterID int64,
	requesterRole string,
	record *challengeports.TeacherSubmissionWriteupRecord,
) error {
	if requesterRole == model.RoleAdmin {
		return nil
	}
	requester, err := repo.FindUserByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrUnauthorized
		}
		return err
	}
	if requester.ClassName == "" || requester.ClassName != record.ClassName {
		return errcode.ErrForbidden
	}
	return nil
}
