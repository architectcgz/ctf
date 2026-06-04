import StudentCategoryProgressPage from './StudentCategoryProgressPage.vue'
import StudentDifficultyPage from './StudentDifficultyPage.vue'
import StudentOverviewStyleEditorial from './StudentOverviewStyleEditorial.vue'
import StudentRecommendationPage from './StudentRecommendationPage.vue'
import { TrainingTimelinePanel } from '@/entities/training-timeline'

import type { DashboardPanelKey } from '../model'

export const dashboardPanelComponents: Record<DashboardPanelKey, unknown> = {
  overview: StudentOverviewStyleEditorial,
  recommendation: StudentRecommendationPage,
  category: StudentCategoryProgressPage,
  timeline: TrainingTimelinePanel,
  difficulty: StudentDifficultyPage,
}

export function resolveDashboardPanelComponent(panelKey: DashboardPanelKey): unknown {
  return dashboardPanelComponents[panelKey]
}
