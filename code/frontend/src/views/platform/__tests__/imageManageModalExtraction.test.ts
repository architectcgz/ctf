import { describe, expect, it } from 'vitest'

import imageManageSource from '@/pages/platform/ImageManageRoutePage.vue?raw'

describe('ImageManage modal extraction', () => {
  it('应将镜像详情和创建弹窗抽到独立平台组件', () => {
    expect(imageManageSource).toContain("from '@/features/image-management'")
    expect(imageManageSource).toContain('ImageDetailModal')
    expect(imageManageSource).toContain('ImageCreateModal')
    expect(imageManageSource).toContain('<ImageDetailModal')
    expect(imageManageSource).toContain('<ImageCreateModal')
  })
})
