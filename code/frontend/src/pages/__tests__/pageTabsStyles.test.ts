import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import userGovernancePageSource from '@/features/platform/user-management/ui/UserGovernancePage.vue?raw'
import userGovernanceOverviewPanelSource from '@/features/platform/user-management/ui/UserGovernanceOverviewPanel.vue?raw'
import userGovernanceDetailModalSource from '@/features/platform/user-management/ui/UserGovernanceDetailModal.vue?raw'
import userGovernanceImportPanelSource from '@/features/platform/user-management/ui/UserGovernanceImportPanel.vue?raw'
import classManagementSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import challengeDetailSource from '@/pages/challenges/ChallengeDetailRoutePage.vue?raw'
import challengeWorkspaceShellSource from '@/features/challenge-detail/ui/ChallengeWorkspaceShell.vue?raw'
import scoreboardWorkspaceShellSource from '@/features/scoreboard/ui/ScoreboardWorkspaceShell.vue?raw'
import skillProfileWorkspaceShellSource from '@/features/skill-profile/ui/SkillProfileWorkspaceShell.vue?raw'
import contestDetailSource from '@/pages/contests/ContestDetailRoutePage.vue?raw'
import challengeManageSource from '@/features/platform/challenges/ui/ChallengeManagePage.vue?raw'
import skillProfileSource from '@/pages/profile/SkillProfileRoutePage.vue?raw'

const themeSource = readFileSync(`${process.cwd()}/src/assets/styles/theme.css`, 'utf-8')
const pageTabsSource = readFileSync(`${process.cwd()}/src/assets/styles/page-tabs.css`, 'utf-8')
const globalStyleSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')
const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`
const userGovernanceSource = [
  userGovernancePageSource,
  userGovernanceOverviewPanelSource,
  userGovernanceDetailModalSource,
  userGovernanceImportPanelSource,
].join('\n')

describe('page tabs shared styles', () => {
  it('应该在共享样式里声明通用页签轨道样式', () => {
    expect(pageTabsSource).toContain('.top-tabs')
    expect(pageTabsSource).toContain('.top-tab')
    expect(pageTabsSource).toContain('.tab-panel')
  })

  it('应提供顶部标题、页签轨道与面板之间的共享变量入口', () => {
    expect(globalStyleSource).toContain('--workspace-topbar-tabs-gap: 0;')
    expect(globalStyleSource).toContain('--workspace-tabs-panel-gap:')
    expect(themeSource).toContain('--space-workspace-tabs-panel-gap:')
  })

  it('使用共享页签轨道的页面应改为注入变量，而不是继续本地重写整套样式', () => {
    for (const source of [
      classManagementSource,
      contestDetailSource,
      userGovernanceSource,
      challengeManageSource,
    ].filter((source) => source.includes('top-tabs'))) {
      expect(source).toContain('--page-top-tabs-gap:')
      expect(source).toContain('--page-top-tab-active-border:')
      expect(source).not.toMatch(/\.top-tabs\s*\{[^}]*display:\s*flex;/s)
      expect(source).not.toMatch(/\.top-tab\s*\{[^}]*border-bottom:\s*2px solid transparent;/s)
    }
  })

  it('workspace 顶部主页签应由共享预设提供默认变量', () => {
    expect(pageTabsSource).toContain('.workspace-tabbar.top-tabs')
    expect(pageTabsSource).toContain('--page-top-tabs-gap:')
    expect(pageTabsSource).toContain('--page-top-tabs-padding:')
    expect(pageTabsSource).toContain('--page-top-tab-min-height:')
    expect(pageTabsSource).toContain('--page-top-tab-active-border:')

    for (const source of [
      challengeWorkspaceShellSource,
      scoreboardWorkspaceShellSource,
      skillProfileWorkspaceSource,
    ].filter((source) => source.includes('class="workspace-tabbar top-tabs"'))) {
      expect(source).not.toContain('--page-top-tabs-gap:')
      expect(source).not.toContain('--page-top-tabs-padding:')
      expect(source).not.toContain('--page-top-tab-min-height:')
    }
  })
})
