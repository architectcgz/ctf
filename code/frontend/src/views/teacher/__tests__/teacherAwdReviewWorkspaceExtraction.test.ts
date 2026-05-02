import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue?raw'

describe('Teacher AWD review workspace extraction', () => {
  it('详情路由页应收敛为 widget 组合层', () => {
    expect(awdReviewDetailSource).toContain(
      "import { TeacherAWDReviewWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(awdReviewDetailSource).toContain('<TeacherAWDReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="teacher-management-shell')
    expect(awdReviewDetailSource).not.toContain('class="teacher-topbar')

    expect(awdReviewWorkspaceSource).toContain('class="teacher-management-shell')
    expect(awdReviewWorkspaceSource).toContain('class="teacher-topbar workspace-tab-heading')
  })
})
