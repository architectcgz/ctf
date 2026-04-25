import { describe, expect, it } from 'vitest'

import imageManageSource from '@/views/platform/ImageManage.vue?raw'

describe('ImageManage page state extraction', () => {
  it('应将镜像管理页面状态与行为抽到独立 composable', () => {
    expect(imageManageSource).toContain(
      "import { useImageManagePage } from '@/composables/useImageManagePage'"
    )
    expect(imageManageSource).toContain('} = useImageManagePage()')
  })
})
