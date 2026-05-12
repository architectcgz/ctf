import { describe, expect, it } from 'vitest'

import metricsSource from './useTeacherDashboardMetrics.ts?raw'

describe('useTeacherDashboardMetrics boundary', () => {
  it('应组合 overview builders，避免主模块内联大段概览与摘要列表构建', () => {
    expect(metricsSource).toContain("from './teacherDashboardOverviewBuilders'")
    expect(metricsSource).toContain('buildOverviewDescription(')
    expect(metricsSource).toContain('buildMetaPills(')
    expect(metricsSource).toContain('buildOverviewMetrics(')
    expect(metricsSource).toContain('buildReviewHighlights(')
    expect(metricsSource).toContain('buildInterventionTargets(')
  })
})
