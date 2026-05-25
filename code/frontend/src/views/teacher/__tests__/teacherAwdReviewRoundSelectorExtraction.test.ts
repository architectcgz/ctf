import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import teacherAwdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewRoundSelectorSource from '@/components/teacher/awd-review/AwdReviewRoundSelector.vue?raw'

describe('Teacher AWD review round selector extraction', () => {
  it('应将轮次切换区块下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import { AwdReviewWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(awdReviewDetailSource).toContain('<AwdReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="awd-review-round-list custom-scrollbar"')
    expect(awdReviewDetailSource).not.toContain('默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。')

    expect(teacherAwdReviewWorkspaceSource).toContain('<AwdReviewRoundSelector')
    expect(awdReviewRoundSelectorSource).toContain(
      'class="awd-review-round-shell workspace-directory-list"'
    )
    expect(awdReviewRoundSelectorSource).toContain('class="awd-review-round-list custom-scrollbar"')
    expect(awdReviewRoundSelectorSource).toContain(
      'class="workspace-directory-chip awd-review-round-chip"'
    )
    expect(awdReviewRoundSelectorSource).toContain('默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。')
  })
})
