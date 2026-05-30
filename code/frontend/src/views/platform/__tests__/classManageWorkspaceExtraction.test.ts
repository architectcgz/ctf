import { describe, expect, it } from 'vitest'

import adminClassManageSource from '@/pages/platform/ClassManageRoutePage.vue?raw'

describe('Platform ClassManage workspace extraction', () => {
  it('应将班级目录工作区抽到独立 platform class 组件', () => {
    expect(adminClassManageSource).toContain("from '@/features/platform/class-management'")
    expect(adminClassManageSource).toContain('ClassManageWorkspacePanel')
    expect(adminClassManageSource).toContain('<ClassManageWorkspacePanel')
  })
})
