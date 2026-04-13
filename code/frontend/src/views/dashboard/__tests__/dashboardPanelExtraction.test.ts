import { describe, expect, it } from 'vitest'

import dashboardViewSource from '../DashboardView.vue?raw'

describe('dashboard panel extraction', () => {
  it('DashboardView 应复用配置驱动的学生面板挂载，而不是手写五段 panel 组件', () => {
    expect(dashboardViewSource).toContain('v-for="tab in panelTabs"')
    expect(dashboardViewSource).toContain(':is="resolveDashboardPanelComponent(tab.key)"')
    expect(dashboardViewSource).toContain('v-bind="resolveDashboardPanelBindings(tab.key)"')
    expect(dashboardViewSource).not.toContain('<StudentOverviewPage')
    expect(dashboardViewSource).not.toContain('<StudentRecommendationPage')
    expect(dashboardViewSource).not.toContain('<StudentCategoryProgressPage')
    expect(dashboardViewSource).not.toContain('<StudentTimelinePage')
    expect(dashboardViewSource).not.toContain('<StudentDifficultyPage')
  })
})
