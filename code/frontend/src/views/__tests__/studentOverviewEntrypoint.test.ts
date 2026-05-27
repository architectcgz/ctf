import { describe, expect, it } from 'vitest'

import studentOverviewPageSource from '@/features/student-dashboard/ui/StudentOverviewPage.vue?raw'
import studentDashboardPanelRegistrySource from '@/features/student-dashboard/ui/studentDashboardPanelRegistry.ts?raw'

describe('student overview entrypoint', () => {
  it('学生仪表盘 registry 应通过稳定入口组件渲染学生概览，而不是直接绑定具体视觉实现', () => {
    expect(studentDashboardPanelRegistrySource).toContain(
      "import StudentOverviewPage from './StudentOverviewPage.vue'"
    )
    expect(studentDashboardPanelRegistrySource).not.toContain(
      "import StudentOverviewStyleEditorial from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue'"
    )
    expect(studentDashboardPanelRegistrySource).toContain('overview: StudentOverviewPage')
  })

  it('StudentOverviewPage 应退化为对当前实现的轻量包装，而不是继续保留旧版完整模板', () => {
    expect(studentOverviewPageSource).toContain(
      "import StudentOverviewStyleEditorial from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue'"
    )
    expect(studentOverviewPageSource).toContain('<StudentOverviewStyleEditorial')
    expect(studentOverviewPageSource).not.toContain('student-overview-legacy-hero')
  })
})
