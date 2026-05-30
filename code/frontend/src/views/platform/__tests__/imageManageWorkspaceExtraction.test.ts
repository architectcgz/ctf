import { describe, expect, it } from 'vitest'

import imageManageSource from '@/pages/platform/ImageManageRoutePage.vue?raw'
import imageManageHeroPanelSource from '@/features/image-management/ui/ImageManageHeroPanel.vue?raw'

describe('ImageManage workspace extraction', () => {
  it('应将镜像目录工作区抽到独立平台组件', () => {
    expect(imageManageSource).toContain("from '@/features/image-management'")
    expect(imageManageSource).toContain('ImageDirectoryPanel')
    expect(imageManageSource).toContain('<ImageDirectoryPanel')
  })

  it('应将镜像头部摘要抽到独立平台组件', () => {
    expect(imageManageSource).toContain("from '@/features/image-management'")
    expect(imageManageSource).toContain('ImageManageHeroPanel')
    expect(imageManageSource).toContain('<ImageManageHeroPanel')
    expect(imageManageHeroPanelSource).toContain('<div class="workspace-overline">')
    expect(imageManageHeroPanelSource).toContain('Image Registry')
    expect(imageManageHeroPanelSource).toContain('class="image-status-strip"')
  })
})
