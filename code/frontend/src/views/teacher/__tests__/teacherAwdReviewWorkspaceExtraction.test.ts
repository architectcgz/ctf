import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewWorkspaceStateSource from '@/widgets/awd-review-workspace/AwdReviewWorkspaceState.vue?raw'
import awdReviewWorkspaceActionsSource from '@/widgets/awd-review-workspace/TeacherAWDReviewWorkspaceActions.vue?raw'
import awdReviewStatusChipSource from '@/widgets/awd-review-workspace/AwdReviewStatusChip.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/TeacherAWDReviewSurfaceShell.vue?raw'
import awdReviewWorkspaceHeaderSource from '@/widgets/awd-review-workspace/TeacherAWDReviewWorkspaceHeader.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/awd-review-workspace/TeacherAWDReviewSummaryPanel.vue?raw'

describe('Teacher AWD review workspace extraction', () => {
  it('详情路由页应收敛为 widget 组合层', () => {
    expect(awdReviewDetailSource).toContain(
      "import { AwdReviewWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(awdReviewDetailSource).toContain('<AwdReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="teacher-management-shell')
    expect(awdReviewDetailSource).not.toContain('class="teacher-topbar')

    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewSurfaceShell')
    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewWorkspaceHeader')
    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewSummaryPanel')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewStatusChip')
    expect(awdReviewWorkspaceSource).toContain('<TeacherAWDReviewWorkspaceActions')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewWorkspaceState')
    expect(awdReviewWorkspaceSource).toContain('buildTeacherAwdReviewSummaryItems')
    expect(awdReviewWorkspaceActionsSource).toContain('data-testid="awd-review-export-archive"')
    expect(awdReviewWorkspaceActionsSource).toContain('data-testid="awd-review-export-report"')
    expect(awdReviewStatusChipSource).toContain('class="awd-review-status-chip"')
    expect(awdReviewStatusChipSource).toContain('awd-review-status-chip--running')
    expect(awdReviewWorkspaceStateSource).toContain('awd-review-loading')
    expect(awdReviewWorkspaceStateSource).toContain('title="复盘详情加载失败"')
    expect(awdReviewSurfaceShellSource).toContain('class="teacher-management-shell')
    expect(awdReviewWorkspaceHeaderSource).toContain('class="workspace-page-header teacher-topbar"')
    expect(awdReviewSummaryPanelSource).toContain('class="progress-card metric-panel-card"')
    expect(awdReviewSummaryPanelSource).toContain(
      '<component :is="item.icon" v-if="item.icon" class="h-4 w-4" />'
    )
  })
})
