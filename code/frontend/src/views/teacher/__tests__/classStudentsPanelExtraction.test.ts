import { describe, expect, it } from 'vitest'

import classStudentsPageSourceBase from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/components/teacher/class-management/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/components/teacher/class-management/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/components/teacher/class-management/ClassStudentsDirectoryPanel.vue?raw'

const classStudentsPageSource = [
  classStudentsPageSourceBase,
  classStudentsOverviewPanelSource,
  classStudentsInsightWindowPanelSource,
  classStudentsDirectoryPanelSource,
].join('\n')

describe('class students panel extraction', () => {
  it('ClassStudentsPage 应复用配置驱动的 panel tabs，而不是为趋势复盘等标准面板重复 section 壳层', () => {
    expect(classStudentsPageSource).toContain('v-for="tab in panelWorkspaceTabs"')
    expect(classStudentsPageSource).toContain(':is="resolveWorkspacePanelComponent(tab.key)"')
    expect(classStudentsPageSource).toContain('v-bind="resolveWorkspacePanelProps(tab.key)"')
    expect(classStudentsPageSource).not.toContain('<TeacherClassTrendPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherClassReviewPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherClassInsightsPanel')
    expect(classStudentsPageSource).not.toContain('<TeacherInterventionPanel')
  })
})
