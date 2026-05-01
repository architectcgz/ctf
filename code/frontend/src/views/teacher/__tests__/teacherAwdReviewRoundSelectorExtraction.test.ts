import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import teacherAwdReviewRoundSelectorSource from '@/components/teacher/awd-review/TeacherAWDReviewRoundSelector.vue?raw'

describe('Teacher AWD review round selector extraction', () => {
  it('应将轮次切换区块下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import TeacherAWDReviewRoundSelector from '@/components/teacher/awd-review/TeacherAWDReviewRoundSelector.vue'"
    )
    expect(awdReviewDetailSource).toContain('<TeacherAWDReviewRoundSelector')
    expect(awdReviewDetailSource).not.toContain('class="awd-review-round-list custom-scrollbar"')
    expect(awdReviewDetailSource).not.toContain('默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。')

    expect(teacherAwdReviewRoundSelectorSource).toContain('class="awd-review-round-list custom-scrollbar"')
    expect(teacherAwdReviewRoundSelectorSource).toContain('默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。')
  })
})
