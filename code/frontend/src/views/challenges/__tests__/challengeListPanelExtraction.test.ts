import { describe, expect, it } from 'vitest'

import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'

describe('ChallengeList panel extraction', () => {
  it('应将题目目录工作区抽到独立 challenge 组件', () => {
    expect(challengeListSource).toContain(
      "import ChallengeDirectoryPanel from '@/components/challenge/ChallengeDirectoryPanel.vue'"
    )
    expect(challengeListSource).toContain('<ChallengeDirectoryPanel')
  })
})
