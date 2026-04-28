import { describe, expect, it } from 'vitest'

import imageManageSource from '@/views/platform/ImageManage.vue?raw'
import imageManageHeroPanelSource from '@/components/platform/images/ImageManageHeroPanel.vue?raw'

describe('ImageManage workspace extraction', () => {
  it('应将镜像目录工作区抽到独立平台组件', () => {
    expect(imageManageSource).toContain(
      "import ImageDirectoryPanel from '@/components/platform/images/ImageDirectoryPanel.vue'"
    )
    expect(imageManageSource).toContain('<ImageDirectoryPanel')
  })

  it('应将镜像头部摘要抽到独立平台组件', () => {
    expect(imageManageSource).toContain(
      "import ImageManageHeroPanel from '@/components/platform/images/ImageManageHeroPanel.vue'"
    )
    expect(imageManageSource).toContain('<ImageManageHeroPanel')
    expect(imageManageHeroPanelSource).toContain('<div class="workspace-overline">')
    expect(imageManageHeroPanelSource).toContain('Image Registry')
    expect(imageManageHeroPanelSource).toContain('class="image-status-strip"')
  })
})
