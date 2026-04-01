package commands

import (
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
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	if req.Visibility == model.WriteupVisibilityScheduled && req.ReleaseAt == nil {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("定时公开必须提供 release_at"))
	}

	writeup := &model.ChallengeWriteup{
		ChallengeID: challengeID,
		Title:       strings.TrimSpace(req.Title),
		Content:     strings.TrimSpace(req.Content),
		Visibility:  req.Visibility,
		ReleaseAt:   req.ReleaseAt,
		CreatedBy:   &actorUserID,
		UpdatedAt:   time.Now(),
	}
	if writeup.Visibility != model.WriteupVisibilityScheduled {
		writeup.ReleaseAt = nil
	}
	if err := s.repo.UpsertWriteup(writeup); err != nil {
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}
	return domain.AdminWriteupRespFromModel(item), nil
}

func (s *WriteupService) Delete(challengeID int64) error {
	if _, err := s.repo.FindByID(challengeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}
	return s.repo.DeleteWriteupByChallengeID(challengeID)
}

func (s *WriteupService) UpsertSubmission(challengeID, actorUserID int64, req *dto.UpsertSubmissionWriteupReq) (*dto.SubmissionWriteupResp, error) {
	challengeItem, err := s.repo.FindByID(challengeID)
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
	reviewStatus := model.SubmissionWriteupReviewPending
	var submittedAt *time.Time
	if submissionStatus == model.SubmissionWriteupStatusSubmitted {
		submittedAt = &now
	}

	existing, err := s.repo.FindSubmissionWriteupByUserChallenge(actorUserID, challengeID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		if existing.SubmissionStatus == model.SubmissionWriteupStatusSubmitted && submissionStatus == model.SubmissionWriteupStatusDraft {
			submissionStatus = model.SubmissionWriteupStatusSubmitted
		}
		if existing.SubmittedAt != nil && submissionStatus == model.SubmissionWriteupStatusSubmitted {
			submittedAt = existing.SubmittedAt
		}
	}

	writeup := &model.SubmissionWriteup{
		UserID:           actorUserID,
		ChallengeID:      challengeID,
		Title:            strings.TrimSpace(req.Title),
		Content:          strings.TrimSpace(req.Content),
		SubmissionStatus: submissionStatus,
		ReviewStatus:     reviewStatus,
		SubmittedAt:      submittedAt,
		UpdatedAt:        now,
	}
	if existing != nil {
		writeup.ID = existing.ID
		writeup.CreatedAt = existing.CreatedAt
	} else {
		writeup.CreatedAt = now
	}
	if submissionStatus == model.SubmissionWriteupStatusDraft {
		writeup.SubmittedAt = nil
	}

	if err := s.repo.UpsertSubmissionWriteup(writeup); err != nil {
		return nil, err
	}
	item, err := s.repo.FindSubmissionWriteupByUserChallenge(actorUserID, challengeID)
	if err != nil {
		return nil, err
	}
	return domain.SubmissionWriteupRespFromModel(item), nil
}

func (s *WriteupService) ReviewSubmission(submissionID, reviewerID int64, reviewerRole string, req *dto.ReviewSubmissionWriteupReq) (*dto.SubmissionWriteupResp, error) {
	record, err := s.repo.GetTeacherSubmissionWriteupByID(submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessSubmission(s.repo, reviewerID, reviewerRole, record); err != nil {
		return nil, err
	}
	if record.Submission.SubmissionStatus != model.SubmissionWriteupStatusSubmitted {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("草稿状态不能评阅"))
	}

	item := record.Submission
	now := time.Now()
	item.ReviewStatus = req.ReviewStatus
	item.ReviewComment = strings.TrimSpace(req.ReviewComment)
	item.ReviewedBy = &reviewerID
	item.ReviewedAt = &now
	item.UpdatedAt = now

	if err := s.repo.UpsertSubmissionWriteup(&item); err != nil {
		return nil, err
	}
	saved, err := s.repo.FindSubmissionWriteupByID(submissionID)
	if err != nil {
		return nil, err
	}
	return domain.SubmissionWriteupRespFromModel(saved), nil
}

func ensureTeacherCanAccessSubmission(
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
