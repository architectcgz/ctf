import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewIndexWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSurfaceShell.vue?raw'

describe('Teacher AWD review index workspace extraction', () => {
  it('目录页路由应收敛为 widget 组合层', () => {
    expect(awdReviewIndexSource).toContain(
      "import { TeacherAWDReviewIndexWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(awdReviewIndexSource).toContain('<TeacherAWDReviewIndexWorkspace')
    expect(awdReviewIndexSource).not.toContain('class="teacher-management-shell')
    expect(awdReviewIndexSource).not.toContain('class="teacher-topbar')

    expect(awdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewSurfaceShell')
    expect(awdReviewIndexWorkspaceSource).toContain(
      'class="teacher-topbar workspace-tab-heading awd-review-index-header"'
    )
    expect(awdReviewSurfaceShellSource).toContain('class="teacher-management-shell')
  })
})
