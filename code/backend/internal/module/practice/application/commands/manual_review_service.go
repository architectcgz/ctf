package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
)

func (s *Service) ReviewManualReviewSubmission(
	ctx context.Context,
	submissionID, reviewerID int64,
	reviewerRole string,
	req *dto.ReviewManualReviewSubmissionReq,
) (*dto.TeacherManualReviewSubmissionDetailResp, error) {
	if err := ensureManualReviewRequesterRole(reviewerRole); err != nil {
		return nil, err
	}
	if err := ensureManualReviewDecisionStatus(req); err != nil {
		return nil, err
	}
	record, err := s.repo.GetTeacherManualReviewSubmissionByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessManualReviewSubmission(ctx, s.repo, reviewerID, reviewerRole, record); err != nil {
		return nil, err
	}
	if record.Submission.ReviewStatus != model.SubmissionReviewStatusPending {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("仅待审核提交可执行评阅"))
	}

	challengeItem, err := s.challengeRepo.FindByID(ctx, record.Submission.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challengeItem.FlagType != model.FlagTypeManualReview {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("当前提交不属于人工审核题"))
	}

	now := time.Now()
	item := record.Submission
	item.ReviewStatus = req.ReviewStatus
	item.ReviewComment = strings.TrimSpace(req.ReviewComment)
	item.ReviewedBy = &reviewerID
	item.ReviewedAt = &now
	item.UpdatedAt = now
	if req.ReviewStatus == model.SubmissionReviewStatusApproved {
		if _, err := s.repo.FindCorrectSubmission(ctx, item.UserID, item.ChallengeID); err == nil {
			return nil, errcode.ErrAlreadySolved
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		item.IsCorrect = true
		item.Score = challengeItem.Points
	} else {
		item.IsCorrect = false
		item.Score = 0
	}

	if err := s.repo.UpdateSubmission(ctx, &item); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if item.IsCorrect {
		if s.redis != nil {
			if err := s.redis.Del(ctx, constants.UserProgressKey(item.UserID)).Err(); err != nil {
				s.logger.Warn("删除进度缓存失败", zap.Int64("user_id", item.UserID), zap.Error(err))
			}
		}
		s.publishWeakEvent(ctx, platformevents.Event{
			Name: practicecontracts.EventFlagAccepted,
			Payload: practicecontracts.FlagAcceptedEvent{
				UserID:      item.UserID,
				ChallengeID: item.ChallengeID,
				Dimension:   challengeItem.Category,
				Points:      item.Score,
				OccurredAt:  now,
			},
		})
		if s.scoreService != nil {
			s.triggerScoreUpdate(item.UserID)
		}
	}

	return manualReviewDetailRespFromRecord(*record, item), nil
}

func (s *Service) ListTeacherManualReviewSubmissions(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherManualReviewSubmissionQuery,
) (*dto.PageResult[*dto.TeacherManualReviewSubmissionItemResp], error) {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	if query == nil {
		query = &dto.TeacherManualReviewSubmissionQuery{}
	}
	normalized, err := normalizeTeacherManualReviewQuery(ctx, s.repo, requesterID, requesterRole, query)
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListTeacherManualReviewSubmissions(ctx, normalized)
	if err != nil {
		return nil, err
	}

	respItems := make([]*dto.TeacherManualReviewSubmissionItemResp, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, manualReviewListItemRespFromRecord(item))
	}

	return &dto.PageResult[*dto.TeacherManualReviewSubmissionItemResp]{
		List:  respItems,
		Total: total,
		Page:  normalized.Page,
		Size:  normalized.Size,
	}, nil
}

func (s *Service) GetTeacherManualReviewSubmission(
	ctx context.Context,
	submissionID, requesterID int64,
	requesterRole string,
) (*dto.TeacherManualReviewSubmissionDetailResp, error) {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	record, err := s.repo.GetTeacherManualReviewSubmissionByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessManualReviewSubmission(ctx, s.repo, requesterID, requesterRole, record); err != nil {
		return nil, err
	}
	return manualReviewDetailRespFromRecord(*record, record.Submission), nil
}

func (s *Service) ListMyChallengeSubmissions(ctx context.Context, userID, challengeID int64) ([]*dto.ChallengeSubmissionRecordResp, error) {
	challengeItem, err := s.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	items, err := s.repo.ListChallengeSubmissions(ctx, userID, challengeID, 20)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*dto.ChallengeSubmissionRecordResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, challengeSubmissionRecordRespFromModel(item))
	}
	return resp, nil
}

func ensureTeacherCanAccessManualReviewSubmission(
	ctx context.Context,
	repo practiceports.PracticeUserLookupRepository,
	requesterID int64,
	requesterRole string,
	record *practiceports.TeacherManualReviewSubmissionRecord,
) error {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return err
	}
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

