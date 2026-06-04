import StudentCategoryProgressPage from './StudentCategoryProgressPage.vue'
import StudentDifficultyPage from './StudentDifficultyPage.vue'
import StudentOverviewStyleEditorial from './StudentOverviewStyleEditorial.vue'
import StudentRecommendationPage from './StudentRecommendationPage.vue'
import { TrainingTimelineContent } from '@/entities/training-timeline'

import type { DashboardPanelKey } from '../model'

export const dashboardPanelComponents: Record<DashboardPanelKey, unknown> = {
  overview: StudentOverviewStyleEditorial,
  recommendation: StudentRecommendationPage,
  category: StudentCategoryProgressPage,
  timeline: TrainingTimelineContent,
  difficulty: StudentDifficultyPage,
}

export function resolveDashboardPanelComponent(panelKey: DashboardPanelKey): unknown {
  return dashboardPanelComponents[panelKey]
}
