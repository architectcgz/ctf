import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'

describe('ChallengeDetail panel extraction', () => {
  it('应将题目、题解、提交记录、题解编辑和右侧工具区抽到独立 challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import { useChallengeDetailPage } from '@/features/challenge-detail'"
    )
    expect(challengeDetailSource).not.toContain("from '@/api/challenge'")
    expect(challengeDetailSource).toContain(
      "import ChallengeQuestionPanel from '@/components/challenge/ChallengeQuestionPanel.vue'"
    )
    expect(challengeDetailSource).toContain(
      "import ChallengeSolutionsPanel from '@/components/challenge/ChallengeSolutionsPanel.vue'"
    )
    expect(challengeDetailSource).toContain(
      "import ChallengeSubmissionRecordsPanel from '@/components/challenge/ChallengeSubmissionRecordsPanel.vue'"
    )
    expect(challengeDetailSource).toContain(
      "import ChallengeWriteupPanel from '@/components/challenge/ChallengeWriteupPanel.vue'"
    )
    expect(challengeDetailSource).toContain(
      "import ChallengeActionAside from '@/components/challenge/ChallengeActionAside.vue'"
    )
    expect(challengeDetailSource).toContain('<ChallengeQuestionPanel')
    expect(challengeDetailSource).toContain('<ChallengeSolutionsPanel')
    expect(challengeDetailSource).toContain('<ChallengeSubmissionRecordsPanel')
    expect(challengeDetailSource).toContain('<ChallengeWriteupPanel')
    expect(challengeDetailSource).toContain('<ChallengeActionAside')
  })
})
