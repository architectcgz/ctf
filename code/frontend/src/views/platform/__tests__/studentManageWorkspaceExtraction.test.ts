import { describe, expect, it } from 'vitest'

import adminStudentManageSource from '@/pages/platform/StudentManageRoutePage.vue?raw'

describe('Platform StudentManage workspace extraction', () => {
  it('应将学生目录工作区抽到独立 platform student 组件', () => {
    expect(adminStudentManageSource).toContain("from '@/features/platform/student-management'")
    expect(adminStudentManageSource).toContain('StudentManageWorkspacePanel')
    expect(adminStudentManageSource).toContain('<StudentManageWorkspacePanel')
  })
})
