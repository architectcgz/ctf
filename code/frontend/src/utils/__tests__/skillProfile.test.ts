import { describe, expect, it } from 'vitest'

import {
  getWeakDimensionLabels,
  normalizeRecommendationData,
  normalizeSkillProfile,
} from '../skillProfile'

describe('skillProfile utils', () => {
  it('应该把后端能力画像转换为前端结构', () => {
    const result = normalizeSkillProfile({
      user_id: 1,
      updated_at: '2026-03-07T12:00:00Z',
      dimensions: [
        { dimension: 'web', score: 0.75 },
        { dimension: 'crypto', score: 0.32 },
      ],
    })

    expect(result.updated_at).toBe('2026-03-07T12:00:00Z')
    expect(result.dimensions).toEqual([
      { key: 'web', name: 'Web', value: 75 },
      { key: 'crypto', name: '密码', value: 32 },
    ])
  })

  it('应该把推荐数据转换为统一契约，并保留后端给出的薄弱维度', () => {
    const result = normalizeRecommendationData({
      weak_dimensions: [
        {
          dimension: 'crypto',
          severity: 'warning',
          confidence: 0.82,
          evidence: '密码维度已形成低分和足量训练证据。',
        },
      ],
      challenges: [
        {
          challenge_id: 12,
          title: 'crypto-lab',
          category: 'crypto',
          difficulty: 'medium',
          summary: '补强密码维度',
          evidence: '当前密码维度已有足够训练证据，适合继续补基础。',
        },
      ],
    })

    expect(result.weak_dimensions).toEqual([
      {
        dimension: 'crypto',
        label: '密码',
        severity: 'warning',
        confidence: 0.82,
        evidence: '密码维度已形成低分和足量训练证据。',
      },
    ])
    expect(getWeakDimensionLabels(result.weak_dimensions)).toEqual(['密码'])
    expect(result.challenges).toEqual([
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        summary: '补强密码维度',
        evidence: '当前密码维度已有足够训练证据，适合继续补基础。',
      },
    ])
  })
})
