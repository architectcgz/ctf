import { describe, expect, it } from 'vitest'

import adminClassManageSource from '../ClassManage.vue?raw'
import classManageHeroPanelSource from '@/components/platform/class/ClassManageHeroPanel.vue?raw'

describe('Platform ClassManage panel extraction', () => {
  it('应将班级管理头部与摘要卡抽到独立 platform class 组件', () => {
    expect(adminClassManageSource).toContain(
      "import ClassManageHeroPanel from '@/components/platform/class/ClassManageHeroPanel.vue'"
    )
    expect(adminClassManageSource).toContain('<ClassManageHeroPanel')
    expect(classManageHeroPanelSource).toContain('Class Workspace')
    expect(classManageHeroPanelSource).toContain('刷新目录')
    expect(classManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-class-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
  })
})
