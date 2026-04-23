import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'

describe('ChallengeDetail panel extraction', () => {
  it('应将题解、提交记录和题解编辑面板抽到独立 challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import ChallengeSolutionsPanel from '@/components/challenge/ChallengeSolutionsPanel.vue'"
    )
    expect(challengeDetailSource).toContain(
      "import ChallengeSubmissionRecordsPanel from '@/components/challenge/ChallengeSubmissionRecordsPanel.vue'"
    )
    expect(challengeDetailSource).toContain(
      "import ChallengeWriteupPanel from '@/components/challenge/ChallengeWriteupPanel.vue'"
    )
    expect(challengeDetailSource).toContain('<ChallengeSolutionsPanel')
    expect(challengeDetailSource).toContain('<ChallengeSubmissionRecordsPanel')
    expect(challengeDetailSource).toContain('<ChallengeWriteupPanel')
  })
})
