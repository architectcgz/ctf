package queries

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

type RecommendationService struct {
	repo          recommendationRepository
	challengeRepo assessmentports.RecommendationChallengeRepository
	cache         assessmentports.AssessmentRecommendationCacheStore
	logger        *zap.Logger
	config        config.RecommendationConfig
}

type recommendationRepository interface {
	assessmentports.RecommendationTeachingFactRepository
	assessmentports.RecommendationSolvedChallengeRepository
}

var _ assessmentcontracts.RecommendationProvider = (*RecommendationService)(nil)

func NewRecommendationService(repo recommendationRepository, challengeRepo assessmentports.RecommendationChallengeRepository, cache assessmentports.AssessmentRecommendationCacheStore, cfg config.RecommendationConfig, logger *zap.Logger) *RecommendationService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RecommendationService{
		repo:          repo,
		challengeRepo: challengeRepo,
		cache:         cache,
		logger:        logger,
		config:        assessmentdomain.NormalizeRecommendationConfig(cfg),
	}
}

func (s *RecommendationService) RegisterPracticeEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(practicecontracts.EventFlagAccepted, s.handlePracticeCacheRefreshEvent)
}

func (s *RecommendationService) RegisterContestEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(contestcontracts.EventAWDAttackAccepted, s.handleContestCacheRefreshEvent)
}

func (s *RecommendationService) handlePracticeCacheRefreshEvent(ctx context.Context, evt platformevents.Event) error {
	if s.cache == nil {
		return nil
	}

	var userID int64
	switch payload := evt.Payload.(type) {
	case practicecontracts.FlagAcceptedEvent:
		userID = payload.UserID
	default:
		return fmt.Errorf("unexpected practice cache refresh payload: %T", evt.Payload)
	}
	if userID <= 0 {
		return nil
	}
	return s.cache.DeleteRecommendations(ctx, userID)
}

func (s *RecommendationService) handleContestCacheRefreshEvent(ctx context.Context, evt platformevents.Event) error {
	if s.cache == nil {
		return nil
	}

	payload, ok := evt.Payload.(contestcontracts.AWDAttackAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest awd cache refresh payload: %T", evt.Payload)
	}
	if payload.UserID <= 0 {
		return nil
	}
	return s.cache.DeleteRecommendations(ctx, payload.UserID)
}

func (s *RecommendationService) Recommend(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error) {
	snapshot, evaluation, err := s.evaluateUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	recommendations, err := s.recommendChallengesWithEvaluation(ctx, userID, limit, snapshot, evaluation)
	if err != nil {
		return nil, err
	}

	return &dto.RecommendationResp{
		WeakDimensions: toWeakDimensionDTOs(evaluation.WeakDimensions),
		Challenges:     recommendations,
	}, nil
}

func (s *RecommendationService) RecommendChallenges(ctx context.Context, userID int64, limit int) ([]*dto.ChallengeRecommendation, error) {
	snapshot, evaluation, err := s.evaluateUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.recommendChallengesWithEvaluation(ctx, userID, limit, snapshot, evaluation)
}

