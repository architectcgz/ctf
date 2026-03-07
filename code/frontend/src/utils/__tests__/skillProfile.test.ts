import { describe, expect, it } from 'vitest'

import { getWeakDimensions, normalizeRecommendations, normalizeSkillProfile } from '../skillProfile'

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

    expect(result.dimensions).toEqual([
      { key: 'web', name: 'Web', value: 75 },
      { key: 'crypto', name: '密码', value: 32 },
    ])
    expect(getWeakDimensions(result)).toEqual(['密码'])
  })

  it('应该把推荐题目转换为统一 challenge_id', () => {
    const result = normalizeRecommendations({
      challenges: [
        {
          id: 12,
          title: 'crypto-lab',
          category: 'crypto',
          difficulty: 'medium',
          reason: '补强密码维度',
        },
      ],
    })

    expect(result).toEqual([
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        reason: '补强密码维度',
      },
    ])
  })
})
