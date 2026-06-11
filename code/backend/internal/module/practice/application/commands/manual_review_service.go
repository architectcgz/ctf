package commands

import (
	"context"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"errors"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
)

func (s *serviceCore) ReviewManualReviewSubmission(
	ctx context.Context,
	submissionID, reviewerID int64,
	reviewerRole string,
	req *practicecontracts.ReviewManualReviewSubmissionReq,
) (*practicecontracts.TeacherManualReviewSubmissionDetailResp, error) {
	if err := ensureManualReviewRequesterRole(reviewerRole); err != nil {
		return nil, err
	}
	if err := ensureManualReviewDecisionStatus(req); err != nil {
		return nil, err
	}
	if s.manualReviewRepo == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("practice manual review repository is nil"))
	}
	if s.runtimeSubject == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("practice runtime subject repository is nil"))
	}

	record, err := s.manualReviewRepo.GetTeacherManualReviewSubmissionByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeManualReviewSubmissionNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessManualReviewSubmission(ctx, s.manualReviewRepo, reviewerID, reviewerRole, record); err != nil {
		return nil, err
	}
	if record.Submission.ReviewStatus != practiceports.SubmissionReviewStatusPending {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("仅待审核提交可执行评阅"))
	}

	challengeItem, err := s.runtimeSubject.FindByID(ctx, record.Submission.ChallengeID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeChallengeNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if challengeItem.FlagType != practiceentity.FlagTypeManualReview {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("当前提交不属于人工审核题"))
	}

	now := time.Now().UTC()
	item := record.Submission
	item.ReviewStatus = req.ReviewStatus
	item.ReviewComment = strings.TrimSpace(req.ReviewComment)
	item.ReviewedBy = &reviewerID
	item.ReviewedAt = &now
	item.UpdatedAt = now
	if req.ReviewStatus == practiceports.SubmissionReviewStatusApproved {
		if _, err := s.manualReviewRepo.FindCorrectSubmission(ctx, item.UserID, item.ChallengeID); err == nil {
			return nil, challengecontracts.ErrAlreadySolved
		} else if err != nil && !errors.Is(err, practiceports.ErrPracticeSolvedSubmissionNotFound) {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		item.IsCorrect = true
		item.Score = challengeItem.Points
	} else {
		item.IsCorrect = false
		item.Score = 0
	}

	if err := s.manualReviewRepo.UpdateSubmission(ctx, &item); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if item.IsCorrect {
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

func (s *serviceCore) ListTeacherManualReviewSubmissions(
	ctx context.Context,
	requesterID int64,
	requesterRole string,
	query *practicecontracts.TeacherManualReviewSubmissionQuery,
) (*practicecontracts.PageResult[*practicecontracts.TeacherManualReviewSubmissionItemResp], error) {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	if query == nil {
		query = &practicecontracts.TeacherManualReviewSubmissionQuery{}
	}
	if err := ensureManualReviewQuery(query); err != nil {
		return nil, err
	}
	if s.manualReviewRepo == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("practice manual review repository is nil"))
	}
	normalized, err := normalizeTeacherManualReviewQuery(ctx, s.manualReviewRepo, requesterID, requesterRole, query)
	if err != nil {
		return nil, err
	}

	items, total, err := s.manualReviewRepo.ListTeacherManualReviewSubmissions(ctx, normalized)
	if err != nil {
		return nil, err
	}

	respItems := make([]*practicecontracts.TeacherManualReviewSubmissionItemResp, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, manualReviewListItemRespFromRecord(item))
	}

	return &practicecontracts.PageResult[*practicecontracts.TeacherManualReviewSubmissionItemResp]{
		List:  respItems,
		Total: total,
		Page:  normalized.Page,
		Size:  normalized.Size,
	}, nil
}

func (s *serviceCore) GetTeacherManualReviewSubmission(
	ctx context.Context,
	submissionID, requesterID int64,
	requesterRole string,
) (*practicecontracts.TeacherManualReviewSubmissionDetailResp, error) {
	if err := ensureManualReviewRequesterRole(requesterRole); err != nil {
		return nil, err
	}
	if s.manualReviewRepo == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("practice manual review repository is nil"))
	}
	record, err := s.manualReviewRepo.GetTeacherManualReviewSubmissionByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeManualReviewSubmissionNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessManualReviewSubmission(ctx, s.manualReviewRepo, requesterID, requesterRole, record); err != nil {
		return nil, err
	}
	return manualReviewDetailRespFromRecord(*record, record.Submission), nil
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
	if requesterRole == identitycontracts.RoleAdmin {
		return nil
	}
	requester, err := repo.FindUserByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeUserNotFound) {
			return apperror.ErrUnauthorized
		}
		return err
	}
	if requester.ClassName == "" || requester.ClassName != record.ClassName {
		return apperror.ErrForbidden
	}
	return nil
}

