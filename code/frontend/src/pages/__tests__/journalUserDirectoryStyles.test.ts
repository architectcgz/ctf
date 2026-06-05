import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import challengeDirectoryPanelSource from '@/features/challenge-list/ui/ChallengeDirectoryPanel.vue?raw'
import instanceListWorkspaceShellSource from '@/features/instance-list/ui/InstanceListWorkspaceShell.vue?raw'
import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import contestListSource from '@/pages/contests/ContestListRoutePage.vue?raw'
import contestListWorkspaceSource from '@/widgets/contest-list-workspace/ContestListWorkspace.vue?raw'
import instanceListSource from '@/pages/instances/InstanceListRoutePage.vue?raw'
import notificationListSource from '@/pages/notifications/NotificationListRoutePage.vue?raw'
import notificationListWorkspaceSource from '@/widgets/notification-list-workspace/NotificationListWorkspace.vue?raw'
import securitySettingsSource from '@/pages/profile/SecuritySettingsRoutePage.vue?raw'
import securitySettingsWorkspaceShellSource from '@/features/profile/ui/SecuritySettingsWorkspaceShell.vue?raw'
import userProfileSource from '@/pages/profile/UserProfileRoutePage.vue?raw'
import userProfileWorkspaceShellSource from '@/features/profile/ui/UserProfileWorkspaceShell.vue?raw'
import scoreboardSource from '@/pages/scoreboard/ScoreboardViewRoutePage.vue?raw'
import scoreboardWorkspaceShellSource from '@/features/scoreboard/ui/ScoreboardWorkspaceShell.vue?raw'

const journalUserDirectorySource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-directory.css`,
  'utf-8'
)
const appStyleSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')

const securitySettingsWorkspaceSource = `${securitySettingsSource}\n${securitySettingsWorkspaceShellSource}`
const userProfileWorkspaceSource = `${userProfileSource}\n${userProfileWorkspaceShellSource}`
const scoreboardWorkspaceSource = `${scoreboardSource}\n${scoreboardWorkspaceShellSource}`
const instanceListWorkspaceSource = `${instanceListSource}\n${instanceListWorkspaceShellSource}`

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('journal user directory shared styles', () => {
  it('应该在共享样式文件中声明学生侧目录页复用的按钮与目录骨架入口', () => {
    expect(journalUserDirectorySource).toContain('.challenge-btn')
    expect(journalUserDirectorySource).toContain('.contest-btn')
    expect(journalUserDirectorySource).toContain('.notification-btn')
    expect(journalUserDirectorySource).toContain('.scoreboard-btn')
    expect(journalUserDirectorySource).toContain('.instance-btn')
    expect(challengeListSource).toContain('workspace-page-header')
    expect(contestListWorkspaceSource).toContain('workspace-page-header')
    expect(notificationListWorkspaceSource).toContain('workspace-page-header')
    expect(instanceListWorkspaceSource).toContain('workspace-page-header')
  })

  it('目标页面不应继续在 scoped style 中重复声明公共目录骨架与按钮基础样式', () => {
    for (const source of [challengeListSource, contestListSource, notificationListSource, scoreboardSource]) {
      const style = extractScopedStyle(source)
      expect(style).not.toMatch(/-btn\s*\{/)
    }

    expect(extractScopedStyle(challengeDirectoryPanelSource)).not.toContain(
      '.student-directory-list-heading__eyebrow'
    )
  })

  it('profile 与 security 页顶部应继续复用共享 topbar 与 summary 骨架', () => {
    expect(userProfileWorkspaceSource).toContain('workspace-page-header')
    expect(userProfileWorkspaceSource).toContain('metric-panel-default-surface')
    expect(userProfileWorkspaceSource).not.toContain('<PageHeader')

    expect(securitySettingsWorkspaceSource).toContain('workspace-page-header')
    expect(securitySettingsWorkspaceSource).toContain('metric-panel-default-surface')
    expect(securitySettingsWorkspaceSource).not.toContain('<PageHeader')
  })

  it('学生侧列表壳应继续复用共享间距与标题结构', () => {
    expect(appStyleSource).toContain('.student-directory-section > .student-directory-shell')
    expect(appStyleSource).toContain('.student-directory-shell__head')
    expect(challengeDirectoryPanelSource).toContain(
      'class="student-directory-shell__head student-directory-list-heading list-heading"'
    )
    expect(contestListWorkspaceSource).toContain(
      'class="student-directory-shell__head student-directory-list-heading list-heading"'
    )
    expect(notificationListWorkspaceSource).toContain(
      'class="student-directory-shell__head student-directory-list-heading list-heading"'
    )
    expect(scoreboardWorkspaceSource).toContain(
      'class="student-directory-shell__head student-directory-list-heading list-heading"'
    )
  })
})
