import { describe, expect, it } from 'vitest'
import categoryProgressSource from '../../components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import difficultyPageSource from '../../components/dashboard/student/StudentDifficultyPage.vue?raw'
import overviewPageSource from '../../components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import recommendationPageSource from '../../components/dashboard/student/StudentRecommendationPage.vue?raw'
import timelinePageSource from '../../components/dashboard/student/StudentTimelinePage.vue?raw'
import notificationListSource from '../notifications/NotificationList.vue?raw'
import securitySettingsSource from '../profile/SecuritySettings.vue?raw'
import userProfileSource from '../profile/UserProfile.vue?raw'

const tokenizedHeroBackgroundPattern =
  /background:\s*radial-gradient\(circle at top right, [\s\S]*linear-gradient\(180deg,\s*color-mix\(in srgb, var\(--journal-surface(?:, var\(--color-bg-surface\))?\) 96%, var\(--color-bg-base\)\),\s*color-mix\(in srgb, var\(--journal-surface-subtle(?:, var\(--color-bg-elevated\))?\) 94%, var\(--color-bg-base\)\)\);/s

describe('member-facing page surfaces', () => {
  it('should use tokenized hero backgrounds that follow the active theme palette', () => {
    const sources = [
      userProfileSource,
      securitySettingsSource,
      notificationListSource,
      recommendationPageSource,
      categoryProgressSource,
      difficultyPageSource,
      timelinePageSource,
      overviewPageSource,
    ]

    for (const source of sources) {
      expect(source).toMatch(tokenizedHeroBackgroundPattern)
    }
  })
})
