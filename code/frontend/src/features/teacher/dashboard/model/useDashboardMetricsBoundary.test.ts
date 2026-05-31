import { describe, expect, it } from 'vitest'

import metricsSource from './useDashboardMetrics.ts?raw'
import overviewBuildersSource from './teacherDashboardOverviewBuilders.ts?raw'
import insightBuildersSource from './teacherDashboardInsightBuilders.ts?raw'
import userPresentationSource from '@/entities/user/model/presentation.ts?raw'

describe('useDashboardMetrics boundary', () => {
  it('应组合 overview builders，避免主模块内联大段概览与摘要列表构建', () => {
    expect(metricsSource).toContain("from './teacherDashboardOverviewBuilders'")
    expect(metricsSource).toContain('buildOverviewDescription(')
    expect(metricsSource).toContain('buildMetaPills(')
    expect(metricsSource).toContain('buildOverviewMetrics(')
    expect(metricsSource).toContain('buildReviewHighlights(')
    expect(metricsSource).toContain('buildInterventionTargets(')
  })

  it('教师仪表盘里的学员显示名应通过 user entity 统一承接', () => {
    expect(overviewBuildersSource).toContain("import { getUserDisplayName } from '@/entities/user'")
    expect(overviewBuildersSource).not.toContain('const displayName = student.name || student.username')
    expect(insightBuildersSource).toContain("import { getUserDisplayName } from '@/entities/user'")
    expect(insightBuildersSource).not.toContain('spotlightStudent.name || spotlightStudent.username')
    expect(userPresentationSource).toContain('getUserDisplayName')
  })
})
