import { describe, expect, it } from 'vitest'

import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'

describe('class students panel extraction', () => {
  it('ClassStudentsPage 应复用配置驱动的 panel tabs，而不是为趋势复盘等标准面板重复 section 壳层', () => {
    expect(classStudentsPageSource).toContain("v-for=\"tab in panelWorkspaceTabs\"")
    expect(classStudentsPageSource).toContain(':is="resolveWorkspacePanelComponent(tab.key)"')
    expect(classStudentsPageSource).toContain('v-bind="resolveWorkspacePanelProps(tab.key)"')
    expect(classStudentsPageSource).not.toContain('<TeacherClassTrendPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherClassReviewPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherClassInsightsPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherInterventionPanel')
  })
})
