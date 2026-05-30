import { describe, expect, it } from 'vitest'

import imageManageSource from '@/pages/platform/ImageManageRoutePage.vue?raw'

describe('ImageManage page state extraction', () => {
  it('应将镜像管理页面状态与行为抽到独立 composable', () => {
    expect(imageManageSource).toContain("from '@/features/image-management'")
    expect(imageManageSource).toContain('useImageManagePage')
    expect(imageManageSource).toContain('} = useImageManagePage()')
  })
})
