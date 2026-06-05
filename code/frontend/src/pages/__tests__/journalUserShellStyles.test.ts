import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import challengeListSource from '@/pages/challenges/ChallengeListRoutePage.vue?raw'
import challengeDetailSource from '@/pages/challenges/ChallengeDetailRoutePage.vue?raw'
import challengeDetailWorkspaceSource from '@/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue?raw'
import contestListSource from '@/pages/contests/ContestListRoutePage.vue?raw'
import contestDetailSource from '@/pages/contests/ContestDetailRoutePage.vue?raw'
import contestListWorkspaceSource from '@/widgets/contest-list-workspace/ContestListWorkspace.vue?raw'
import contestDetailWorkspaceSource from '@/widgets/contest-detail-workspace/ContestDetailWorkspace.vue?raw'
import instanceListWorkspaceShellSource from '@/features/instance-list/ui/InstanceListWorkspaceShell.vue?raw'
import instanceListSource from '@/pages/instances/InstanceListRoutePage.vue?raw'
import notificationDetailSource from '@/pages/notifications/NotificationDetailRoutePage.vue?raw'
import notificationListSource from '@/pages/notifications/NotificationListRoutePage.vue?raw'
import notificationDetailWorkspaceSource from '@/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue?raw'
import notificationListWorkspaceSource from '@/widgets/notification-list-workspace/NotificationListWorkspace.vue?raw'
import securitySettingsSource from '@/pages/profile/SecuritySettingsRoutePage.vue?raw'
import securitySettingsWorkspaceShellSource from '@/features/profile/ui/SecuritySettingsWorkspaceShell.vue?raw'
import skillProfileSource from '@/pages/profile/SkillProfileRoutePage.vue?raw'
import skillProfileWorkspaceShellSource from '@/features/skill-profile/ui/SkillProfileWorkspaceShell.vue?raw'
import userProfileSource from '@/pages/profile/UserProfileRoutePage.vue?raw'
import userProfileWorkspaceShellSource from '@/features/profile/ui/UserProfileWorkspaceShell.vue?raw'
import scoreboardSource from '@/pages/scoreboard/ScoreboardViewRoutePage.vue?raw'
import categoryProgressSource from '@/features/student-dashboard/ui/StudentCategoryProgressContent.vue?raw'
import difficultyPageSource from '@/features/student-dashboard/ui/StudentDifficultyContent.vue?raw'
import recommendationPageSource from '@/features/student-dashboard/ui/StudentRecommendationContent.vue?raw'
import overviewPageSource from '@/features/student-dashboard/ui/StudentOverviewContent.vue?raw'
import trainingTimelineSource from '@/entities/training-timeline/ui/TrainingTimelinePanel.vue?raw'
import dashboardViewSource from '@/pages/dashboard/DashboardRoutePage.vue?raw'

const journalUserShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-shell.css`,
  'utf-8'
)
const surfaceShellBackgroundSource = readFileSync(
  `${process.cwd()}/src/assets/styles/surface-shell-background.css`,
  'utf-8'
)
const workspaceShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/workspace-shell.css`,
  'utf-8'
)
const challengeDetailWorkspaceCombinedSource = `${challengeDetailSource}\n${challengeDetailWorkspaceSource}`
const contestListWorkspaceCombinedSource = `${contestListSource}\n${contestListWorkspaceSource}`
const contestDetailWorkspaceCombinedSource = `${contestDetailSource}\n${contestDetailWorkspaceSource}`
const instanceListWorkspaceSource = `${instanceListSource}\n${instanceListWorkspaceShellSource}`
const notificationDetailWorkspaceCombinedSource = `${notificationDetailSource}\n${notificationDetailWorkspaceSource}`
const notificationListWorkspaceCombinedSource = `${notificationListSource}\n${notificationListWorkspaceSource}`
const securitySettingsWorkspaceSource = `${securitySettingsSource}\n${securitySettingsWorkspaceShellSource}`
const skillProfileWorkspaceSource = `${skillProfileSource}\n${skillProfileWorkspaceShellSource}`
const userProfileWorkspaceSource = `${userProfileSource}\n${userProfileWorkspaceShellSource}`

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('journal user shell shared styles', () => {
  it('应该在共享样式文件中声明学生侧与 profile 页复用的 shell 与 hero 规则', () => {
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user')
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user.journal-hero')
    expect(journalUserShellSource).toContain('--journal-shell-accent')
    expect(journalUserShellSource).toContain(
      "[data-theme='dark'] .journal-shell.journal-shell-user"
    )
    expect(surfaceShellBackgroundSource).toContain(
      '.journal-soft-surface .journal-shell.journal-hero'
    )
    expect(surfaceShellBackgroundSource).toContain(
      "[data-theme='dark'] .journal-shell.journal-shell-user.journal-hero"
    )
    expect(surfaceShellBackgroundSource).toContain(
      "[data-theme='dark'] .journal-soft-surface .journal-shell.journal-hero"
    )
    expect(surfaceShellBackgroundSource).toMatch(
      /background:\s*radial-gradient\([\s\S]*linear-gradient\(180deg,\s*var\(--surface-shell-top\),\s*var\(--surface-shell-end\)\);/s
    )
  })

  it('列表页和 profile 页应通过 journal-shell-user 接入共享 shell', () => {
    for (const source of [
      challengeListSource,
      challengeDetailWorkspaceCombinedSource,
      contestListWorkspaceCombinedSource,
      contestDetailWorkspaceCombinedSource,
      instanceListWorkspaceSource,
      notificationDetailWorkspaceCombinedSource,
      notificationListWorkspaceCombinedSource,
      scoreboardSource,
      securitySettingsWorkspaceSource,
      skillProfileWorkspaceSource,
      userProfileWorkspaceSource,
    ]) {
      expect(source).toContain('journal-shell-user')
    }
  })

  it('成员侧页面应把 hero 背景放在 section 根节点或软表面 root 上，而不是退回 div 包裹壳层', () => {
    const directHeroRootSources = [
      challengeListSource,
      challengeDetailWorkspaceCombinedSource,
      contestListWorkspaceCombinedSource,
      contestDetailWorkspaceCombinedSource,
      instanceListWorkspaceSource,
      notificationDetailWorkspaceCombinedSource,
      notificationListWorkspaceCombinedSource,
      scoreboardSource,
      securitySettingsWorkspaceSource,
      skillProfileWorkspaceSource,
      userProfileWorkspaceSource,
      dashboardViewSource,
    ]
    const embeddableHeroRootSources = [
      recommendationPageSource,
      categoryProgressSource,
      difficultyPageSource,
      overviewPageSource,
    ]

    for (const source of [...directHeroRootSources, ...embeddableHeroRootSources]) {
      expect(source).not.toMatch(/<div class="journal-shell/)
    }

    for (const source of directHeroRootSources) {
      expect(source).toMatch(
        /<section[\s\S]*?class="[^"]*journal-shell[^"]*journal-hero[^"]*min-h-full[^"]*"/s
      )
    }

    for (const source of embeddableHeroRootSources) {
      expect(source).toMatch(
        /<(?:section|div)[\s\S]*?class="[^"]*journal-soft-surface[^"]*flex[^"]*min-h-full[^"]*flex-1[^"]*flex-col[^"]*"/s
      )
    }
  })

  it('目标页面不应继续本地重写 hero 背景壳子', () => {
    for (const source of [
      challengeListSource,
      challengeDetailWorkspaceCombinedSource,
      contestListWorkspaceCombinedSource,
      contestDetailWorkspaceCombinedSource,
      instanceListWorkspaceSource,
      notificationDetailWorkspaceCombinedSource,
      notificationListWorkspaceCombinedSource,
      scoreboardSource,
      securitySettingsWorkspaceSource,
      skillProfileWorkspaceSource,
      userProfileWorkspaceSource,
    ]) {
      expect(extractScopedStyle(source)).not.toMatch(/^\.journal-hero\s*\{/m)
      expect(extractScopedStyle(source)).not.toMatch(
        /^:global\(\[data-theme='dark'\]\) \.journal-shell\s*\{/m
      )
      expect(extractScopedStyle(source)).not.toMatch(
        /^:global\(\[data-theme='dark'\]\) \.journal-hero\s*\{/m
      )
    }
  })

  it('profile 与 security 顶部概况应显式使用 metric-panel 类，旧共享 CSS 只保留变量桥接', () => {
    expect(userProfileWorkspaceSource).toContain('metric-panel-default-surface')
    expect(userProfileWorkspaceSource).toContain('metric-panel-grid')
    expect(userProfileWorkspaceSource).toContain('metric-panel-card')
    expect(userProfileWorkspaceSource).toContain('metric-panel-label')
    expect(userProfileWorkspaceSource).toContain('metric-panel-value')

    expect(securitySettingsWorkspaceSource).toContain('metric-panel-default-surface')
    expect(securitySettingsWorkspaceSource).toContain('metric-panel-grid')
    expect(securitySettingsWorkspaceSource).toContain('metric-panel-card')
    expect(securitySettingsWorkspaceSource).toContain('metric-panel-label')
    expect(securitySettingsWorkspaceSource).toContain('metric-panel-value')
    expect(securitySettingsWorkspaceSource).toContain('metric-panel-helper')

    expect(journalUserShellSource).toContain('--metric-panel-columns:')
    expect(journalUserShellSource).toContain('--metric-panel-value-size:')
  })

  it('学生侧共享控件应通过主题变量驱动文本与光标，而不是写死浅色模式颜色', () => {
    expect(journalUserShellSource).toContain('--ui-control-background: color-mix(')
    expect(journalUserShellSource).toContain(
      '--ui-control-color: var(--journal-ink, var(--color-text-primary));'
    )
    expect(journalUserShellSource).toContain(
      '--ui-control-placeholder: var(--journal-muted, var(--color-text-muted));'
    )
    expect(journalUserShellSource).not.toContain('--ui-control-background: #f8fafc;')
    expect(journalUserShellSource).not.toContain('--ui-control-color: #0f172a;')
    expect(workspaceShellSource).toContain(
      'caret-color: var(--ui-control-caret, var(--ui-control-color, var(--color-text-primary)));'
    )
  })
})
