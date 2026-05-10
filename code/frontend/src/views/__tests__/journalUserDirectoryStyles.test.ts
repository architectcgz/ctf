import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import challengeDirectoryPanelSource from '@/components/challenge/ChallengeDirectoryPanel.vue?raw'
import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import contestListSource from '@/views/contests/ContestList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'

const journalUserDirectorySource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-directory.css`,
  'utf-8'
)
const appStyleSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('journal user directory shared styles', () => {
  it('应该在共享样式文件中声明学生侧目录页复用的骨架与按钮规则', () => {
    expect(journalUserDirectorySource).toContain('.challenge-topbar')
    expect(journalUserDirectorySource).toContain('.contest-topbar')
    expect(journalUserDirectorySource).toContain('.profile-topbar')
    expect(journalUserDirectorySource).toContain('.security-topbar')
    expect(journalUserDirectorySource).toContain('.notification-topbar')
    expect(journalUserDirectorySource).toContain('.scoreboard-topbar')
    expect(journalUserDirectorySource).toContain('.instance-topbar')
    expect(journalUserDirectorySource).toContain('.challenge-btn')
    expect(journalUserDirectorySource).toContain('.contest-btn')
    expect(journalUserDirectorySource).toContain('.notification-btn')
    expect(journalUserDirectorySource).toContain('.scoreboard-btn')
    expect(journalUserDirectorySource).toContain('.instance-btn')
  })

  it('目标页面不应继续在 scoped style 中重复声明公共目录骨架与按钮基础样式', () => {
    const challengeStyle = extractScopedStyle(challengeListSource)
    expect(challengeStyle).not.toMatch(/^\.challenge-topbar\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-summary\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-summary-title\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-summary-grid\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-summary-item\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-summary-label\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-summary-value\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-summary-helper\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-directory-top\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-directory-title\s*\{/m)
    expect(challengeStyle).not.toMatch(/^\.challenge-btn\s*\{/m)

    const contestStyle = extractScopedStyle(contestListSource)
    expect(contestStyle).not.toMatch(/^\.contest-topbar\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-summary\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-summary-title\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-summary-grid\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-summary-item\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-summary-label\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-summary-value\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-summary-helper\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-directory-top\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-directory-title\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-directory-meta\s*\{/m)
    expect(contestStyle).not.toMatch(/^\.contest-btn\s*\{/m)

    const notificationStyle = extractScopedStyle(notificationListSource)
    expect(notificationStyle).not.toMatch(/^\.notification-topbar\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-summary\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-summary-title\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-summary-grid\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-summary-item\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-summary-label\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-summary-value\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-summary-helper\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-directory-top\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-directory-title\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-directory-meta\s*\{/m)
    expect(notificationStyle).not.toMatch(/^\.notification-btn\s*\{/m)

    const scoreboardStyle = extractScopedStyle(scoreboardSource)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-topbar\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-summary\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-summary-title\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-summary-grid\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-summary-item\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-summary-label\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-directory-top\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-directory-title\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-directory-meta\s*\{/m)
    expect(scoreboardStyle).not.toMatch(/^\.scoreboard-btn\s*\{/m)

    const instanceStyle = extractScopedStyle(instanceListSource)
    expect(instanceStyle).not.toMatch(/^\.instance-topbar\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-summary\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-summary-title\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-summary-grid\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-summary-item\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-directory-top\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-directory-title\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-directory-meta\s*\{/m)
    expect(instanceStyle).not.toMatch(/^\.instance-btn\s*\{/m)

    const userProfileStyle = extractScopedStyle(userProfileSource)
    expect(userProfileStyle).not.toMatch(/^\.profile-topbar\s*\{/m)
    expect(userProfileStyle).not.toMatch(/^\.profile-heading\s*\{/m)
    expect(userProfileStyle).not.toMatch(/^\.profile-summary\s*\{/m)
    expect(userProfileStyle).not.toMatch(/^\.profile-summary-title\s*\{/m)
    expect(userProfileStyle).not.toMatch(/^\.profile-summary-grid\s*\{/m)
    expect(userProfileStyle).not.toMatch(/^\.profile-summary-item\s*\{/m)

    const securityStyle = extractScopedStyle(securitySettingsSource)
    expect(securityStyle).not.toMatch(/^\.security-topbar\s*\{/m)
    expect(securityStyle).not.toMatch(/^\.security-heading\s*\{/m)
    expect(securityStyle).not.toMatch(/^\.security-summary\s*\{/m)
    expect(securityStyle).not.toMatch(/^\.security-summary-title\s*\{/m)
    expect(securityStyle).not.toMatch(/^\.security-summary-grid\s*\{/m)
    expect(securityStyle).not.toMatch(/^\.security-summary-item\s*\{/m)
  })

  it('challenge 等目录页的概况卡片不应在目录共享样式里继续覆写 metric-panel 单卡外观', () => {
    expect(challengeListSource).toContain('class="challenge-summary-grid metric-panel-grid"')
    expect(challengeListSource).toContain('class="challenge-summary-item metric-panel-card"')
    expect(challengeListSource).toContain('class="challenge-summary-label metric-panel-label"')
    expect(challengeListSource).toContain('class="challenge-summary-value metric-panel-value"')
    expect(challengeListSource).toContain('class="challenge-summary-helper metric-panel-helper"')

    expect(journalUserDirectorySource).toContain('.challenge-summary-grid')
    expect(journalUserDirectorySource).not.toMatch(
      /:is\(\s*\.challenge-summary-item,[\s\S]*?--metric-panel-border:/s
    )
    expect(journalUserDirectorySource).not.toMatch(
      /:is\(\s*\.challenge-summary-label,[\s\S]*?--metric-panel-label-size:/s
    )
    expect(journalUserDirectorySource).not.toMatch(
      /:is\(\s*\.challenge-summary-value,[\s\S]*?--metric-panel-value-size:/s
    )
    expect(journalUserDirectorySource).not.toMatch(
      /:is\(\s*\.challenge-summary-helper,[\s\S]*?--metric-panel-helper-size:/s
    )
  })

  it('profile 与 security 页顶部也应接入共享 topbar 与 summary 骨架', () => {
    expect(userProfileSource).toContain('class="profile-topbar"')
    expect(userProfileSource).toContain('class="profile-topbar-meta"')
    expect(userProfileSource).toContain('class="profile-summary metric-panel-default-surface"')
    expect(userProfileSource).toContain('class="profile-summary-title"')
    expect(userProfileSource).toContain('class="profile-summary-grid metric-panel-grid"')
    expect(userProfileSource).toContain('class="workspace-overline">Profile</div>')
    expect(userProfileSource).not.toContain('<PageHeader')

    expect(securitySettingsSource).toContain('class="security-topbar"')
    expect(securitySettingsSource).toContain('class="security-topbar-meta"')
    expect(securitySettingsSource).toContain('class="security-summary metric-panel-default-surface"')
    expect(securitySettingsSource).toContain('class="security-summary-title"')
    expect(securitySettingsSource).toContain('class="security-summary-grid metric-panel-grid"')
    expect(securitySettingsSource).toContain('class="workspace-overline">Security</div>')
    expect(securitySettingsSource).not.toContain('<PageHeader')
  })

  it('学生侧列表外框不应叠加 workspace-directory-list 的默认上间距', () => {
    expect(appStyleSource).toContain('.student-directory-section > .student-directory-shell')
    expect(appStyleSource).toMatch(
      /\.student-directory-section\s*>\s*\.student-directory-shell\s*\{[^}]*margin-top:\s*0;/s
    )
    expect(contestListSource).toContain(
      'class="student-directory-shell contest-directory workspace-directory-list"'
    )
    expect(scoreboardSource).toContain(
      'class="student-directory-shell scoreboard-directory workspace-directory-list"'
    )
  })

  it('学生侧列表 item 不应被 student-directory-shell 的 grid gap 额外撑开', () => {
    expect(appStyleSource).toMatch(/\.student-directory-shell\s*\{[^}]*gap:\s*0;/s)
    expect(appStyleSource).toMatch(
      /\.student-directory-shell__head[\s\S]*\+\s*:where\([\s\S]*\.workspace-directory-grid-head[\s\S]*\)\s*\{[^}]*margin-top:\s*var\(--space-4\);/s
    )
    expect(contestListSource).toContain('class="workspace-directory-grid-row contest-row"')
    expect(scoreboardSource).toContain(
      'class="workspace-directory-grid-row scoreboard-card scoreboard-card-link"'
    )
  })

  it('学生侧筛选区与列表项之间应使用通用间距规则', () => {
    expect(appStyleSource).toMatch(
      /\.student-directory-filters[\s\S]*\+\s*:where\([\s\S]*\.workspace-directory-grid-head[\s\S]*\.notification-directory[\s\S]*\)\s*\{[^}]*margin-top:\s*var\(--space-4\);/s
    )
    expect(challengeDirectoryPanelSource).toContain(
      'class="student-directory-filters challenge-directory-filters"'
    )
    expect(notificationListSource).toContain(
      'class="student-directory-filters notification-filter-section"'
    )
    expect(contestListSource).toContain('class="student-directory-filters contest-directory-filters"')
    expect(scoreboardSource).toContain(
      'class="student-directory-filters scoreboard-directory-filters"'
    )
  })

  it('学生侧列表标题区应复用 challenge 目录的通用 header 结构', () => {
    const expectedHeaderClass =
      'class="student-directory-shell__head student-directory-list-heading list-heading"'
    const expectedHeadingClass =
      'class="student-directory-shell__heading student-directory-list-heading__body"'
    const expectedEyebrowClass =
      'class="journal-note-label student-directory-shell__eyebrow student-directory-list-heading__eyebrow"'
    const expectedTitleClass =
      'class="student-directory-shell__title student-directory-list-heading__title"'

    expect(appStyleSource).toContain('.student-directory-list-heading__eyebrow')
    expect(appStyleSource).toMatch(
      /\.student-directory-list-heading__eyebrow\s*\{[^}]*color:\s*color-mix\(in srgb, var\(--color-primary\) 72%, var\(--journal-muted\)\);/s
    )

    for (const source of [
      challengeDirectoryPanelSource,
      contestListSource,
      notificationListSource,
      scoreboardSource,
    ]) {
      expect(source).toContain(expectedHeaderClass)
      expect(source).toContain(expectedHeadingClass)
      expect(source).toContain(expectedEyebrowClass)
      expect(source).toContain(expectedTitleClass)
    }

    expect(challengeDirectoryPanelSource).not.toContain('challenge-directory-shell__head')
    expect(challengeDirectoryPanelSource).not.toContain('challenge-directory-shell__heading')
    expect(challengeDirectoryPanelSource).not.toContain('challenge-directory-shell__eyebrow')
    expect(challengeDirectoryPanelSource).not.toContain('challenge-directory-shell__title')
    expect(extractScopedStyle(challengeDirectoryPanelSource)).not.toContain(
      '.student-directory-list-heading__eyebrow'
    )
  })
})
