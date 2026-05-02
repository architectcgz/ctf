import StudentCategoryProgressPage from './StudentCategoryProgressPage.vue'
import StudentDifficultyPage from './StudentDifficultyPage.vue'
import StudentOverviewPage from './StudentOverviewPage.vue'
import StudentRecommendationPage from './StudentRecommendationPage.vue'
import StudentTimelinePage from './StudentTimelinePage.vue'

import type { DashboardPanelKey } from '@/features/student-dashboard'

export const dashboardPanelComponents: Record<DashboardPanelKey, unknown> = {
  overview: StudentOverviewPage,
  recommendation: StudentRecommendationPage,
  category: StudentCategoryProgressPage,
  timeline: StudentTimelinePage,
  difficulty: StudentDifficultyPage,
}

export function resolveDashboardPanelComponent(panelKey: DashboardPanelKey): unknown {
  return dashboardPanelComponents[panelKey]
}
