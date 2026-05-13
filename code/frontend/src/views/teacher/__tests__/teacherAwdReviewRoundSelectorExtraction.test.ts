import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import teacherAwdReviewWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue?raw'
import teacherAwdReviewRoundSelectorSource from '@/components/teacher/awd-review/TeacherAWDReviewRoundSelector.vue?raw'

describe('Teacher AWD review round selector extraction', () => {
  it('应将轮次切换区块下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import { TeacherAWDReviewWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(awdReviewDetailSource).toContain('<TeacherAWDReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="awd-review-round-list custom-scrollbar"')
    expect(awdReviewDetailSource).not.toContain('默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。')

    expect(teacherAwdReviewWorkspaceSource).toContain('<TeacherAWDReviewRoundSelector')
    expect(teacherAwdReviewRoundSelectorSource).toContain(
      'class="awd-review-round-shell workspace-directory-list"'
    )
    expect(teacherAwdReviewRoundSelectorSource).toContain('class="awd-review-round-list custom-scrollbar"')
    expect(teacherAwdReviewRoundSelectorSource).toContain(
      'class="workspace-directory-chip awd-review-round-chip"'
    )
    expect(teacherAwdReviewRoundSelectorSource).toContain('默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。')
  })
})
