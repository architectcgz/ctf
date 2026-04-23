import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '@/views/platform/AWDReviewIndex.vue?raw'

describe('AWDReviewIndex directory extraction', () => {
  it('应将复盘赛事目录工作区抽到独立平台组件', () => {
    expect(awdReviewIndexSource).toContain(
      "import AwdReviewDirectoryPanel from '@/components/platform/awd-review/AwdReviewDirectoryPanel.vue'"
    )
    expect(awdReviewIndexSource).toContain('<AwdReviewDirectoryPanel')
  })
})
