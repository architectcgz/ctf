import { describe, expect, it } from 'vitest'

import studentOverviewPageSource from '@/components/dashboard/student/StudentOverviewPage.vue?raw'
import dashboardViewSource from '@/views/dashboard/DashboardView.vue?raw'

describe('student overview entrypoint', () => {
  it('dashboard 应通过稳定入口组件渲染学生概览，而不是直接绑定具体视觉实现', () => {
    expect(dashboardViewSource).toContain(
      "import StudentOverviewPage from '@/components/dashboard/student/StudentOverviewPage.vue'"
    )
    expect(dashboardViewSource).not.toContain(
      "import StudentOverviewStyleEditorial from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue'"
    )
    expect(dashboardViewSource).toContain('return StudentOverviewPage')
  })

  it('StudentOverviewPage 应退化为对当前实现的轻量包装，而不是继续保留旧版完整模板', () => {
    expect(studentOverviewPageSource).toContain(
      "import StudentOverviewStyleEditorial from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue'"
    )
    expect(studentOverviewPageSource).toContain('<StudentOverviewStyleEditorial')
    expect(studentOverviewPageSource).not.toContain('student-overview-legacy-hero')
  })
})
