import { describe, expect, it } from 'vitest'

import imageManageSource from '@/views/platform/ImageManage.vue?raw'

describe('ImageManage workspace extraction', () => {
  it('应将镜像目录工作区抽到独立平台组件', () => {
    expect(imageManageSource).toContain(
      "import ImageDirectoryPanel from '@/components/platform/images/ImageDirectoryPanel.vue'"
    )
    expect(imageManageSource).toContain('<ImageDirectoryPanel')
  })
})
