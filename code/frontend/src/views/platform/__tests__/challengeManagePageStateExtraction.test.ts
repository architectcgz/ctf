import { describe, expect, it } from 'vitest'

import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'

describe('ChallengeManage page state extraction', () => {
  it('应将题目管理页面状态与交互 owner 抽到独立 composable', () => {
    expect(challengeManageSource).toContain(
      "import { useChallengeManagePage } from '@/features/platform-challenges'"
    )
    expect(challengeManageSource).toContain('} = useChallengeManagePage()')
  })
})
