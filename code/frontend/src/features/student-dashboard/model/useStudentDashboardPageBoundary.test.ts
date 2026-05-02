import { describe, expect, it } from 'vitest'

import source from './useStudentDashboardPage.ts?raw'

describe('useStudentDashboardPage boundary', () => {
  it('应组合数据加载与面板绑定子模块，避免主组合器内联查询与绑定拼装', () => {
    expect(source).toContain("from './useStudentDashboardData'")
    expect(source).toContain("from './useStudentDashboardPanelBindings'")
    expect(source).not.toContain("from '@/api/assessment'")
    expect(source).not.toContain('async function loadDashboard()')
    expect(source).not.toContain('function resolveDashboardPanelBindings(')
  })
})
