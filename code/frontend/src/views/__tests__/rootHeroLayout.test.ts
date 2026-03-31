import { describe, expect, it } from 'vitest'

import challengeListSource from '../challenges/ChallengeList.vue?raw'
import contestListSource from '../contests/ContestList.vue?raw'
import securitySettingsSource from '../profile/SecuritySettings.vue?raw'
import scoreboardViewSource from '../scoreboard/ScoreboardView.vue?raw'
import categoryProgressSource from '../../components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import difficultyPageSource from '../../components/dashboard/student/StudentDifficultyPage.vue?raw'
import overviewPageSource from '../../components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import recommendationPageSource from '../../components/dashboard/student/StudentRecommendationPage.vue?raw'
import timelinePageSource from '../../components/dashboard/student/StudentTimelinePage.vue?raw'

describe('full-bleed hero roots', () => {
  it('uses a section root that carries the hero background', () => {
    const sources = [
      challengeListSource,
      contestListSource,
      scoreboardViewSource,
      securitySettingsSource,
      recommendationPageSource,
      categoryProgressSource,
      timelinePageSource,
      difficultyPageSource,
      overviewPageSource,
    ]

    for (const source of sources) {
      expect(source).not.toMatch(/<div class="journal-shell/)
      expect(source).toMatch(/<section class="journal-shell[^"]*journal-hero[^"]*min-h-full/s)
    }
  })
})
