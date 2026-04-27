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
})