func (s *RecommendationService) recommendChallengesWithEvaluation(
	ctx context.Context,
	userID int64,
	limit int,
	snapshot *teachingadvice.StudentFactSnapshot,
	evaluation teachingadvice.StudentEvaluation,
) ([]*dto.ChallengeRecommendation, error) {
	if limit <= 0 {
		limit = s.config.DefaultLimit
	}
	if limit > s.config.MaxLimit {
		limit = s.config.MaxLimit
	}

	useCache := limit == s.config.DefaultLimit
	if useCache && s.cache != nil {
		recommendations, found, err := s.cache.LoadRecommendations(ctx, userID)
		if err == nil && found {
			return recommendations, nil
		}
		if err != nil {
			s.logger.Warn("推荐缓存读取失败", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	targetDimensions := recommendationTargetDimensions(evaluation)
	if len(targetDimensions) == 0 {
		return []*dto.ChallengeRecommendation{}, nil
	}

	solvedIDs, err := s.getSolvedChallengeIDs(ctx, userID)
	if err != nil {
		s.logger.Error("查询已解题目失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	challenges, err := s.challengeRepo.FindPublishedForRecommendation(ctx, limit, targetDimensions, solvedIDs)
	if err != nil {
		s.logger.Error("查询推荐靶场失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	if snapshot == nil {
		return []*dto.ChallengeRecommendation{}, nil
	}
	candidates := make([]teachingadvice.ChallengeCandidate, 0, len(challenges))
	for _, challenge := range challenges {
		if challenge == nil {
			continue
		}
		candidates = append(candidates, teachingadvice.ChallengeCandidate{
			ID:         challenge.ID,
			Title:      challenge.Title,
			Category:   challenge.Category,
			Difficulty: challenge.Difficulty,
			Points:     challenge.Points,
		})
	}

	plan := teachingadvice.BuildRecommendationPlan(*snapshot, evaluation, candidates)
	reasonsByChallengeID := make(map[int64]teachingadvice.RecommendationReason, len(plan.Reasons))
	for index, reason := range plan.Reasons {
		if index >= len(candidates) {
			break
		}
		reasonsByChallengeID[candidates[index].ID] = reason
	}

	recommendations := make([]*dto.ChallengeRecommendation, 0, len(challenges))
	for _, challenge := range challenges {
		if challenge == nil {
			continue
		}
		reason := reasonsByChallengeID[challenge.ID]
		recommendations = append(recommendations, &dto.ChallengeRecommendation{
			ID:             challenge.ID,
			Title:          challenge.Title,
			Category:       challenge.Category,
			Difficulty:     challenge.Difficulty,
			Points:         challenge.Points,
			Dimension:      reason.Dimension,
			DifficultyBand: string(reason.DifficultyBand),
			Severity:       string(reason.Severity),
			ReasonCodes:    append([]string(nil), reason.ReasonCodes...),
			Summary:        reason.Summary,
			Evidence:       reason.Evidence,
		})
	}

	if useCache && s.cache != nil {
		if err := s.cache.StoreRecommendations(ctx, userID, recommendations, s.config.CacheTTL); err != nil {
			s.logger.Warn("推荐缓存写入失败", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	return recommendations, nil
}

func (s *RecommendationService) getSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error) {
	return s.repo.ListSolvedChallengeIDs(ctx, userID)
}

type RecommendationQuery struct {
	Limit int `form:"limit"`
}

func (s *RecommendationService) evaluateUser(
	ctx context.Context,
	userID int64,
) (*teachingadvice.StudentFactSnapshot, teachingadvice.StudentEvaluation, error) {
	snapshot, err := s.repo.GetStudentTeachingFactSnapshot(ctx, userID)
	if err != nil {
		s.logger.Error("查询教学事实快照失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, teachingadvice.StudentEvaluation{}, err
	}
	if snapshot == nil {
		return nil, teachingadvice.StudentEvaluation{}, nil
	}
	return snapshot, teachingadvice.EvaluateStudent(*snapshot), nil
}

func recommendationTargetDimensions(evaluation teachingadvice.StudentEvaluation) []string {
	targets := make([]string, 0, len(evaluation.RecommendationTargets))
	seen := make(map[string]struct{}, len(evaluation.RecommendationTargets))
	for _, item := range evaluation.RecommendationTargets {
		dimension := item.Dimension
		if dimension == "" {
			continue
		}
		if _, ok := seen[dimension]; ok {
			continue
		}
		seen[dimension] = struct{}{}
		targets = append(targets, dimension)
	}
	return targets
}

func toWeakDimensionDTOs(items []teachingadvice.DimensionAdvice) []dto.RecommendationWeakDimension {
	result := make([]dto.RecommendationWeakDimension, 0, len(items))
	for _, item := range items {
		result = append(result, dto.RecommendationWeakDimension{
			Dimension:  item.Dimension,
			Severity:   string(item.Severity),
			Confidence: item.Confidence,
			Evidence:   item.Evidence,
		})
	}
	return result
}
