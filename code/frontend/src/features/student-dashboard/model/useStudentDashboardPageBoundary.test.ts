import { describe, expect, it } from 'vitest'

import source from './useStudentDashboardPage.ts?raw'

describe('useStudentDashboardPage boundary', () => {
  it('应组合数据加载、route target 与面板绑定子模块，避免主组合器重新持有 router transport', () => {
    expect(source).toContain("from '@/composables/routeNavigationTransport'")
    expect(source).toContain("from './useStudentDashboardData'")
    expect(source).toContain("from './useStudentDashboardPanelBindings'")
    expect(source).toContain("from './studentDashboardRoutes'")
    expect(source).not.toContain("from '@/api/assessment'")
    expect(source).not.toContain("from 'vue-router'")
    expect(source).not.toContain('useRouter(')
    expect(source).not.toContain('useRoute(')
    expect(source).not.toContain('async function loadDashboard()')
    expect(source).not.toContain('function resolveDashboardPanelBindings(')
  })
})
