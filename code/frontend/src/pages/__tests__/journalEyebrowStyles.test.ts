import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import studentAnalysisSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue?raw'
import instanceListWorkspaceShellSource from '@/components/instance/InstanceListWorkspaceShell.vue?raw'
import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import contestListSource from '@/pages/contests/ContestListRoutePage.vue?raw'
import instanceListSource from '@/pages/instances/InstanceListRoutePage.vue?raw'
import notificationListSource from '@/pages/notifications/NotificationListRoutePage.vue?raw'
import securitySettingsSource from '@/pages/profile/SecuritySettingsRoutePage.vue?raw'
import securitySettingsWorkspaceShellSource from '@/components/profile/SecuritySettingsWorkspaceShell.vue?raw'
import skillProfileSource from '@/pages/profile/SkillProfileRoutePage.vue?raw'
import skillProfileWorkspaceShellSource from '@/components/profile/SkillProfileWorkspaceShell.vue?raw'
import userProfileSource from '@/pages/profile/UserProfileRoutePage.vue?raw'
import userProfileWorkspaceShellSource from '@/components/profile/UserProfileWorkspaceShell.vue?raw'
import scoreboardSource from '@/pages/scoreboard/ScoreboardViewRoutePage.vue?raw'

const journalEyebrowsSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-eyebrows.css`,
  'utf-8'
)
const instanceListWorkspaceSource = `${instanceListSource}\n${instanceListWorkspaceShellSource}`
const securitySettingsWorkspaceSource = `${securitySettingsSource}\n${securitySettingsWorkspaceShellSource}`
const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`
const userProfileWorkspaceSource = `${userProfileSource}\n${userProfileWorkspaceShellSource}`

describe('journal eyebrow shared styles', () => {
  it('应该在共享样式文件中声明文字型 eyebrow 规则', () => {
    expect(journalEyebrowsSource).toContain(
      ':is(.journal-shell, .workspace-shell).journal-eyebrow-text .journal-eyebrow'
    )
    expect(journalEyebrowsSource).toContain(
      'letter-spacing: var(--journal-eyebrow-spacing, 0.18em);'
    )
  })

  it('仍使用 journal eyebrow 的 workspace 页应通过根节点 class 接入共享样式', () => {
    expect(studentAnalysisSource).toContain('journal-eyebrow-text')
    expect(studentAnalysisSource).not.toMatch(/^\.journal-eyebrow\s*\{/m)
  })

  it('已切到 workspace overline 的页面不应继续携带旧 eyebrow 根节点修饰类', () => {
    for (const source of [
      challengeListSource,
      contestListSource,
      instanceListWorkspaceSource,
      notificationListSource,
      scoreboardSource,
      securitySettingsWorkspaceSource,
      skillProfileWorkspaceSource,
      userProfileWorkspaceSource,
    ]) {
      expect(source).not.toContain('journal-eyebrow-text')
      expect(source).not.toMatch(/^\.journal-eyebrow\s*\{/m)
    }
  })
})
