import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '@/views/platform/AWDReviewIndex.vue?raw'

describe('AWDReviewIndex hero extraction', () => {
  it('应将复盘页头部与摘要卡抽到独立平台组件', () => {
    expect(awdReviewIndexSource).toContain(
      "import AwdReviewHeroPanel from '@/components/platform/awd-review/AwdReviewHeroPanel.vue'"
    )
    expect(awdReviewIndexSource).toContain('<AwdReviewHeroPanel')
  })
})
