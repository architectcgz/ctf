import { describe, expect, it } from 'vitest'

import awdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewIndexWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSurfaceShell.vue?raw'
import awdReviewWorkspaceHeaderSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspaceHeader.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSummaryPanel.vue?raw'
import awdReviewContestDirectorySource from '@/widgets/teacher-awd-review/TeacherAWDReviewContestDirectory.vue?raw'
import awdReviewIndexFiltersSource from '@/widgets/teacher-awd-review/TeacherAWDReviewIndexFilters.vue?raw'
import awdReviewDirectorySectionSource from '@/widgets/teacher-awd-review/TeacherAWDReviewDirectorySection.vue?raw'
import awdReviewContestHeadSource from '@/widgets/teacher-awd-review/TeacherAWDReviewContestHead.vue?raw'
import awdReviewContestRowSource from '@/widgets/teacher-awd-review/TeacherAWDReviewContestRow.vue?raw'
import awdReviewContestRowCtaSource from '@/widgets/teacher-awd-review/TeacherAWDReviewContestRowCta.vue?raw'
import awdReviewContestRowMetricsSource from '@/widgets/teacher-awd-review/TeacherAWDReviewContestRowMetrics.vue?raw'
import awdReviewContestRowStatusTagsSource from '@/widgets/teacher-awd-review/TeacherAWDReviewContestRowStatusTags.vue?raw'
import awdReviewDirectoryStateSource from '@/widgets/teacher-awd-review/TeacherAWDReviewDirectoryState.vue?raw'

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
    expect(awdReviewIndexWorkspaceSource).toContain('<TeacherAWDReviewContestDirectory')
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
