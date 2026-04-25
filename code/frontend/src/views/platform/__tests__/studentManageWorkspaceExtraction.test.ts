import { describe, expect, it } from 'vitest'

import adminStudentManageSource from '../StudentManage.vue?raw'

describe('Platform StudentManage workspace extraction', () => {
  it('应将学生目录工作区抽到独立 platform student 组件', () => {
    expect(adminStudentManageSource).toContain(
      "import StudentManageWorkspacePanel from '@/components/platform/student/StudentManageWorkspacePanel.vue'"
    )
    expect(adminStudentManageSource).toContain('<StudentManageWorkspacePanel')
  })
})
