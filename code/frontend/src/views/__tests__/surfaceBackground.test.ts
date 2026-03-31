import { describe, expect, it } from 'vitest'
import categoryProgressSource from '../../components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import difficultyPageSource from '../../components/dashboard/student/StudentDifficultyPage.vue?raw'
import overviewPageSource from '../../components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import recommendationPageSource from '../../components/dashboard/student/StudentRecommendationPage.vue?raw'
import timelinePageSource from '../../components/dashboard/student/StudentTimelinePage.vue?raw'
import notificationListSource from '../notifications/NotificationList.vue?raw'
import securitySettingsSource from '../profile/SecuritySettings.vue?raw'
import userProfileSource from '../profile/UserProfile.vue?raw'

const lightHeroBackgroundPattern =
  /background:\s*radial-gradient\(circle at top right, rgba\(37, 99, 235, 0\.08\), transparent 18rem\),\s*linear-gradient\(180deg, #ffffff, #f8fafc\);/s

describe('member-facing page surfaces', () => {
  it('should use the same white hero background as admin dashboard', () => {
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
      expect(source).toMatch(lightHeroBackgroundPattern)
    }
  })
})
