import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '@/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewWorkspaceStateSource from '@/widgets/awd-review-workspace/AwdReviewWorkspaceState.vue?raw'
import awdReviewWorkspaceActionsSource from '@/widgets/awd-review-workspace/AwdReviewWorkspaceActions.vue?raw'
import awdReviewStatusChipSource from '@/widgets/awd-review-workspace/AwdReviewStatusChip.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'
import awdReviewWorkspaceHeaderSource from '@/widgets/awd-review-workspace/AwdReviewWorkspaceHeader.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/awd-review-workspace/AwdReviewSummaryPanel.vue?raw'
import awdReviewPresentationSource from '@/widgets/awd-review-workspace/model/presentation.ts?raw'

describe('Teacher AWD review workspace extraction', () => {
  it('详情路由页应收敛为 widget 组合层', () => {
    expect(awdReviewDetailSource).toContain(
      "import { AwdReviewWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(awdReviewDetailSource).toContain('<AwdReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="teacher-management-shell')
    expect(awdReviewDetailSource).not.toContain('class="teacher-topbar')

    expect(awdReviewWorkspaceSource).toContain('<AwdReviewSurfaceShell')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewWorkspaceHeader')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewSummaryPanel')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewStatusChip')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewWorkspaceActions')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewWorkspaceState')
    expect(awdReviewWorkspaceSource).toContain('buildAwdReviewSummaryItems')
    expect(awdReviewWorkspaceSource).not.toContain('buildTeacherAwdReviewSummaryItems')
    expect(awdReviewWorkspaceSource).toContain('AWD_REVIEW_WORKSPACE_COPY')
    expect(awdReviewWorkspaceSource).not.toContain('TEACHER_AWD_REVIEW_WORKSPACE_COPY')
    expect(awdReviewWorkspaceSource).toContain('AwdReviewArchiveData')
    expect(awdReviewWorkspaceSource).toContain('AwdReviewTeamItemData')
    expect(awdReviewWorkspaceSource).not.toContain('TeacherAWDReviewArchiveData')
    expect(awdReviewWorkspaceSource).not.toContain('TeacherAWDReviewTeamItemData')
    expect(awdReviewWorkspaceSource).toContain("from '@/components/awd-review'")
    expect(awdReviewWorkspaceSource).not.toContain(
      "from '@/components/teacher/awd-review/AwdReviewRoundSelector.vue'"
    )
    expect(awdReviewPresentationSource).toContain('export interface AwdReviewSummaryStats')
    expect(awdReviewPresentationSource).toContain('export interface AwdReviewSummaryItem')
    expect(awdReviewPresentationSource).toContain('export interface AwdReviewIndexSummaryStats')
    expect(awdReviewPresentationSource).toContain('export const AWD_REVIEW_WORKSPACE_COPY')
    expect(awdReviewPresentationSource).toContain('export const AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(awdReviewPresentationSource).toContain('export function buildAwdReviewSummaryItems')
    expect(awdReviewPresentationSource).toContain('export function buildAwdReviewIndexSummaryItems')
    expect(awdReviewPresentationSource).not.toContain('TeacherAwdReviewSummaryStats')
    expect(awdReviewPresentationSource).not.toContain('TeacherAwdReviewSummaryItem')
    expect(awdReviewPresentationSource).not.toContain('TeacherAwdReviewIndexSummaryStats')
    expect(awdReviewPresentationSource).not.toContain('TEACHER_AWD_REVIEW_WORKSPACE_COPY')
    expect(awdReviewPresentationSource).not.toContain('TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(awdReviewPresentationSource).not.toContain('buildTeacherAwdReviewSummaryItems')
    expect(awdReviewPresentationSource).not.toContain('buildTeacherAwdReviewIndexSummaryItems')
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
