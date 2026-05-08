import type {
  AdviceSeverity,
  ChallengeCategory,
  RecommendationData,
  RecommendationItem,
  RecommendationWeakDimension,
  SkillProfileData,
} from '@/api/contracts'

export interface RawSkillProfileDimension {
  dimension: ChallengeCategory
  score: number
}

export interface RawSkillProfileResponse {
  user_id: string | number
  dimensions: RawSkillProfileDimension[]
  updated_at?: string
}

export interface RawRecommendationChallenge {
  id?: string | number
  challenge_id?: string | number
  title: string
  category: ChallengeCategory
  difficulty: RecommendationItem['difficulty']
  dimension?: string
  difficulty_band?: RecommendationItem['difficulty_band']
  severity?: AdviceSeverity
  reason_codes?: string[]
  summary: string
  evidence?: string
}

export interface RawRecommendationResponse {
  weak_dimensions?: RawRecommendationWeakDimension[]
  challenges: RawRecommendationChallenge[]
}

export interface RawRecommendationWeakDimension {
  dimension: string
  severity: AdviceSeverity
  confidence: number
  evidence?: string
}

const dimensionLabels: Record<ChallengeCategory, string> = {
  web: 'Web',
  pwn: 'Pwn',
  reverse: '逆向',
  crypto: '密码',
  misc: '杂项',
  forensics: '取证',
}

const dimensionColors: Record<ChallengeCategory, string> = {
  web: 'var(--color-cat-web)',
  pwn: 'var(--color-cat-pwn)',
  reverse: 'var(--color-cat-reverse)',
  crypto: 'var(--color-cat-crypto)',
  misc: 'var(--color-cat-misc)',
  forensics: 'var(--color-cat-forensics)',
}

export function normalizeSkillProfile(raw: RawSkillProfileResponse): SkillProfileData {
  return {
    updated_at: raw.updated_at,
    dimensions: raw.dimensions.map((item) => ({
      key: item.dimension,
      name: dimensionLabels[item.dimension],
      value: Math.max(0, Math.min(100, Math.round(item.score * 100))),
    })),
  }
}

function toDimensionLabel(dimension: string): string {
  return dimensionLabels[dimension as ChallengeCategory] || dimension
}

function normalizeWeakDimensions(
  weakDimensions: RawRecommendationWeakDimension[] | undefined
): RecommendationWeakDimension[] {
  return (weakDimensions ?? []).map((item) => ({
    ...item,
    label: toDimensionLabel(item.dimension),
  }))
}

function normalizeRecommendationItems(
  challenges: RawRecommendationChallenge[]
): RecommendationItem[] {
  return challenges.map((challenge) => ({
    challenge_id: String(challenge.challenge_id ?? challenge.id),
    title: challenge.title,
    category: challenge.category,
    difficulty: challenge.difficulty,
    dimension: challenge.dimension,
    difficulty_band: challenge.difficulty_band,
    severity: challenge.severity,
    reason_codes: challenge.reason_codes,
    summary: challenge.summary,
    evidence: challenge.evidence,
  }))
}

export function normalizeRecommendationData(raw: RawRecommendationResponse): RecommendationData {
  return {
    weak_dimensions: normalizeWeakDimensions(raw.weak_dimensions),
    challenges: normalizeRecommendationItems(raw.challenges),
  }
}

export function toRadarScores(profile: SkillProfileData | null) {
  if (!profile) return []
  return profile.dimensions.map((item) => ({
    name: item.name,
    value: item.value,
    color: dimensionColors[item.key as ChallengeCategory] || 'var(--color-primary)',
  }))
}

export function getWeakDimensionLabels(weakDimensions: RecommendationWeakDimension[]): string[] {
  return weakDimensions.map((item) => item.label)
}
