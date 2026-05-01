import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/views/platform/ChallengeDetail.vue?raw'
import adminChallengeTopbarPanelSource from '@/components/platform/challenge/AdminChallengeTopbarPanel.vue?raw'
import adminChallengeWorkspaceTabsSource from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue?raw'
import adminChallengeProfilePanelSource from '@/components/platform/challenge/AdminChallengeProfilePanel.vue?raw'
import challengeProfileMetaGridSource from '@/entities/challenge/ui/ChallengeProfileMetaGrid.vue?raw'
import platformChallengeFlagConfigPanelSource from '@/features/platform-challenge-detail/ui/PlatformChallengeFlagConfigPanel.vue?raw'
import platformChallengeFlagActionBarSource from '@/features/platform-challenge-detail/ui/PlatformChallengeFlagActionBar.vue?raw'
import platformChallengeFlagFieldGridSource from '@/features/platform-challenge-detail/ui/PlatformChallengeFlagFieldGrid.vue?raw'
import platformChallengeFlagNoticeStackSource from '@/features/platform-challenge-detail/ui/PlatformChallengeFlagNoticeStack.vue?raw'
import platformChallengeDetailWorkspaceSource from '@/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue?raw'

describe('Admin ChallengeDetail panel extraction', () => {
  it('应将题目详情 tab 抽到独立 platform challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import { PlatformChallengeDetailWorkspace } from '@/widgets/platform-challenge-detail'"
    )
    expect(challengeDetailSource).toContain('<PlatformChallengeDetailWorkspace')
    expect(platformChallengeDetailWorkspaceSource).toContain(
      "import AdminChallengeWorkspaceTabs from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue'"
    )
    expect(platformChallengeDetailWorkspaceSource).toContain('<AdminChallengeWorkspaceTabs')
    expect(adminChallengeWorkspaceTabsSource).toContain(
      "import AdminChallengeProfilePanel from '@/components/platform/challenge/AdminChallengeProfilePanel.vue'"
    )
    expect(adminChallengeWorkspaceTabsSource).toContain('<AdminChallengeProfilePanel')
  })

  it('应将题目详情顶栏抽到独立 platform challenge 组件', () => {
    expect(platformChallengeDetailWorkspaceSource).toContain(
      "import AdminChallengeTopbarPanel from '@/components/platform/challenge/AdminChallengeTopbarPanel.vue'"
    )
    expect(platformChallengeDetailWorkspaceSource).toContain('<AdminChallengeTopbarPanel')
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

  it('题目详情判题配置应复用 platform challenge detail feature ui', () => {
    expect(adminChallengeProfilePanelSource).toContain("from '@/features/platform-challenge-detail'")
    expect(adminChallengeProfilePanelSource).not.toContain('@/features/platform-challenge-detail/model/')
    expect(adminChallengeProfilePanelSource).not.toContain('@/features/platform-challenge-detail/ui/')
    expect(adminChallengeProfilePanelSource).toContain('<PlatformChallengeFlagConfigPanel')
    expect(adminChallengeProfilePanelSource).not.toContain('<section class="journal-panel challenge-flag-panel')
    expect(platformChallengeFlagConfigPanelSource).toContain(
      '<section class="journal-panel challenge-flag-panel p-5 md:p-6">'
    )
    expect(platformChallengeFlagConfigPanelSource).toContain('<PlatformChallengeFlagFieldGrid')
    expect(platformChallengeFlagConfigPanelSource).toContain('<PlatformChallengeFlagNoticeStack')
    expect(platformChallengeFlagConfigPanelSource).toContain('<PlatformChallengeFlagActionBar')
    expect(platformChallengeFlagFieldGridSource).toContain('正则表达式')
    expect(platformChallengeFlagNoticeStackSource).toContain('共享实例只适用于无状态题')
    expect(platformChallengeFlagActionBarSource).toContain("@click=\"emit('save')\"")
  })
})
