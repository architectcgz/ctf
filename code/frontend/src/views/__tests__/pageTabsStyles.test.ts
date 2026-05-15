import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'
import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'
import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'

const themeSource = readFileSync(`${process.cwd()}/src/assets/styles/theme.css`, 'utf-8')
const pageTabsSource = readFileSync(`${process.cwd()}/src/assets/styles/page-tabs.css`, 'utf-8')
const globalStyleSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')

describe('page tabs shared styles', () => {
  it('应该在共享样式里声明通用页签轨道样式', () => {
    expect(pageTabsSource).toContain('.top-tabs')
    expect(pageTabsSource).toContain('.top-tab')
    expect(pageTabsSource).toContain('.tab-panel')
  })

  it('应提供顶部标题、页签轨道与面板之间的全局间距语义变量', () => {
    expect(globalStyleSource).toContain('--workspace-topbar-tabs-gap: 0;')
    expect(globalStyleSource).toContain('--workspace-tabs-panel-gap: var(--space-workspace-tabs-panel-gap);')
    expect(themeSource).toContain('--space-workspace-tabs-panel-gap: var(--space-3-5);')
    expect(pageTabsSource).toContain(
      'padding-bottom: var(--journal-topbar-padding-bottom, var(--workspace-topbar-tabs-gap, 0));'
    )
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
    expect(pageTabsSource).toContain('--page-top-tabs-gap: var(--space-7);')
    expect(pageTabsSource).toContain('--page-top-tabs-padding: 0 var(--space-7);')
    expect(pageTabsSource).toContain('--page-top-tab-min-height: 52px;')
    expect(pageTabsSource).toContain('--page-top-tab-active-border: var(--brand);')

    for (const source of [challengeDetailSource, skillProfileSource].filter((source) =>
      source.includes('class="workspace-tabbar top-tabs"')
    )) {
      expect(source).not.toContain('--page-top-tabs-gap: var(--space-7);')
      expect(source).not.toContain('--page-top-tabs-padding: 0 var(--space-7);')
      expect(source).not.toContain('--page-top-tab-min-height: 52px;')
    }
  })
})
