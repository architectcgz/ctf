import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'
import categoryProgressSource from '../../components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import difficultyPageSource from '../../components/dashboard/student/StudentDifficultyPage.vue?raw'
import overviewPageSource from '../../components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import recommendationPageSource from '../../components/dashboard/student/StudentRecommendationPage.vue?raw'
import timelinePageSource from '../../components/dashboard/student/StudentTimelinePage.vue?raw'
import dashboardViewSource from '../dashboard/DashboardView.vue?raw'
import notificationListSource from '../notifications/NotificationList.vue?raw'
import securitySettingsSource from '../profile/SecuritySettings.vue?raw'
import userProfileSource from '../profile/UserProfile.vue?raw'

const surfaceShellBackgroundSource = readFileSync(
  `${process.cwd()}/src/assets/styles/surface-shell-background.css`,
  'utf-8'
)

describe('member-facing page surfaces', () => {
  it('should centralize tokenized hero background formula in shared surface stylesheet', () => {
    expect(surfaceShellBackgroundSource).toContain('.journal-shell.journal-shell-user.journal-hero')
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

  it('member pages should consume shared hero shell classes instead of duplicating formula', () => {
    const userShellSources = [
      dashboardViewSource,
      userProfileSource,
      securitySettingsSource,
      notificationListSource,
    ]
    const softSurfaceSources = [
      recommendationPageSource,
      categoryProgressSource,
      difficultyPageSource,
      timelinePageSource,
      overviewPageSource,
    ]

    for (const source of userShellSources) {
      expect(source).toContain('journal-shell-user')
      expect(source).toContain('journal-hero')
    }

    for (const source of softSurfaceSources) {
      expect(source).toContain('journal-hero')
    }
  })
})
