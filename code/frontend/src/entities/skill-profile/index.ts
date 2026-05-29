export type {
  AdviceSeverity,
  RawRecommendationChallenge,
  RawRecommendationResponse,
  RawRecommendationWeakDimension,
  RawSkillProfileDimension,
  RawSkillProfileResponse,
  RecommendationData,
  RecommendationDifficultyBand,
  RecommendationItem,
  RecommendationWeakDimension,
  SkillDimensionScore,
  SkillProfileData,
} from './model'
export {
  getWeakDimensionLabels,
  normalizeRecommendationData,
  normalizeSkillProfile,
  toRadarScores,
} from './model'
