package queries

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	assessmentconfig "ctf-platform/internal/module/assessment/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/shared/taxonomy"
)

type RecommendationService struct {
	repo          recommendationRepository
	challengeRepo assessmentports.RecommendationChallengeRepository
	cache         assessmentports.AssessmentRecommendationCacheStore
	logger        *zap.Logger
	config        config.RecommendationConfig
}

type recommendationRepository interface {
	assessmentports.RecommendationProfileRepository
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

func (s *RecommendationService) Recommend(ctx context.Context, userID int64, limit int) (*assessmentcontracts.Recommendation, error) {
	weakDimensions, weakDimensionScores, targetDimensions, difficultyBand, err := s.loadWeakDimensions(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(targetDimensions) == 0 {
		return &assessmentcontracts.Recommendation{
			WeakDimensions: weakDimensions,
			Challenges:     []*assessmentcontracts.ChallengeRecommendation{},
		}, nil
	}

	recommendations, err := s.recommendChallenges(ctx, userID, limit, targetDimensions, difficultyBand, weakDimensionScores)
	if err != nil {
		return nil, err
	}

	return &assessmentcontracts.Recommendation{
		WeakDimensions: weakDimensions,
		Challenges:     recommendations,
	}, nil
}

func (s *RecommendationService) RecommendChallenges(ctx context.Context, userID int64, limit int) ([]*assessmentcontracts.ChallengeRecommendation, error) {
	_, weakDimensionScores, targetDimensions, difficultyBand, err := s.loadWeakDimensions(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(targetDimensions) == 0 {
		return []*assessmentcontracts.ChallengeRecommendation{}, nil
	}
	return s.recommendChallenges(ctx, userID, limit, targetDimensions, difficultyBand, weakDimensionScores)
}

func (s *RecommendationService) recommendChallenges(
	ctx context.Context,
	userID int64,
	limit int,
	targetDimensions []string,
	difficultyBand string,
	weakDimensionScores map[string]float64,
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
		targetDimensions,
		difficultyBand,
		solvedIDs,
	)
	if err != nil {
		s.logger.Error("查询推荐靶场失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	recommendations := make([]*assessmentcontracts.ChallengeRecommendation, 0, len(challenges))
	for _, challenge := range challenges {
		if challenge == nil {
			continue
		}
		dimension := strings.TrimSpace(challenge.RecommendationDimension)
		if dimension == "" {
			dimension = challenge.Category
		}
		score := weakDimensionScores[dimension]
		severity := weakDimensionSeverity(score, s.config.WeakThreshold)
		recommendations = append(recommendations, &assessmentcontracts.ChallengeRecommendation{
			ID:             challenge.ID,
			Title:          challenge.Title,
			Category:       challenge.Category,
			Difficulty:     challenge.Difficulty,
			Points:         challenge.Points,
			Dimension:      dimension,
			DifficultyBand: difficultyBand,
			Severity:       string(severity),
			ReasonCodes:    []string{"low_dimension_score", "coverage_gap"},
			Summary:        buildRecommendationSummary(dimension, score, difficultyBand),
			Evidence:       buildRecommendationEvidence(dimension, score),
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

func (s *RecommendationService) loadWeakDimensions(
	ctx context.Context,
	userID int64,
) ([]assessmentcontracts.RecommendationWeakDimension, map[string]float64, []string, string, error) {
	profiles, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询能力画像失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, nil, nil, "", err
	}
	weakDimensions := buildWeakDimensionsFromProfiles(profiles, s.config.WeakThreshold)
	if len(weakDimensions) == 0 {
		return []assessmentcontracts.RecommendationWeakDimension{}, map[string]float64{}, []string{}, "", nil
	}
	scores := make(map[string]float64, len(weakDimensions))
	targets := make([]string, 0, 1)
	for index, item := range weakDimensions {
		score := clampScore(1 - item.Confidence)
		scores[item.Dimension] = score
		if index == 0 {
			targets = append(targets, item.Dimension)
		}
	}
	lowestScore := scores[weakDimensions[0].Dimension]
	return weakDimensions, scores, targets, preferredDifficultyForWeakScore(lowestScore, s.config.WeakThreshold), nil
}

func buildWeakDimensionsFromProfiles(
	profiles []*assessmententity.SkillProfile,
	weakThreshold float64,
) []assessmentcontracts.RecommendationWeakDimension {
	type scoreItem struct {
		dimension string
		score     float64
	}

	items := make([]scoreItem, 0, len(profiles))
	for _, item := range profiles {
		if item == nil || !taxonomy.IsValidDimension(item.Dimension) {
			continue
		}
		score := clampScore(item.Score)
		if score >= weakThreshold {
			continue
		}
		items = append(items, scoreItem{dimension: item.Dimension, score: score})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score < items[j].score
		}
		return items[i].dimension < items[j].dimension
	})

	result := make([]assessmentcontracts.RecommendationWeakDimension, 0, len(items))
	for _, item := range items {
		result = append(result, assessmentcontracts.RecommendationWeakDimension{
			Dimension:  item.dimension,
			Severity:   string(weakDimensionSeverity(item.score, weakThreshold)),
			Confidence: clampScore(1 - item.score),
			Evidence:   buildWeakDimensionEvidence(item.dimension, item.score),
		})
	}
	return result
}

func weakDimensionSeverity(score, weakThreshold float64) string {
	score = clampScore(score)
	dangerThreshold := clampScore(weakThreshold / 2)
	if score <= dangerThreshold {
		return "danger"
	}
	return "warning"
}

func preferredDifficultyForWeakScore(score, weakThreshold float64) string {
	score = clampScore(score)
	if score <= clampScore(weakThreshold/2) {
		return taxonomy.DifficultyBeginner
	}
	return taxonomy.DifficultyEasy
}

func buildWeakDimensionEvidence(dimension string, score float64) string {
	return fmt.Sprintf("%s 维度当前完成分值占已发布题目分值的 %.0f%%。", dimensionLabel(dimension), score*100)
}

func buildRecommendationSummary(dimension string, score float64, difficultyBand string) string {
	return fmt.Sprintf("当前%s维度训练覆盖为 %.0f%%，建议先补%s难度题。", dimensionLabel(dimension), score*100, difficultyLabel(difficultyBand))
}

func buildRecommendationEvidence(dimension string, score float64) string {
	return fmt.Sprintf("%s。优先补齐该维度的基础训练样本。", buildWeakDimensionEvidence(dimension, score))
}

func dimensionLabel(dimension string) string {
	switch strings.TrimSpace(dimension) {
	case taxonomy.DimensionWeb:
		return "Web"
	case taxonomy.DimensionPwn:
		return "Pwn"
	case taxonomy.DimensionReverse:
		return "逆向"
	case taxonomy.DimensionCrypto:
		return "密码"
	case taxonomy.DimensionMisc:
		return "杂项"
	case taxonomy.DimensionForensics:
		return "取证"
	default:
		return dimension
	}
}

func difficultyLabel(difficulty string) string {
	switch strings.TrimSpace(difficulty) {
	case taxonomy.DifficultyBeginner:
		return "入门"
	case taxonomy.DifficultyEasy:
		return "简单"
	case taxonomy.DifficultyMedium:
		return "中等"
	case taxonomy.DifficultyHard:
		return "困难"
	case taxonomy.DifficultyInsane:
		return "高难"
	default:
		return difficulty
	}
}

func clampScore(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
