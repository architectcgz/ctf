import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

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
    expect(userProfileSource).toContain('class="profile-summary"')
    expect(userProfileSource).toContain('class="profile-summary-title"')
    expect(userProfileSource).toContain('class="profile-summary-grid metric-panel-grid"')
    expect(userProfileSource).toContain('<PageHeader')

    expect(securitySettingsSource).toContain('class="security-topbar"')
    expect(securitySettingsSource).toContain('class="security-summary"')
    expect(securitySettingsSource).toContain('class="security-summary-title"')
    expect(securitySettingsSource).toContain('class="security-summary-grid metric-panel-grid"')
    expect(securitySettingsSource).toContain('<PageHeader')
  })
})
