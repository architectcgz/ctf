import { describe, expect, it } from 'vitest'

import platformAwdReviewIndexSource from '@/pages/awd-review/PlatformAwdReviewIndexRoutePage.vue?raw'
import teacherAwdReviewDetailSource from '@/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue?raw'
import teacherAwdReviewIndexSource from '@/pages/awd-review/TeacherAwdReviewIndexRoutePage.vue?raw'
import awdReviewWidgetIndexSource from '@/widgets/awd-review-workspace/index.ts?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'
import awdReviewWorkspaceHeaderSource from '@/widgets/awd-review-workspace/AwdReviewWorkspaceHeader.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/awd-review-workspace/AwdReviewSummaryPanel.vue?raw'
import awdReviewHeroPanelSource from '@/widgets/awd-review-workspace/AwdReviewHeroPanel.vue?raw'
import awdReviewDirectoryPanelSource from '@/widgets/awd-review-workspace/AwdReviewDirectoryPanel.vue?raw'
import awdReviewContestDirectorySource from '@/widgets/awd-review-workspace/AwdReviewContestDirectory.vue?raw'
import awdReviewIndexFiltersSource from '@/widgets/awd-review-workspace/AwdReviewIndexFilters.vue?raw'
import awdReviewDirectorySectionSource from '@/widgets/awd-review-workspace/AwdReviewDirectorySection.vue?raw'
import awdReviewDirectoryStateSource from '@/widgets/awd-review-workspace/AwdReviewDirectoryState.vue?raw'
import awdReviewContestHeadSource from '@/widgets/awd-review-workspace/AwdReviewContestHead.vue?raw'
import awdReviewContestRowSource from '@/widgets/awd-review-workspace/AwdReviewContestRow.vue?raw'
import awdReviewContestRowCtaSource from '@/widgets/awd-review-workspace/AwdReviewContestRowCta.vue?raw'
import awdReviewContestRowMetricsSource from '@/widgets/awd-review-workspace/AwdReviewContestRowMetrics.vue?raw'
import awdReviewContestRowStatusTagsSource from '@/widgets/awd-review-workspace/AwdReviewContestRowStatusTags.vue?raw'
import awdReviewAnalysisSectionSource from '@/widgets/awd-review-workspace/AwdReviewAnalysisSection.vue?raw'
import awdReviewEvidenceGridSource from '@/widgets/awd-review-workspace/AwdReviewEvidenceGrid.vue?raw'
import awdReviewRoundSelectorSource from '@/widgets/awd-review-workspace/AwdReviewRoundSelector.vue?raw'

