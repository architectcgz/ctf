import { describe, expect, it } from 'vitest'

import adminStudentManageSource from '../StudentManage.vue?raw'
import studentManageHeroPanelSource from '@/components/platform/student/StudentManageHeroPanel.vue?raw'

describe('Platform StudentManage panel extraction', () => {
  it('应将学生管理头部与摘要卡抽到独立 platform student 组件', () => {
    expect(adminStudentManageSource).toContain(
      "import StudentManageHeroPanel from '@/components/platform/student/StudentManageHeroPanel.vue'"
    )
    expect(adminStudentManageSource).toContain('<StudentManageHeroPanel')
    expect(studentManageHeroPanelSource).toContain('Student Workspace')
    expect(studentManageHeroPanelSource).toContain('刷新目录')
    expect(studentManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-student-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
  })
})
