import { describe, expect, it } from 'vitest'

import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'
import challengeManageHeroPanelSource from '@/components/platform/challenge/ChallengeManageHeroPanel.vue?raw'

describe('ChallengeManage directory extraction', () => {
  it('应将题目目录工作区抽到独立平台组件', () => {
    expect(challengeManageSource).toContain(
      "import ChallengeManageDirectoryPanel from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue'"
    )
    expect(challengeManageSource).toContain('<ChallengeManageDirectoryPanel')
  })

  it('应将题目管理头部摘要抽到独立平台组件', () => {
    expect(challengeManageSource).toContain(
      "import ChallengeManageHeroPanel from '@/components/platform/challenge/ChallengeManageHeroPanel.vue'"
    )
    expect(challengeManageSource).toContain('<ChallengeManageHeroPanel')
    expect(challengeManageHeroPanelSource).toMatch(
      /<div class="workspace-overline">\s*Challenge Workspace\s*<\/div>/
    )
    expect(challengeManageHeroPanelSource).toContain('导入题目')
    expect(challengeManageHeroPanelSource).toContain('题目总量')
  })
})
