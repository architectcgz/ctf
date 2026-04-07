import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import studentAnalysisSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import contestListSource from '@/views/contests/ContestList.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import scoreboardSource from '@/views/scoreboard/ScoreboardView.vue?raw'

const journalEyebrowsSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-eyebrows.css`,
  'utf-8'
)

describe('journal eyebrow shared styles', () => {
  it('应该在共享样式文件中声明文字型 eyebrow 规则', () => {
    expect(journalEyebrowsSource).toContain(
      ':is(.journal-shell, .workspace-shell).journal-eyebrow-text .journal-eyebrow'
    )
    expect(journalEyebrowsSource).toContain(
      'letter-spacing: var(--journal-eyebrow-spacing, 0.18em);'
    )
  })

  it('列表页、profile 页和 workspace 页应通过根节点 class 接入共享 eyebrow 样式', () => {
    for (const source of [
      challengeListSource,
      contestListSource,
      instanceListSource,
      notificationListSource,
      scoreboardSource,
      securitySettingsSource,
      skillProfileSource,
      userProfileSource,
      studentAnalysisSource,
    ]) {
      expect(source).toContain('journal-eyebrow-text')
      expect(source).not.toMatch(/^\.journal-eyebrow\s*\{/m)
    }
  })
})
