import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import adminChallengeWorkspaceTabsSource from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue?raw'

describe('Admin ChallengeDetail workspace extraction', () => {
  it('应将题目管理页的 tab rail 与 workspace 壳层抽到独立 platform challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import AdminChallengeWorkspaceTabs from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue'"
    )
    expect(challengeDetailSource).toContain('<AdminChallengeWorkspaceTabs')
    expect(adminChallengeWorkspaceTabsSource).toContain('aria-label="题目管理视图切换"')
    expect(adminChallengeWorkspaceTabsSource).toContain('admin-challenge-panel-writeup')
    expect(adminChallengeWorkspaceTabsSource).toContain('ChallengeWriteupManagePanel')
    expect(adminChallengeWorkspaceTabsSource).toMatch(
      /\.content-pane\s*\{[\s\S]*padding-top:\s*var\(--workspace-tabs-panel-gap\);/s
    )
    expect(adminChallengeWorkspaceTabsSource).toMatch(/\.challenge-panel\s*\{[\s\S]*padding-top:\s*0;/s)
    expect(adminChallengeWorkspaceTabsSource).not.toContain('padding-top: var(--space-6);')
  })

  it('应将题目详情加载与 Flag 配置流程下沉到 feature model', () => {
    expect(challengeDetailSource).toContain(
      "import { usePlatformChallengeDetailPage } from '@/features/platform-challenge-detail'"
    )
    expect(challengeDetailSource).toContain('} = usePlatformChallengeDetailPage()')
    expect(challengeDetailSource).not.toContain("from '@/api/admin/authoring'")
    expect(challengeDetailSource).not.toContain("from '@/api/challenge'")
    expect(challengeDetailSource).not.toContain('async function saveFlagConfig()')
    expect(challengeDetailSource).not.toContain('async function loadChallenge(')
    expect(challengeDetailSource).not.toContain('function summarizeFlagConfig(')
  })
})
