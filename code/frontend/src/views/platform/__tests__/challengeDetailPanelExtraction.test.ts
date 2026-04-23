import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/views/platform/ChallengeDetail.vue?raw'

describe('Admin ChallengeDetail panel extraction', () => {
  it('应将题目详情 tab 抽到独立 platform challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import AdminChallengeProfilePanel from '@/components/platform/challenge/AdminChallengeProfilePanel.vue'"
    )
    expect(challengeDetailSource).toContain('<AdminChallengeProfilePanel')
  })
})
