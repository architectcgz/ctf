import { describe, expect, it } from 'vitest'

import adminInstanceManageSource from '@/pages/platform/InstanceManageRoutePage.vue?raw'
import instanceManageHeroPanelSource from '@/features/platform/instance-management/ui/InstanceManageHeroPanel.vue?raw'

describe('Platform InstanceManage panel extraction', () => {
  it('应将实例管理头部与摘要卡抽到独立 platform instance 组件', () => {
    expect(adminInstanceManageSource).toContain("from '@/features/platform/instance-management'")
    expect(adminInstanceManageSource).toContain('InstanceManageHeroPanel')
    expect(adminInstanceManageSource).toContain('<InstanceManageHeroPanel')
    expect(instanceManageHeroPanelSource).toContain('<div class="workspace-overline">')
    expect(instanceManageHeroPanelSource).toContain('Instance Workspace')
    expect(instanceManageHeroPanelSource).toContain('返回概览')
    expect(instanceManageHeroPanelSource).toContain('刷新列表')
    expect(instanceManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-instance-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
  })
})
