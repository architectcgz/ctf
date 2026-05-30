import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '@/pages/awd-review/PlatformAwdReviewIndexRoutePage.vue?raw'

describe('AWDReviewIndex hero extraction', () => {
  it('应将复盘页头部与摘要卡抽到独立平台组件', () => {
    expect(awdReviewIndexSource).toContain(
      "AwdReviewHeroPanel"
    )
    expect(awdReviewIndexSource).toContain('<AwdReviewHeroPanel')
  })
})
