import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'
import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'
import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import challengeManageSource from '@/views/platform/ChallengeManage.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'

const pageTabsSource = readFileSync(`${process.cwd()}/src/assets/styles/page-tabs.css`, 'utf-8')

describe('page tabs shared styles', () => {
  it('应该在共享样式里声明通用页签轨道样式', () => {
    expect(pageTabsSource).toContain('.top-tabs')
    expect(pageTabsSource).toContain('.top-tab')
    expect(pageTabsSource).toContain('.tab-panel')
  })

  it('使用共享页签轨道的页面应改为注入变量，而不是继续本地重写整套样式', () => {
    for (const source of [
      classManagementSource,
      challengeDetailSource,
      contestDetailSource,
      skillProfileSource,
      userGovernanceSource,
      challengeManageSource,
    ].filter((source) => source.includes('top-tabs'))) {
      expect(source).toContain('--page-top-tabs-gap:')
      expect(source).toContain('--page-top-tab-active-border:')
      expect(source).not.toMatch(/\.top-tabs\s*\{[^}]*display:\s*flex;/s)
      expect(source).not.toMatch(/\.top-tab\s*\{[^}]*border-bottom:\s*2px solid transparent;/s)
    }
  })
})