func normalizeTeacherManualReviewQuery(
	ctx context.Context,
	repo practiceports.PracticeUserLookupRepository,
	requesterID int64,
	requesterRole string,
	query *practicecontracts.TeacherManualReviewSubmissionQuery,
) (*practicecontracts.TeacherManualReviewSubmissionQuery, error) {
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
	if requesterRole == identitycontracts.RoleAdmin {
		return &normalized, nil
	}

	requester, err := repo.FindUserByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeUserNotFound) {
			return nil, apperror.ErrUnauthorized
		}
		return nil, err
	}
	if requester.ClassName == "" {
		return nil, apperror.ErrForbidden
	}
	if normalized.ClassName != "" && normalized.ClassName != requester.ClassName {
		return nil, apperror.ErrForbidden
	}
	normalized.ClassName = requester.ClassName
	return &normalized, nil
}

func ensureManualReviewRequesterRole(role string) error {
	if role == identitycontracts.RoleAdmin || role == identitycontracts.RoleTeacher {
		return nil
	}
	return apperror.ErrForbidden
}

func ensureManualReviewDecisionStatus(req *practicecontracts.ReviewManualReviewSubmissionReq) error {
	if req == nil {
		return apperror.ErrInvalidParams.WithCause(errors.New("评阅请求不能为空"))
	}
	if len([]rune(strings.TrimSpace(req.ReviewComment))) > 4000 {
		return apperror.ErrInvalidParams.WithCause(errors.New("review_comment 不能超过 4000 个字符"))
	}
	if req.ReviewStatus == practiceports.SubmissionReviewStatusApproved || req.ReviewStatus == practiceports.SubmissionReviewStatusRejected {
		return nil
	}
	return apperror.ErrInvalidParams.WithCause(errors.New("review_status 仅支持 approved 或 rejected"))
}

func ensureManualReviewQuery(query *practicecontracts.TeacherManualReviewSubmissionQuery) error {
	if query == nil {
		return nil
	}
	if query.StudentID != nil && *query.StudentID <= 0 {
		return apperror.ErrInvalidParams.WithCause(errors.New("student_id 必须大于 0"))
	}
	if query.ChallengeID != nil && *query.ChallengeID <= 0 {
		return apperror.ErrInvalidParams.WithCause(errors.New("challenge_id 必须大于 0"))
	}
	if len([]rune(strings.TrimSpace(query.ClassName))) > 128 {
		return apperror.ErrInvalidParams.WithCause(errors.New("class_name 不能超过 128 个字符"))
	}
	if query.Size > 100 {
		return apperror.ErrInvalidParams.WithCause(errors.New("page_size 不能超过 100"))
	}
	if query.ReviewStatus == "" ||
		query.ReviewStatus == practiceports.SubmissionReviewStatusPending ||
		query.ReviewStatus == practiceports.SubmissionReviewStatusApproved ||
		query.ReviewStatus == practiceports.SubmissionReviewStatusRejected {
		return nil
	}
	return apperror.ErrInvalidParams.WithCause(errors.New("review_status 仅支持 pending、approved 或 rejected"))
}

func manualReviewDetailRespFromRecord(
	record practiceports.TeacherManualReviewSubmissionRecord,
	submission practiceports.SubmissionRecord,
) *practicecontracts.TeacherManualReviewSubmissionDetailResp {
	resp := &practicecontracts.TeacherManualReviewSubmissionDetailResp{
		ID:              submission.ID,
		UserID:          submission.UserID,
		StudentUsername: record.StudentUsername,
		StudentName:     record.StudentName,
		ClassName:       record.ClassName,
		ChallengeID:     submission.ChallengeID,
		ChallengeTitle:  record.ChallengeTitle,
		Answer:          submission.Flag,
		IsCorrect:       submission.IsCorrect,
		Score:           submission.Score,
		ReviewStatus:    submission.ReviewStatus,
		ReviewedAt:      CopyTimePtr(submission.ReviewedAt),
		ReviewComment:   submission.ReviewComment,
		SubmittedAt:     CopyTime(submission.SubmittedAt),
		UpdatedAt:       CopyTime(submission.UpdatedAt),
		ReviewerName:    record.ReviewerName,
	}
	if submission.ReviewedBy != nil {
		reviewedBy := *submission.ReviewedBy
		resp.ReviewedBy = &reviewedBy
	}
	return resp
}

func manualReviewListItemRespFromRecord(record practiceports.TeacherManualReviewSubmissionRecord) *practicecontracts.TeacherManualReviewSubmissionItemResp {
	answerPreview := strings.TrimSpace(record.Submission.Flag)
	if len([]rune(answerPreview)) > 80 {
		answerPreview = string([]rune(answerPreview)[:80]) + "..."
	}
	return &practicecontracts.TeacherManualReviewSubmissionItemResp{
		ID:              record.Submission.ID,
		UserID:          record.Submission.UserID,
		StudentUsername: record.StudentUsername,
		StudentName:     record.StudentName,
		ClassName:       record.ClassName,
		ChallengeID:     record.Submission.ChallengeID,
		ChallengeTitle:  record.ChallengeTitle,
		AnswerPreview:   answerPreview,
		ReviewStatus:    record.Submission.ReviewStatus,
		SubmittedAt:     CopyTime(record.Submission.SubmittedAt),
		ReviewedAt:      CopyTimePtr(record.Submission.ReviewedAt),
		UpdatedAt:       CopyTime(record.Submission.UpdatedAt),
	}
}
