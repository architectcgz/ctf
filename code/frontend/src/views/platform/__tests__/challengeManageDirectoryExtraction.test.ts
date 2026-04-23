import { describe, expect, it } from 'vitest'

import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'

describe('ChallengeManage directory extraction', () => {
  it('应将题目目录工作区抽到独立平台组件', () => {
    expect(challengeManageSource).toContain(
      "import ChallengeManageDirectoryPanel from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue'"
    )
    expect(challengeManageSource).toContain('<ChallengeManageDirectoryPanel')
  })
})
