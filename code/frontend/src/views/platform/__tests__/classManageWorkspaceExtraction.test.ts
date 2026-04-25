import { describe, expect, it } from 'vitest'

import adminClassManageSource from '../ClassManage.vue?raw'

describe('Platform ClassManage workspace extraction', () => {
  it('应将班级目录工作区抽到独立 platform class 组件', () => {
    expect(adminClassManageSource).toContain(
      "import ClassManageWorkspacePanel from '@/components/platform/class/ClassManageWorkspacePanel.vue'"
    )
    expect(adminClassManageSource).toContain('<ClassManageWorkspacePanel')
  })
})
