import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/views/platform/ChallengeDetail.vue?raw'
import adminChallengeTopbarPanelSource from '@/components/platform/challenge/AdminChallengeTopbarPanel.vue?raw'
import adminChallengeWorkspaceTabsSource from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue?raw'
import adminChallengeProfilePanelSource from '@/components/platform/challenge/AdminChallengeProfilePanel.vue?raw'
import challengeProfileMetaGridSource from '@/entities/challenge/ui/ChallengeProfileMetaGrid.vue?raw'

describe('Admin ChallengeDetail panel extraction', () => {
  it('应将题目详情 tab 抽到独立 platform challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import AdminChallengeWorkspaceTabs from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue'"
    )
    expect(challengeDetailSource).toContain('<AdminChallengeWorkspaceTabs')
    expect(adminChallengeWorkspaceTabsSource).toContain(
      "import AdminChallengeProfilePanel from '@/components/platform/challenge/AdminChallengeProfilePanel.vue'"
    )
    expect(adminChallengeWorkspaceTabsSource).toContain('<AdminChallengeProfilePanel')
  })

  it('应将题目详情顶栏抽到独立 platform challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import AdminChallengeTopbarPanel from '@/components/platform/challenge/AdminChallengeTopbarPanel.vue'"
    )
    expect(challengeDetailSource).toContain('<AdminChallengeTopbarPanel')
    expect(adminChallengeTopbarPanelSource).toContain('<span class="workspace-overline">Challenge Profile</span>')
    expect(adminChallengeTopbarPanelSource).toContain('拓扑编排')
    expect(adminChallengeTopbarPanelSource).toContain('返回题库')
  })

  it('题目详情概览应复用 challenge entity 的分类与难度文本单元', () => {
    expect(adminChallengeProfilePanelSource).toContain("from '@/entities/challenge'")
    expect(adminChallengeProfilePanelSource).toContain('<ChallengeProfileSummaryStrip')
    expect(adminChallengeProfilePanelSource).not.toContain('class="challenge-overview-summary')
  })

  it('题目详情基础信息应复用 challenge entity 的元信息网格单元', () => {
    expect(adminChallengeProfilePanelSource).toContain("from '@/entities/challenge'")
    expect(adminChallengeProfilePanelSource).toContain('<ChallengeProfileMetaGrid')
    expect(adminChallengeProfilePanelSource).not.toContain('<dl class="challenge-meta-grid">')
    expect(challengeProfileMetaGridSource).toContain('<dl class="challenge-meta-grid">')
    expect(challengeProfileMetaGridSource).toContain("getChallengeInstanceSharingLabel(challenge.instance_sharing)")
    expect(challengeProfileMetaGridSource).toContain("formatChallengeDateTime(challenge.created_at)")
  })
})
