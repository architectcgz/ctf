import StudentCategoryProgressPage from './StudentCategoryProgressPage.vue'
import StudentDifficultyPage from './StudentDifficultyPage.vue'
import StudentOverviewPage from './StudentOverviewPage.vue'
import StudentRecommendationPage from './StudentRecommendationPage.vue'
import { TrainingTimelinePanel } from '@/entities/training-timeline'

import type { DashboardPanelKey } from '../model'

export const dashboardPanelComponents: Record<DashboardPanelKey, unknown> = {
  overview: StudentOverviewPage,
  recommendation: StudentRecommendationPage,
  category: StudentCategoryProgressPage,
  timeline: TrainingTimelinePanel,
  difficulty: StudentDifficultyPage,
}

export function resolveDashboardPanelComponent(panelKey: DashboardPanelKey): unknown {
  return dashboardPanelComponents[panelKey]
}
