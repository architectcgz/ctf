import { describe, expect, it } from 'vitest'

import adminInstanceManageSource from '@/pages/platform/InstanceManageRoutePage.vue?raw'

describe('Platform InstanceManage workspace extraction', () => {
  it('应将实例目录工作区抽到独立 platform instance 组件', () => {
    expect(adminInstanceManageSource).toContain("from '@/features/platform/instance-management'")
    expect(adminInstanceManageSource).toContain('InstanceManageWorkspacePanel')
    expect(adminInstanceManageSource).toContain('<InstanceManageWorkspacePanel')
  })
})
