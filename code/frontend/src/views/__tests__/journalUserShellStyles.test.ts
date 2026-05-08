import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'
import contestListSource from '@/views/contests/ContestList.vue?raw'
import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationDetailSource from '@/views/notifications/NotificationDetail.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'

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
      "[data-theme='dark'] .journal-shell.journal-shell-user.journal-hero"
    )
  })

  it('列表页和 profile 页应通过 journal-shell-user 接入共享 shell', () => {
    for (const source of [
      challengeListSource,
      challengeDetailSource,
      contestListSource,
      contestDetailSource,
      instanceListSource,
      notificationDetailSource,
      notificationListSource,
      scoreboardSource,
      securitySettingsSource,
      skillProfileSource,
      userProfileSource,
    ]) {
      expect(source).toContain('journal-shell-user')
    }
  })

  it('目标页面不应继续本地重写 hero 背景壳子', () => {
    for (const source of [
      challengeListSource,
      challengeDetailSource,
      contestListSource,
      contestDetailSource,
      instanceListSource,
      notificationDetailSource,
      notificationListSource,
      scoreboardSource,
      securitySettingsSource,
      skillProfileSource,
      userProfileSource,
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
    expect(userProfileSource).toContain('class="profile-summary-grid metric-panel-grid"')
    expect(userProfileSource).toContain('class="profile-summary-item metric-panel-card"')
    expect(userProfileSource).toContain('class="journal-note-label metric-panel-label"')
    expect(userProfileSource).toContain('class="profile-summary-value metric-panel-value"')

    expect(securitySettingsSource).toContain('class="security-summary-grid metric-panel-grid"')
    expect(securitySettingsSource).toContain('class="security-summary-item metric-panel-card"')
    expect(securitySettingsSource).toContain('class="journal-note-label metric-panel-label"')
    expect(securitySettingsSource).toContain('class="security-summary-value metric-panel-value')
    expect(securitySettingsSource).toContain('class="journal-note-helper metric-panel-helper"')

    expect(journalUserShellSource).toContain('--metric-panel-columns: repeat(2, minmax(0, 1fr));')
    expect(journalUserShellSource).toContain(
      '--metric-panel-border: color-mix(in srgb, var(--journal-border) 86%, transparent);'
    )
    expect(journalUserShellSource).toContain('--metric-panel-value-size: var(--font-size-0-98);')
    expect(journalUserShellSource).not.toMatch(
      /\.journal-shell\.journal-shell-user :is\(\.profile-summary-item, \.security-summary-item\)\s*\{[\s\S]*border-top:\s*1px solid/s
    )
    expect(journalUserShellSource).not.toMatch(
      /\.journal-shell\.journal-shell-user :is\(\.profile-summary-value, \.security-summary-value\)\s*\{[\s\S]*font-size:\s*var\(--font-size-0-98\)/s
    )
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
