import { describe, expect, it } from 'vitest'

import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'
import challengeManageHeroPanelSource from '@/components/platform/challenge/ChallengeManageHeroPanel.vue?raw'
import challengeManageDirectoryPanelSource from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue?raw'

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

  it('题目管理目录面板应把分类 pill 和难度文本下沉到 challenge entity ui', () => {
    expect(challengeManageDirectoryPanelSource).toContain("from '@/entities/challenge'")
    expect(challengeManageDirectoryPanelSource).toContain('<ChallengeCategoryPill')
    expect(challengeManageDirectoryPanelSource).toContain('<ChallengeDifficultyText')
    expect(challengeManageDirectoryPanelSource).not.toContain(':get-category-label=')
    expect(challengeManageDirectoryPanelSource).not.toContain(':get-difficulty-label=')
  })
})
