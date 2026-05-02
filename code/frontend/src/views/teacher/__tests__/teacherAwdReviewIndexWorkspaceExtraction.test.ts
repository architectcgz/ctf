import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewIndexWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSurfaceShell.vue?raw'
import awdReviewWorkspaceHeaderSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspaceHeader.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSummaryPanel.vue?raw'

describe('Teacher AWD review index workspace extraction', () => {
  it('目录页路由应收敛为 widget 组合层', () => {
    expect(awdReviewIndexSource).toContain(
      "import { TeacherAWDReviewIndexWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(awdReviewIndexSource).toContain('<TeacherAWDReviewIndexWorkspace')
    expect(awdReviewIndexSource).not.toContain('class="teacher-management-shell')
    expect(awdReviewIndexSource).not.toContain('class="teacher-topbar')

    expect(awdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewSurfaceShell')
    expect(awdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewWorkspaceHeader')
    expect(awdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewSummaryPanel')
    expect(awdReviewSurfaceShellSource).toContain('class="teacher-management-shell')
    expect(awdReviewWorkspaceHeaderSource).toContain('class="teacher-topbar workspace-tab-heading"')
    expect(awdReviewSummaryPanelSource).toContain('class="progress-card metric-panel-card"')
  })
})
