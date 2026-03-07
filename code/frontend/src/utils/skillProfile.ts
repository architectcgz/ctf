import type { ChallengeCategory, RecommendationItem, SkillProfileData } from '@/api/contracts'

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
  id: string | number
  title: string
  category: ChallengeCategory
  difficulty: RecommendationItem['difficulty']
  reason: string
}

export interface RawRecommendationResponse {
  weak_dimensions?: string[]
  challenges: RawRecommendationChallenge[]
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
  web: '#3b82f6',
  pwn: '#ef4444',
  reverse: '#8b5cf6',
  crypto: '#f59e0b',
  misc: '#10b981',
  forensics: '#06b6d4',
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

export function normalizeRecommendations(raw: RawRecommendationResponse): RecommendationItem[] {
  return raw.challenges.map((challenge) => ({
    challenge_id: String(challenge.id),
    title: challenge.title,
    category: challenge.category,
    difficulty: challenge.difficulty,
    reason: challenge.reason,
  }))
}

export function toRadarScores(profile: SkillProfileData | null) {
  if (!profile) return []
  return profile.dimensions.map((item) => ({
    name: item.name,
    value: item.value,
    color: dimensionColors[item.key as ChallengeCategory] || '#0891b2',
  }))
}

export function getWeakDimensions(profile: SkillProfileData | null, threshold = 60): string[] {
  if (!profile) return []
  return profile.dimensions.filter((item) => item.value < threshold).map((item) => item.name)
}
