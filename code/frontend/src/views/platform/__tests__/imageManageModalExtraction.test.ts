import { describe, expect, it } from 'vitest'

import imageManageSource from '@/views/platform/ImageManage.vue?raw'

describe('ImageManage modal extraction', () => {
  it('应将镜像详情和创建弹窗抽到独立平台组件', () => {
    expect(imageManageSource).toContain("import ImageDetailModal from '@/components/platform/images/ImageDetailModal.vue'")
    expect(imageManageSource).toContain("import ImageCreateModal from '@/components/platform/images/ImageCreateModal.vue'")
    expect(imageManageSource).toContain('<ImageDetailModal')
    expect(imageManageSource).toContain('<ImageCreateModal')
  })
})
