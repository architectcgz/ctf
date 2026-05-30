import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '@/pages/awd-review/PlatformAwdReviewIndexRoutePage.vue?raw'

describe('AWDReviewIndex directory extraction', () => {
  it('应将复盘赛事目录工作区抽到独立平台组件', () => {
    expect(awdReviewIndexSource).toContain(
      "AwdReviewDirectoryPanel"
    )
    expect(awdReviewIndexSource).toContain('<AwdReviewDirectoryPanel')
  })
})