func normalizeTeacherManualReviewQuery(
	ctx context.Context,
	repo practiceports.PracticeUserLookupRepository,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherManualReviewSubmissionQuery,
) (*dto.TeacherManualReviewSubmissionQuery, error) {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	if err := ensureManualReviewQuery(query); err != nil {
		return nil, err
	}
	normalized := *query
	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.Size <= 0 {
		normalized.Size = 20
	}
	if requesterRole == model.RoleAdmin {
		return &normalized, nil
	}

	requester, err := repo.FindUserByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUnauthorized
		}
		return nil, err
	}
	if requester.ClassName == "" {
		return nil, errcode.ErrForbidden
	}
	if normalized.ClassName != "" && normalized.ClassName != requester.ClassName {
		return nil, errcode.ErrForbidden
	}
	normalized.ClassName = requester.ClassName
	return &normalized, nil
}

func ensureManualReviewRequesterRole(role string) error {
	if role == model.RoleAdmin || role == model.RoleTeacher {
		return nil
	}
	return errcode.ErrForbidden
}

func ensureManualReviewDecisionStatus(req *dto.ReviewManualReviewSubmissionReq) error {
	if req == nil {
		return errcode.ErrInvalidParams.WithCause(errors.New("评阅请求不能为空"))
	}
	if len([]rune(strings.TrimSpace(req.ReviewComment))) > 4000 {
		return errcode.ErrInvalidParams.WithCause(errors.New("review_comment 不能超过 4000 个字符"))
	}
	if req.ReviewStatus == model.SubmissionReviewStatusApproved || req.ReviewStatus == model.SubmissionReviewStatusRejected {
		return nil
	}
	return errcode.ErrInvalidParams.WithCause(errors.New("review_status 仅支持 approved 或 rejected"))
}

func ensureManualReviewQuery(query *dto.TeacherManualReviewSubmissionQuery) error {
	if query == nil {
		return nil
	}
	if query.StudentID != nil && *query.StudentID <= 0 {
		return errcode.ErrInvalidParams.WithCause(errors.New("student_id 必须大于 0"))
	}
	if query.ChallengeID != nil && *query.ChallengeID <= 0 {
		return errcode.ErrInvalidParams.WithCause(errors.New("challenge_id 必须大于 0"))
	}
	if len([]rune(strings.TrimSpace(query.ClassName))) > 128 {
		return errcode.ErrInvalidParams.WithCause(errors.New("class_name 不能超过 128 个字符"))
	}
	if query.Size > 100 {
		return errcode.ErrInvalidParams.WithCause(errors.New("page_size 不能超过 100"))
	}
	if query.ReviewStatus == "" ||
		query.ReviewStatus == model.SubmissionReviewStatusPending ||
		query.ReviewStatus == model.SubmissionReviewStatusApproved ||
		query.ReviewStatus == model.SubmissionReviewStatusRejected {
		return nil
	}
	return errcode.ErrInvalidParams.WithCause(errors.New("review_status 仅支持 pending、approved 或 rejected"))
}

func manualReviewDetailRespFromRecord(
	record practiceports.TeacherManualReviewSubmissionRecord,
	submission model.Submission,
) *dto.TeacherManualReviewSubmissionDetailResp {
	resp := practiceCommandResponseMapperInst.ToTeacherManualReviewSubmissionDetailRespBase(submission)
	resp.StudentUsername = record.StudentUsername
	resp.StudentName = record.StudentName
	resp.ClassName = record.ClassName
	resp.ChallengeTitle = record.ChallengeTitle
	resp.Answer = submission.Flag
	resp.ReviewerName = record.ReviewerName
	return resp
}

func manualReviewListItemRespFromRecord(record practiceports.TeacherManualReviewSubmissionRecord) *dto.TeacherManualReviewSubmissionItemResp {
	answerPreview := strings.TrimSpace(record.Submission.Flag)
	if len([]rune(answerPreview)) > 80 {
		answerPreview = string([]rune(answerPreview)[:80]) + "..."
	}
	resp := practiceCommandResponseMapperInst.ToTeacherManualReviewSubmissionItemRespBase(record.Submission)
	resp.StudentUsername = record.StudentUsername
	resp.StudentName = record.StudentName
	resp.ClassName = record.ClassName
	resp.ChallengeTitle = record.ChallengeTitle
	resp.AnswerPreview = answerPreview
	return resp
}

func challengeSubmissionRecordRespFromModel(item model.Submission) *dto.ChallengeSubmissionRecordResp {
	status := dto.SubmissionStatusIncorrect
	answer := ""

	if item.ReviewStatus == model.SubmissionReviewStatusPending {
		status = dto.SubmissionStatusPendingReview
		answer = item.Flag
	} else if item.IsCorrect {
		status = dto.SubmissionStatusCorrect
	}

	resp := practiceCommandResponseMapperInst.ToChallengeSubmissionRecordRespBase(item)
	resp.Status = status
	resp.Answer = answer
	return resp
}
