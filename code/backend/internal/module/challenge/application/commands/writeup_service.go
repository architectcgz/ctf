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

func (s *WriteupService) Upsert(challengeID, actorUserID int64, req *dto.UpsertChallengeWriteupReq) (*dto.AdminChallengeWriteupResp, error) {
	return s.UpsertWithContext(context.Background(), challengeID, actorUserID, req)
}

func (s *WriteupService) UpsertWithContext(ctx context.Context, challengeID, actorUserID int64, req *dto.UpsertChallengeWriteupReq) (*dto.AdminChallengeWriteupResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := s.repo.FindByIDWithContext(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	existing, err := s.repo.FindWriteupByChallengeIDWithContext(ctx, challengeID)
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
	if err := s.repo.UpsertWriteupWithContext(ctx, writeup); err != nil {
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.AdminWriteupRespFromModel(item), nil
}

func (s *WriteupService) Delete(challengeID int64) error {
	return s.DeleteWithContext(context.Background(), challengeID)
}

func (s *WriteupService) DeleteWithContext(ctx context.Context, challengeID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := s.repo.FindByIDWithContext(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}
	return s.repo.DeleteWriteupByChallengeIDWithContext(ctx, challengeID)
}

func (s *WriteupService) UpsertSubmission(challengeID, actorUserID int64, req *dto.UpsertSubmissionWriteupReq) (*dto.SubmissionWriteupResp, error) {
	return s.UpsertSubmissionWithContext(context.Background(), challengeID, actorUserID, req)
}

func (s *WriteupService) UpsertSubmissionWithContext(ctx context.Context, challengeID, actorUserID int64, req *dto.UpsertSubmissionWriteupReq) (*dto.SubmissionWriteupResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	challengeItem, err := s.repo.FindByIDWithContext(ctx, challengeID)
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

	existing, err := s.repo.FindSubmissionWriteupByUserChallengeWithContext(ctx, actorUserID, challengeID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if submissionStatus == model.SubmissionWriteupStatusPublished {
		isSolved, solveErr := s.repo.GetSolvedStatusWithContext(ctx, actorUserID, challengeID)
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

	if err := s.repo.UpsertSubmissionWriteupWithContext(ctx, writeup); err != nil {
		return nil, err
	}
	item, err := s.repo.FindSubmissionWriteupByUserChallengeWithContext(ctx, actorUserID, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.SubmissionWriteupRespFromModel(item), nil
}

func (s *WriteupService) RecommendOfficial(challengeID, actorUserID int64) (*dto.AdminChallengeWriteupResp, error) {
	return s.RecommendOfficialWithContext(context.Background(), challengeID, actorUserID)
}

func (s *WriteupService) RecommendOfficialWithContext(ctx context.Context, challengeID, actorUserID int64) (*dto.AdminChallengeWriteupResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	item, err := s.loadOfficialWriteupForModerationWithContext(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	item.IsRecommended = true
	item.RecommendedAt = &now
	item.RecommendedBy = &actorUserID
	item.UpdatedAt = now

	if err := s.repo.UpsertWriteupWithContext(ctx, item); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindWriteupByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.AdminWriteupRespFromModel(updated), nil
}

func (s *WriteupService) UnrecommendOfficial(challengeID, actorUserID int64) (*dto.AdminChallengeWriteupResp, error) {
	return s.UnrecommendOfficialWithContext(context.Background(), challengeID, actorUserID)
}

func (s *WriteupService) UnrecommendOfficialWithContext(ctx context.Context, challengeID, _ int64) (*dto.AdminChallengeWriteupResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	item, err := s.loadOfficialWriteupForModerationWithContext(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	item.IsRecommended = false
	item.RecommendedAt = nil
	item.RecommendedBy = nil
	item.UpdatedAt = time.Now()

	if err := s.repo.UpsertWriteupWithContext(ctx, item); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindWriteupByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.AdminWriteupRespFromModel(updated), nil
}

func (s *WriteupService) RecommendCommunity(submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(submissionID, requesterID, requesterRole)
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

	if err := s.repo.UpsertSubmissionWriteup(record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(submissionID)
	if err != nil {
		return nil, err
	}
	return domain.SubmissionWriteupRespFromModel(updated), nil
}

func (s *WriteupService) UnrecommendCommunity(submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(submissionID, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}

	record.IsRecommended = false
	record.RecommendedAt = nil
	record.RecommendedBy = nil
	record.UpdatedAt = time.Now()

	if err := s.repo.UpsertSubmissionWriteup(record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(submissionID)
	if err != nil {
		return nil, err
	}
	return domain.SubmissionWriteupRespFromModel(updated), nil
}

func (s *WriteupService) HideCommunity(submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(submissionID, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}

	record.VisibilityStatus = model.SubmissionWriteupVisibilityHidden
	record.IsRecommended = false
	record.RecommendedAt = nil
	record.RecommendedBy = nil
	record.UpdatedAt = time.Now()

	if err := s.repo.UpsertSubmissionWriteup(record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(submissionID)
	if err != nil {
		return nil, err
	}
	return domain.SubmissionWriteupRespFromModel(updated), nil
}

func (s *WriteupService) RestoreCommunity(submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error) {
	record, err := s.loadCommunityWriteupForModeration(submissionID, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}

	record.VisibilityStatus = model.SubmissionWriteupVisibilityVisible
	record.UpdatedAt = time.Now()

	if err := s.repo.UpsertSubmissionWriteup(record); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindSubmissionWriteupByID(submissionID)
	if err != nil {
		return nil, err
	}
	return domain.SubmissionWriteupRespFromModel(updated), nil
}

func (s *WriteupService) loadOfficialWriteupForModeration(challengeID int64) (*model.ChallengeWriteup, error) {
	return s.loadOfficialWriteupForModerationWithContext(context.Background(), challengeID)
}

func (s *WriteupService) loadOfficialWriteupForModerationWithContext(ctx context.Context, challengeID int64) (*model.ChallengeWriteup, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := s.repo.FindByIDWithContext(ctx, challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *WriteupService) loadCommunityWriteupForModeration(submissionID, requesterID int64, requesterRole string) (*model.SubmissionWriteup, error) {
	record, err := s.repo.GetTeacherSubmissionWriteupByID(submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanModerateCommunityWriteup(s.repo, requesterID, requesterRole, record); err != nil {
		return nil, err
	}

	item, err := s.repo.FindSubmissionWriteupByID(submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func ensureTeacherCanModerateCommunityWriteup(
	repo challengeports.ChallengeWriteupRepository,
	requesterID int64,
	requesterRole string,
	record *challengeports.TeacherSubmissionWriteupRecord,
) error {
	if requesterRole == model.RoleAdmin {
		return nil
	}
	requester, err := repo.FindUserByID(requesterID)
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
