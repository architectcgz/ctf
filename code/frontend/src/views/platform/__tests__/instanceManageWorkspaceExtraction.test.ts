import { describe, expect, it } from 'vitest'

import adminInstanceManageSource from '../InstanceManage.vue?raw'

describe('Platform InstanceManage workspace extraction', () => {
  it('应将实例目录工作区抽到独立 platform instance 组件', () => {
    expect(adminInstanceManageSource).toContain(
      "import InstanceManageWorkspacePanel from '@/components/platform/instance/InstanceManageWorkspacePanel.vue'"
    )
    expect(adminInstanceManageSource).toContain('<InstanceManageWorkspacePanel')
  })
})