describe('awd review workspace ui strategy', () => {
  it('route pages should stay as widget composition shells instead of rebuilding review workspace internals', () => {
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewIndexWorkspace } from './AwdReviewIndexWorkspace.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewWorkspace } from './AwdReviewWorkspace.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewHeroPanel } from './AwdReviewHeroPanel.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewDirectoryPanel } from './AwdReviewDirectoryPanel.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewAnalysisSection } from './AwdReviewAnalysisSection.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewEvidenceGrid } from './AwdReviewEvidenceGrid.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewRoundSelector } from './AwdReviewRoundSelector.vue'"
    )
    expect(awdReviewWidgetIndexSource).toContain(
      "export { default as AwdReviewTeamDrawer } from './AwdReviewTeamDrawer.vue'"
    )
    expect(teacherAwdReviewIndexSource).toContain(
      "import { AwdReviewIndexWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(teacherAwdReviewIndexSource).toContain('<AwdReviewIndexWorkspace')
    expect(teacherAwdReviewDetailSource).toContain(
      "import { AwdReviewWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(teacherAwdReviewDetailSource).toContain('<AwdReviewWorkspace')
    expect(platformAwdReviewIndexSource).toContain('AwdReviewHeroPanel')
    expect(platformAwdReviewIndexSource).toContain('AwdReviewDirectoryPanel')
  })

  it('index workspace should keep hero, summary, directory, and route target owners inside the widget slice', () => {
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewSurfaceShell')
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewWorkspaceHeader')
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewSummaryPanel')
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewContestDirectory')
    expect(awdReviewIndexWorkspaceSource).toContain('buildAwdReviewIndexSummaryItems')
    expect(awdReviewIndexWorkspaceSource).toContain('AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(awdReviewIndexWorkspaceSource).not.toContain('buildTeacherAwdReviewIndexSummaryItems')
    expect(awdReviewIndexWorkspaceSource).not.toContain('TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(awdReviewHeroPanelSource).toContain('<header class="workspace-page-header admin-awd-review-shell__hero">')
    expect(awdReviewHeroPanelSource).toContain('返回平台概览')
    expect(awdReviewDirectoryPanelSource).toContain("from '@/shared/ui/navigation/AppRouteLink.vue'")
    expect(awdReviewDirectoryPanelSource).toContain('<WorkspaceDirectoryToolbar')
    expect(awdReviewDirectoryPanelSource).toContain('<WorkspaceDataTable')
    expect(awdReviewContestDirectorySource).toContain('<AwdReviewDirectorySection')
    expect(awdReviewContestDirectorySource).toContain('<AwdReviewIndexFilters')
    expect(awdReviewContestDirectorySource).toContain('<AwdReviewDirectoryState')
    expect(awdReviewContestDirectorySource).toContain('<AwdReviewContestHead')
    expect(awdReviewContestDirectorySource).toContain('<AwdReviewContestRow')
    expect(awdReviewDirectorySectionSource).toContain(
      'class="workspace-directory-section teacher-directory-section"'
    )
    expect(awdReviewIndexFiltersSource).toContain('class="teacher-directory-filters"')
    expect(awdReviewDirectoryStateSource).toContain('title="AWD复盘目录加载失败"')
    expect(awdReviewContestHeadSource).toContain('class="teacher-directory-head"')
    expect(awdReviewContestRowSource).toContain('<AwdReviewContestRowCta')
    expect(awdReviewContestRowSource).toContain('<AwdReviewContestRowMetrics')
    expect(awdReviewContestRowSource).toContain('<AwdReviewContestRowStatusTags')
    expect(awdReviewContestRowCtaSource).toContain('class="teacher-directory-row-cta"')
    expect(awdReviewContestRowMetricsSource).toContain('class="teacher-directory-row-metrics"')
    expect(awdReviewContestRowStatusTagsSource).toContain('class="teacher-directory-row-tags"')
  })

  it('detail workspace should keep round selector, analysis, and evidence sections as extracted sub-owners', () => {
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewSurfaceShell')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewWorkspaceHeader')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewSummaryPanel')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewRoundSelector')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewAnalysisSection')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewEvidenceGrid')
    expect(awdReviewWorkspaceSource).toContain('buildAwdReviewSummaryItems')
    expect(awdReviewWorkspaceSource).toContain('AWD_REVIEW_WORKSPACE_COPY')
    expect(awdReviewWorkspaceSource).not.toContain('buildTeacherAwdReviewSummaryItems')
    expect(awdReviewWorkspaceSource).not.toContain('TEACHER_AWD_REVIEW_WORKSPACE_COPY')
    expect(teacherAwdReviewDetailSource).not.toContain('class="awd-review-round-list custom-scrollbar"')
    expect(teacherAwdReviewDetailSource).not.toContain('class="awd-review-round-grid"')
    expect(teacherAwdReviewDetailSource).not.toContain('data-testid="awd-review-service-id"')
    expect(teacherAwdReviewDetailSource).not.toContain('data-testid="awd-review-attack-service-id"')
    expect(teacherAwdReviewDetailSource).not.toContain('data-testid="awd-review-traffic-service-id"')
    expect(awdReviewRoundSelectorSource).toContain(
      'class="awd-review-round-shell workspace-directory-list"'
    )
    expect(awdReviewRoundSelectorSource).toContain('class="awd-review-round-list custom-scrollbar"')
    expect(awdReviewAnalysisSectionSource).toContain('class="awd-review-round-grid"')
    expect(awdReviewAnalysisSectionSource).toContain('class="teacher-directory"')
    expect(awdReviewEvidenceGridSource).toContain('data-testid="awd-review-service-id"')
    expect(awdReviewEvidenceGridSource).toContain('data-testid="awd-review-attack-service-id"')
    expect(awdReviewEvidenceGridSource).toContain('data-testid="awd-review-traffic-service-id"')
    expect(awdReviewSurfaceShellSource).toContain('class="teacher-management-shell')
    expect(awdReviewWorkspaceHeaderSource).toContain('class="workspace-page-header teacher-topbar"')
    expect(awdReviewSummaryPanelSource).toContain('class="progress-card metric-panel-card"')
  })
})
