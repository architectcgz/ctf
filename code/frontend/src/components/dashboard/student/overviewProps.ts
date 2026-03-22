import type { MyProgressData, RecommendationItem, SkillDimensionScore, TimelineEvent } from '@/api/contracts'

import type { DashboardHighlightItem } from './types'

export interface StudentOverviewProps {
  displayName: string
  className?: string
  progress: MyProgressData
  completionRate: number
  highlightItems: DashboardHighlightItem[]
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  weakDimensions: string[]
  skillDimensions: SkillDimensionScore[]
}
