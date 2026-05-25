import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'
import awdReviewWidgetIndexSource from '@/widgets/awd-review-workspace/index.ts?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'
import awdReviewWorkspaceHeaderSource from '@/widgets/awd-review-workspace/AwdReviewWorkspaceHeader.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/awd-review-workspace/AwdReviewSummaryPanel.vue?raw'
import awdReviewContestDirectorySource from '@/widgets/awd-review-workspace/AwdReviewContestDirectory.vue?raw'
import awdReviewIndexFiltersSource from '@/widgets/awd-review-workspace/TeacherAWDReviewIndexFilters.vue?raw'
import awdReviewDirectorySectionSource from '@/widgets/awd-review-workspace/TeacherAWDReviewDirectorySection.vue?raw'
import awdReviewContestHeadSource from '@/widgets/awd-review-workspace/TeacherAWDReviewContestHead.vue?raw'
import awdReviewContestRowSource from '@/widgets/awd-review-workspace/TeacherAWDReviewContestRow.vue?raw'
import awdReviewContestRowCtaSource from '@/widgets/awd-review-workspace/TeacherAWDReviewContestRowCta.vue?raw'
import awdReviewContestRowMetricsSource from '@/widgets/awd-review-workspace/TeacherAWDReviewContestRowMetrics.vue?raw'
import awdReviewContestRowStatusTagsSource from '@/widgets/awd-review-workspace/TeacherAWDReviewContestRowStatusTags.vue?raw'
import awdReviewDirectoryStateSource from '@/widgets/awd-review-workspace/TeacherAWDReviewDirectoryState.vue?raw'

describe('Teacher AWD review index workspace extraction', () => {
  it('目录页路由应收敛为 widget 组合层', () => {
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewIndexWorkspace } from './AwdReviewIndexWorkspace.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewWorkspace } from './AwdReviewWorkspace.vue'"
    )
    expect(awdReviewWidgetIndexSource).not.toContain('AwdReviewWorkspaceState')
    expect(awdReviewWidgetIndexSource).not.toContain('AwdReviewContestDirectory.vue')

    expect(awdReviewIndexSource).toContain(
      "import { AwdReviewIndexWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(awdReviewIndexSource).toContain('<AwdReviewIndexWorkspace')
    expect(awdReviewIndexSource).not.toContain('TeacherAWDReviewIndexWorkspace')
    expect(awdReviewIndexSource).not.toContain('class="teacher-management-shell')
    expect(awdReviewIndexSource).not.toContain('class="teacher-topbar')

    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewSurfaceShell')
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewWorkspaceHeader')
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewSummaryPanel')
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewContestDirectory')
    expect(awdReviewIndexWorkspaceSource).toContain('buildTeacherAwdReviewIndexSummaryItems')
    expect(awdReviewIndexWorkspaceSource).toContain('TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(awdReviewSurfaceShellSource).toContain('class="teacher-management-shell')
    expect(awdReviewWorkspaceHeaderSource).toContain('class="workspace-page-header teacher-topbar"')
    expect(awdReviewSummaryPanelSource).toContain('class="progress-card metric-panel-card"')
    expect(awdReviewSummaryPanelSource).toContain(
      '<component :is="item.icon" v-if="item.icon" class="h-4 w-4" />'
    )
    expect(awdReviewContestDirectorySource).toContain('<TeacherAWDReviewDirectorySection')
    expect(awdReviewContestDirectorySource).toContain('<TeacherAWDReviewIndexFilters')
    expect(awdReviewContestDirectorySource).toContain('<TeacherAWDReviewDirectoryState')
    expect(awdReviewContestDirectorySource).toContain('<TeacherAWDReviewContestHead')
    expect(awdReviewContestDirectorySource).toContain('<TeacherAWDReviewContestRow')
    expect(awdReviewDirectorySectionSource).toContain('class="workspace-directory-section teacher-directory-section"')
    expect(awdReviewDirectorySectionSource).toContain('AWD_REVIEW_DIRECTORY_COLUMNS')
    expect(awdReviewIndexFiltersSource).toContain('class="teacher-directory-filters"')
    expect(awdReviewDirectoryStateSource).toContain('title="AWD复盘目录加载失败"')
    expect(awdReviewContestHeadSource).toContain('class="teacher-directory-head"')
    expect(awdReviewContestHeadSource).toContain("AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA")
    expect(awdReviewContestRowSource).toContain('<TeacherAWDReviewContestRowCta')
    expect(awdReviewContestRowSource).toContain('<TeacherAWDReviewContestRowMetrics')
    expect(awdReviewContestRowSource).toContain('<TeacherAWDReviewContestRowStatusTags')
    expect(awdReviewContestRowSource).toContain("AWD_REVIEW_DIRECTORY_COLUMN_SCHEMA")
    expect(awdReviewContestRowSource).toContain('class="teacher-directory-row"')
    expect(awdReviewContestRowCtaSource).toContain('class="teacher-directory-row-cta"')
    expect(awdReviewContestRowMetricsSource).toContain('class="teacher-directory-row-metrics"')
    expect(awdReviewContestRowStatusTagsSource).toContain('class="teacher-directory-row-tags"')
  })
})
