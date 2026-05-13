package queries

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type WriteupService struct {
	repo writeupQueryRepository
}

type writeupQueryRepository interface {
	challengeports.ChallengeWriteupChallengeLookupRepository
	challengeports.ChallengeWriteupUserLookupRepository
	challengeports.ChallengeAdminWriteupRepository
	challengeports.ChallengeReleasedWriteupRepository
	challengeports.ChallengeWriteupSolveStatusRepository
	challengeports.ChallengeSubmissionWriteupRepository
	challengeports.ChallengeTeacherSubmissionWriteupRepository
	challengeports.ChallengeSolutionQueryRepository
}

func NewWriteupService(repo writeupQueryRepository) *WriteupService {
	return &WriteupService{repo: repo}
}

func (s *WriteupService) GetAdmin(ctx context.Context, challengeID int64) (*dto.AdminChallengeWriteupResp, error) {
	if _, err := s.repo.FindByID(ctx, challengeID); err != nil {
		if errors.Is(err, challengeports.ErrChallengeWriteupChallengeNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	item, err := s.repo.FindWriteupByChallengeID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeOfficialWriteupNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	return domain.ResponseMapper().ToAdminChallengeWriteupRespPtr(item), nil
}

func (s *WriteupService) GetPublished(ctx context.Context, userID, challengeID int64) (*dto.ChallengeWriteupResp, error) {
	challengeItem, err := s.repo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeWriteupChallengeNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	item, err := s.repo.FindReleasedWriteupByChallengeID(ctx, challengeID, time.Now())
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeReleasedWriteupNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}

	isSolved, err := s.repo.GetSolvedStatus(ctx, userID, challengeID)
	if err != nil {
		isSolved = false
	}

	resp := challengeQueryResponseMapperInst.ToChallengeWriteupRespBasePtr(item)
	resp.RequiresSpoilerWarning = !isSolved
	return resp, nil
}

func (s *WriteupService) GetMySubmission(ctx context.Context, userID, challengeID int64) (*dto.SubmissionWriteupResp, error) {
	challengeItem, err := s.repo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeWriteupChallengeNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}
	item, err := s.repo.FindSubmissionWriteupByUserChallenge(ctx, userID, challengeID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeSubmissionWriteupNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return domain.ResponseMapper().ToSubmissionWriteupRespPtr(item), nil
}

func (s *WriteupService) ListRecommendedSolutions(ctx context.Context, userID, challengeID int64) (*dto.PageResult[*dto.RecommendedChallengeSolutionResp], error) {
	if err := s.ensureSolvedChallengeVisible(ctx, userID, challengeID); err != nil {
		return nil, err
	}

	items, err := s.repo.ListRecommendedSolutionsByChallengeID(ctx, challengeID, time.Now())
	if err != nil {
		return nil, err
	}
	respItems := make([]*dto.RecommendedChallengeSolutionResp, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, domain.RecommendedSolutionRespFromRecord(item))
	}
	return &dto.PageResult[*dto.RecommendedChallengeSolutionResp]{
		List:  respItems,
		Total: int64(len(respItems)),
		Page:  1,
		Size:  len(respItems),
	}, nil
}

func (s *WriteupService) ListCommunitySolutions(ctx context.Context, userID, challengeID int64, query *dto.CommunityChallengeSolutionQuery) (*dto.PageResult[*dto.CommunityChallengeSolutionResp], error) {
	if err := s.ensureSolvedChallengeVisible(ctx, userID, challengeID); err != nil {
		return nil, err
	}

	normalized := &dto.CommunityChallengeSolutionQuery{Page: 1, Size: 20}
	if query != nil {
		normalized = &dto.CommunityChallengeSolutionQuery{
			Q:    query.Q,
			Sort: query.Sort,
			Page: query.Page,
			Size: query.Size,
		}
		if normalized.Page <= 0 {
			normalized.Page = 1
		}
		if normalized.Size <= 0 {
			normalized.Size = 20
		}
	}

	items, total, err := s.repo.ListCommunitySolutionsByChallengeID(ctx, challengeID, normalized)
	if err != nil {
		return nil, err
	}
	respItems := make([]*dto.CommunityChallengeSolutionResp, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, domain.CommunitySolutionRespFromRecord(item))
	}
	return &dto.PageResult[*dto.CommunityChallengeSolutionResp]{
		List:  respItems,
		Total: total,
		Page:  normalized.Page,
		Size:  normalized.Size,
	}, nil
}

func (s *WriteupService) ListTeacherSubmissions(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherSubmissionWriteupQuery) (*dto.PageResult[*dto.TeacherSubmissionWriteupItemResp], error) {
	if query == nil {
		query = &dto.TeacherSubmissionWriteupQuery{}
	}
	normalized, err := normalizeTeacherSubmissionQuery(ctx, s.repo, requesterID, requesterRole, query)
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListTeacherSubmissionWriteups(ctx, normalized)
	if err != nil {
		return nil, err
	}

	respItems := make([]*dto.TeacherSubmissionWriteupItemResp, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, domain.TeacherSubmissionWriteupItemRespFromRecord(item))
	}

	return &dto.PageResult[*dto.TeacherSubmissionWriteupItemResp]{
		List:  respItems,
		Total: total,
		Page:  normalized.Page,
		Size:  normalized.Size,
	}, nil
}

func (s *WriteupService) GetTeacherSubmission(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.TeacherSubmissionWriteupDetailResp, error) {
	record, err := s.repo.GetTeacherSubmissionWriteupByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeTeacherSubmissionWriteupNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if err := ensureTeacherCanAccessQueryRecord(ctx, s.repo, requesterID, requesterRole, record); err != nil {
		return nil, err
	}
	return domain.TeacherSubmissionWriteupDetailRespFromRecord(*record), nil
}

func normalizeTeacherSubmissionQuery(
	ctx context.Context,
	repo challengeports.ChallengeWriteupUserLookupRepository,
	requesterID int64,
	requesterRole string,
	query *dto.TeacherSubmissionWriteupQuery,
) (*dto.TeacherSubmissionWriteupQuery, error) {
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
		if errors.Is(err, challengeports.ErrChallengeWriteupRequesterNotFound) {
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

func ensureTeacherCanAccessQueryRecord(
	ctx context.Context,
	repo challengeports.ChallengeWriteupUserLookupRepository,
	requesterID int64,
	requesterRole string,
	record *challengeports.TeacherSubmissionWriteupRecord,
) error {
	if requesterRole == model.RoleAdmin {
		return nil
	}
	requester, err := repo.FindUserByID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeWriteupRequesterNotFound) {
			return errcode.ErrUnauthorized
		}
		return err
	}
	if requester.ClassName == "" || requester.ClassName != record.ClassName {
		return errcode.ErrForbidden
	}
	return nil
}

func (s *WriteupService) ensureSolvedChallengeVisible(ctx context.Context, userID, challengeID int64) error {
	challengeItem, err := s.repo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeWriteupChallengeNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}
	if challengeItem.Status != model.ChallengeStatusPublished {
		return errcode.ErrChallengeNotPublish
	}
	isSolved, err := s.repo.GetSolvedStatus(ctx, userID, challengeID)
	if err != nil {
		return err
	}
	if !isSolved {
		return errcode.ErrForbidden
	}
	return nil
}
