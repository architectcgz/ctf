import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue?raw'
import awdReviewWorkspaceStateSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspaceState.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSurfaceShell.vue?raw'
import awdReviewWorkspaceHeaderSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspaceHeader.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSummaryPanel.vue?raw'

describe('Teacher AWD review workspace extraction', () => {
  it('详情路由页应收敛为 widget 组合层', () => {
    expect(awdReviewDetailSource).toContain(
      "import { TeacherAWDReviewWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(awdReviewDetailSource).toContain('<TeacherAWDReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="teacher-management-shell')
    expect(awdReviewDetailSource).not.toContain('class="teacher-topbar')

    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewSurfaceShell')
    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewWorkspaceHeader')
    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewSummaryPanel')
    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewWorkspaceState')
    expect(awdReviewWorkspaceStateSource).toContain('awd-review-loading')
    expect(awdReviewWorkspaceStateSource).toContain('title="复盘详情加载失败"')
    expect(awdReviewSurfaceShellSource).toContain('class="teacher-management-shell')
    expect(awdReviewWorkspaceHeaderSource).toContain('class="teacher-topbar workspace-tab-heading"')
    expect(awdReviewSummaryPanelSource).toContain('class="progress-card metric-panel-card"')
  })
})
