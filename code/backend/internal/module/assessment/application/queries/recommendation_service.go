package queries

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	assessmentconfig "ctf-platform/internal/module/assessment/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
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
		config:        assessmentconfig.NormalizeRecommendationConfig(cfg),
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
	bus.Subscribe(contestcontracts.EventFlagAccepted, s.handleContestCacheRefreshEvent)
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

	var userID int64
	switch payload := evt.Payload.(type) {
	case contestcontracts.FlagAcceptedEvent:
		userID = payload.UserID
	case contestcontracts.AWDAttackAcceptedEvent:
		userID = payload.UserID
	default:
		return fmt.Errorf("unexpected contest cache refresh payload: %T", evt.Payload)
	}
	if userID <= 0 {
		return nil
	}
	return s.cache.DeleteRecommendations(ctx, userID)
}

func (s *RecommendationService) Recommend(ctx context.Context, userID int64, limit int) (*assessmentcontracts.Recommendation, error) {
	input, err := s.loadRecommendationInput(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(input.TargetDimensions) == 0 {
		return &assessmentcontracts.Recommendation{
			WeakDimensions: input.WeakDimensions,
			Challenges:     []*assessmentcontracts.ChallengeRecommendation{},
		}, nil
	}

	recommendations, err := s.recommendChallenges(ctx, userID, limit, input)
	if err != nil {
		return nil, err
	}

	return &assessmentcontracts.Recommendation{
		WeakDimensions: input.WeakDimensions,
		Challenges:     recommendations,
	}, nil
}

func (s *RecommendationService) RecommendChallenges(ctx context.Context, userID int64, limit int) ([]*assessmentcontracts.ChallengeRecommendation, error) {
	input, err := s.loadRecommendationInput(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(input.TargetDimensions) == 0 {
		return []*assessmentcontracts.ChallengeRecommendation{}, nil
	}
	return s.recommendChallenges(ctx, userID, limit, input)
}

func (s *RecommendationService) recommendChallenges(
	ctx context.Context,
	userID int64,
	limit int,
	input recommendationInput,
) ([]*assessmentcontracts.ChallengeRecommendation, error) {
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

	solvedIDs, err := s.getSolvedChallengeIDs(ctx, userID)
	if err != nil {
		s.logger.Error("查询已解题目失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	challenges, err := s.challengeRepo.FindPublishedForRecommendation(
		ctx,
		limit,
		input.TargetDimensions,
		string(input.Evaluation.RecommendedDifficultyBand),
		solvedIDs,
	)
	if err != nil {
		s.logger.Error("查询推荐靶场失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	plan := teachingadvice.BuildRecommendationPlan(
		input.Snapshot,
		input.Evaluation,
		toAdviceChallengeCandidates(challenges),
	)
	recommendations := buildChallengeRecommendations(challenges, plan)

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

type recommendationInput struct {
	Snapshot         teachingadvice.StudentFactSnapshot
	Evaluation       teachingadvice.StudentEvaluation
	WeakDimensions   []assessmentcontracts.RecommendationWeakDimension
	TargetDimensions []string
}

func (s *RecommendationService) loadRecommendationInput(
	ctx context.Context,
	userID int64,
) (recommendationInput, error) {
	snapshot, err := s.repo.GetStudentTeachingFactSnapshot(ctx, userID)
	if err != nil {
		s.logger.Error("查询 recommendation teaching snapshot 失败", zap.Int64("user_id", userID), zap.Error(err))
		return recommendationInput{}, err
	}
	if snapshot == nil {
		return recommendationInput{
			WeakDimensions:   []assessmentcontracts.RecommendationWeakDimension{},
			TargetDimensions: []string{},
		}, nil
	}

	evaluation := teachingadvice.EvaluateStudent(*snapshot)
	return recommendationInput{
		Snapshot:         *snapshot,
		Evaluation:       evaluation,
		WeakDimensions:   toRecommendationWeakDimensions(evaluation.WeakDimensions),
		TargetDimensions: recommendationTargetDimensions(evaluation.RecommendationTargets),
	}, nil
}

func toRecommendationWeakDimensions(
	items []teachingadvice.DimensionAdvice,
) []assessmentcontracts.RecommendationWeakDimension {
	result := make([]assessmentcontracts.RecommendationWeakDimension, 0, len(items))
	for _, item := range items {
		dimension := strings.TrimSpace(item.Dimension)
		if dimension == "" {
			continue
		}
		result = append(result, assessmentcontracts.RecommendationWeakDimension{
			Dimension:  dimension,
			Severity:   string(item.Severity),
			Confidence: item.Confidence,
			Evidence:   item.Evidence,
		})
	}
	return result
}

func recommendationTargetDimensions(items []teachingadvice.DimensionAdvice) []string {
	dimensions := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		dimension := strings.ToLower(strings.TrimSpace(item.Dimension))
		if dimension == "" {
			continue
		}
		if _, exists := seen[dimension]; exists {
			continue
		}
		seen[dimension] = struct{}{}
		dimensions = append(dimensions, dimension)
	}
	return dimensions
}

func toAdviceChallengeCandidates(
	challenges []*challengecontracts.RecommendationChallenge,
) []teachingadvice.ChallengeCandidate {
	items := make([]teachingadvice.ChallengeCandidate, 0, len(challenges))
	for _, challenge := range challenges {
		if challenge == nil {
			continue
		}
		items = append(items, teachingadvice.ChallengeCandidate{
			ID:         challenge.ID,
			Title:      challenge.Title,
			Category:   challenge.Category,
			Dimension:  challenge.RecommendationDimension,
			Difficulty: challenge.Difficulty,
			Points:     challenge.Points,
		})
	}
	return items
}

func buildChallengeRecommendations(
	challenges []*challengecontracts.RecommendationChallenge,
	plan teachingadvice.RecommendationPlan,
) []*assessmentcontracts.ChallengeRecommendation {
	recommendations := make([]*assessmentcontracts.ChallengeRecommendation, 0, len(challenges))
	for index, challenge := range challenges {
		if challenge == nil || index >= len(plan.Reasons) {
			continue
		}
		reason := plan.Reasons[index]
		dimension := strings.TrimSpace(reason.Dimension)
		if dimension == "" {
			dimension = strings.TrimSpace(challenge.RecommendationDimension)
		}
		if dimension == "" {
			dimension = challenge.Category
		}
		recommendations = append(recommendations, &assessmentcontracts.ChallengeRecommendation{
			ID:             challenge.ID,
			Title:          challenge.Title,
			Category:       challenge.Category,
			Difficulty:     challenge.Difficulty,
			Points:         challenge.Points,
			Dimension:      dimension,
			DifficultyBand: string(reason.DifficultyBand),
			Severity:       string(reason.Severity),
			ReasonCodes:    append([]string(nil), reason.ReasonCodes...),
			Summary:        reason.Summary,
			Evidence:       reason.Evidence,
		})
	}
	return recommendations
}
