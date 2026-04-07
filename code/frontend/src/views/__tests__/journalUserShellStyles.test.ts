import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
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

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('journal user shell shared styles', () => {
  it('应该在共享样式文件中声明学生侧与 profile 页复用的 shell 与 hero 规则', () => {
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user')
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user.journal-hero')
    expect(journalUserShellSource).toContain('--journal-shell-accent')
  })

  it('列表页和 profile 页应通过 journal-shell-user 接入共享 shell', () => {
    for (const source of [
      challengeListSource,
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

  it('除 skill profile 的圆角补充外，目标页面不应继续本地重写 hero 背景壳子', () => {
    for (const source of [
      challengeListSource,
      contestListSource,
      contestDetailSource,
      instanceListSource,
      notificationDetailSource,
      notificationListSource,
      scoreboardSource,
      securitySettingsSource,
      userProfileSource,
    ]) {
      expect(extractScopedStyle(source)).not.toMatch(/^\.journal-hero\s*\{/m)
    }

    const skillProfileStyle = extractScopedStyle(skillProfileSource)
    const localHeroRule = skillProfileStyle.match(/^\.journal-hero\s*\{([\s\S]*?)^\}/m)

    expect(localHeroRule).not.toBeNull()
    expect(localHeroRule?.[1]).toContain('border-radius: 16px !important;')
    expect(localHeroRule?.[1]).toContain('overflow: hidden;')
    expect(localHeroRule?.[1]).not.toContain('background:')
  })
})
