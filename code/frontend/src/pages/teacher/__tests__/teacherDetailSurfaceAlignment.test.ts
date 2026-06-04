import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classStudentsSourceBase from '@/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue?raw'
import studentAnalysisSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue?raw'
import studentAnalysisOverviewHeroSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue?raw'
import classTrendPanelSource from '@/entities/class-insight/ui/ClassTrendPanel.vue?raw'
import classInsightsPanelSource from '@/entities/class-insight/ui/ClassInsightsPanel.vue?raw'
import classReviewPanelSource from '@/entities/class-insight/ui/ClassReviewPanel.vue?raw'
import interventionPanelSource from '@/features/teaching/student-analysis-review/ui/InterventionPanel.vue?raw'
import studentInsightPanelSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue?raw'
import studentInsightPrimarySectionsSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue?raw'
import studentInsightOverviewSectionSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue?raw'
import studentInsightRecommendationsSectionSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue?raw'
import studentInsightWriteupsSource from '@/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue?raw'
import studentInsightManualReviewSource from '@/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue?raw'
import reviewArchiveSource from '@/pages/review-archive/StudentReviewArchiveRoutePage.vue?raw'
import reviewArchiveWorkspaceSource from '@/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue?raw'
import reviewArchiveSummarySectionSource from '@/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue?raw'
import sectionCardSource from '@/shared/ui/common/SectionCard.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)
const styleSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')
const teacherPanelShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-panel-shell.css`,
  'utf-8'
)
const classStudentsSource = [
  classStudentsSourceBase,
  classStudentsOverviewPanelSource,
  classStudentsInsightWindowPanelSource,
  classStudentsDirectoryPanelSource,
].join('\n')
const reviewArchiveCombinedSource = [
  reviewArchiveSource,
  reviewArchiveWorkspaceSource,
  reviewArchiveSummarySectionSource,
].join('\n')
const studentInsightCompositeSource = [
  studentInsightPanelSource,
  studentInsightPrimarySectionsSource,
  studentInsightOverviewSectionSource,
  studentInsightRecommendationsSectionSource,
  studentInsightWriteupsSource,
  studentInsightManualReviewSource,
].join('\n')
const studentAnalysisCompositeSource = [
  studentAnalysisSource,
  studentAnalysisOverviewHeroSource,
].join('\n')

describe('teacher detail surface alignment', () => {
  it('class students and student analysis pages should soften control and section borders', () => {
    expect(styleSource).toMatch(
      /\.header-btn\s*\{[\s\S]*border:\s*1px solid[\s\S]*--header-control-border/s
    )
    expect(teacherSurfaceSource).toContain('--header-control-border: var(--teacher-control-border);')

    expect(classStudentsSource).toContain('--teacher-card-border:')
    expect(classStudentsSource).toContain('--teacher-control-border:')
    expect(classStudentsSource).toContain('--teacher-divider:')
    expect(classStudentsSource).not.toMatch(/\.teacher-(?:btn)\s*\{/)
    expect(classStudentsSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(classStudentsSource).toContain('class="teacher-directory-shell workspace-directory-list"')
    expect(classStudentsSource).toContain('class="teacher-student-directory-table"')
    expect(classStudentsSource).not.toContain('teacher-table-shell')
    expect(classStudentsSource).toMatch(
      /\.teacher-directory-shell\s*\{[\s\S]*--workspace-directory-shell-border:\s*color-mix\(/s
    )
    expect(classStudentsSource).toMatch(
      /\.teacher-student-directory-table\s*\{[\s\S]*--workspace-directory-shell-border:\s*color-mix\(\s*in srgb,\s*var\(--teacher-card-border\)\s*86%,\s*transparent\s*\);/s
    )

    expect(studentAnalysisSource).toContain('--teacher-card-border:')
    expect(studentAnalysisSource).toContain('--teacher-divider:')
    expect(studentAnalysisSource).toContain(
      'class="workspace-shell student-analysis-shell journal-eyebrow-text flex min-h-full flex-1 flex-col"'
    )
    expect(studentAnalysisSource).not.toContain('class="workspace-topbar"')
    expect(studentAnalysisSource).toContain('StudentAnalysisWorkspaceTabs')
    expect(studentAnalysisOverviewHeroSource).toContain(
      'class="workspace-panel-header student-analysis-overview-head"'
    )
    expect(studentAnalysisOverviewHeroSource).toContain(
      'class="workspace-panel-header__actions header-actions"'
    )
    expect(studentAnalysisOverviewHeroSource).toContain(
      'class="workspace-panel-header__summary summary-strip metric-panel-grid"'
    )
    expect(studentAnalysisOverviewHeroSource).not.toContain('class="workspace-panel-divider"')
    expect(studentAnalysisOverviewHeroSource).toContain(
      'class="teacher-title workspace-page-title student-analysis-title"'
    )
    expect(studentAnalysisSource).not.toContain(
      '查看当前学员的学习进度、推荐任务、题解与审核信息。'
    )
    expect(studentAnalysisOverviewHeroSource).toMatch(
      /\.student-analysis-title\s*\{[\s\S]*--workspace-page-title-margin-top:\s*0;[\s\S]*max-width:\s*min\(100%,\s*38rem\);/s
    )
    expect(sectionCardSource).toContain("type SectionCardVariant = 'default' | 'teacher-flat' | 'teacher-surface'")
    expect(sectionCardSource).toMatch(
      /\.section-card--teacher-flat\s*\{[\s\S]*--section-card-border-top-color:\s*color-mix\(in srgb,\s*var\(--teacher-divider\)\s*88%,\s*transparent\);/s
    )
    expect(sectionCardSource).toMatch(
      /\.section-card--teacher-surface\s*\{[\s\S]*--section-card-border:\s*1px solid var\(--teacher-card-border\);[\s\S]*--section-card-header-border-bottom:\s*1px dashed var\(--teacher-divider\);/s
    )
    expect(studentAnalysisSource).not.toMatch(/\.content-pane\s*\{[\s\S]*padding-top:/s)
    expect(studentAnalysisSource).toMatch(
      /\.content-pane\s*\{[\s\S]*flex:\s*1 1 auto;[\s\S]*align-content:\s*start;/s
    )
    expect(studentAnalysisSource).toMatch(
      /\.student-analysis-shell\s*\{[\s\S]*border:\s*0;[\s\S]*background:\s*transparent;[\s\S]*box-shadow:\s*none;[\s\S]*overflow:\s*visible;/s
    )
    expect(studentAnalysisOverviewHeroSource).toMatch(
      /\.summary-strip\s*\{[\s\S]*?margin:\s*var\(--space-6\)\s*0\s*0;[\s\S]*?padding:\s*0;/s
    )
    expect(studentAnalysisOverviewHeroSource).not.toMatch(/\.summary-strip\s*\{[^}]*border-bottom:/s)
    expect(studentAnalysisOverviewHeroSource).toContain(
      'class="summary-card summary-card--solved progress-card metric-panel-card"'
    )
    expect(studentAnalysisOverviewHeroSource).toContain(
      'class="summary-card summary-card--completion progress-card metric-panel-card"'
    )
    expect(studentAnalysisOverviewHeroSource).toContain(
      'class="summary-card summary-card--weakness progress-card metric-panel-card"'
    )
    expect(studentAnalysisOverviewHeroSource).toContain('--metric-panel-border:')
    expect(studentAnalysisOverviewHeroSource).toContain('var(--teacher-card-border)')
    expect(studentAnalysisOverviewHeroSource).toContain('--summary-card-accent: var(--workspace-brand);')
    expect(studentAnalysisOverviewHeroSource).toContain('--summary-card-accent: var(--color-primary);')
    expect(studentAnalysisOverviewHeroSource).toContain('--summary-card-accent: var(--color-success);')
    expect(studentAnalysisOverviewHeroSource).not.toContain('--summary-card-accent: var(--color-warning);')
    expect(studentAnalysisOverviewHeroSource).toContain('--metric-panel-value-color:')
    expect(studentAnalysisSource).not.toContain(':deep(.section-card)')
    expect(studentAnalysisSource).not.toContain(':deep(.section-card__header)')
    expect(studentInsightRecommendationsSectionSource).toContain('variant="teacher-flat"')
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /\.insight-tab-section-card\s*\{[\s\S]*--section-card-border-top-width:\s*0;[\s\S]*--section-card-header-border-bottom:\s*0;/s
    )
    expect(studentAnalysisSource).toContain('StudentAnalysisWorkspaceContent')

    expect(reviewArchiveCombinedSource).toContain('--teacher-card-border:')
    expect(reviewArchiveCombinedSource).toContain('--teacher-divider:')
    expect(reviewArchiveCombinedSource).toContain('--journal-accent: var(--color-primary);')
    expect(reviewArchiveCombinedSource).toContain(
      '--journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));'
    )
    expect(reviewArchiveCombinedSource).toContain('variant="teacher-surface"')
    expect(reviewArchiveCombinedSource).not.toContain(':deep(.section-card)')
    expect(reviewArchiveCombinedSource).not.toContain(':deep(.section-card__header)')
    expect(reviewArchiveCombinedSource).toContain('metric-panel-card')
    expect(reviewArchiveCombinedSource).toContain('--metric-panel-border: var(--teacher-card-border);')
    expect(reviewArchiveCombinedSource).toContain('class="summary-grid metric-panel-grid metric-panel-default-surface"')
    expect(reviewArchiveCombinedSource).toContain(
      'class="summary-card progress-card metric-panel-card"'
    )
    expect(reviewArchiveCombinedSource).toContain('class="summary-card__label progress-card-label metric-panel-label"')
    expect(reviewArchiveCombinedSource).toContain('class="summary-card__value progress-card-value metric-panel-value"')
    expect(reviewArchiveCombinedSource).toContain('class="summary-card__hint progress-card-hint metric-panel-helper"')
    expect(reviewArchiveCombinedSource).not.toContain('--journal-accent: #2563eb;')
    expect(reviewArchiveCombinedSource).not.toContain('--journal-accent-strong: #1d4ed8;')
    expect(reviewArchiveCombinedSource).not.toContain(
      'color-mix(in srgb, #f59e0b 14%, var(--journal-surface))'
    )
  })

  it('teacher detail panels should use softened panel border fallbacks instead of bright rgba fallback lines', () => {
    expect(teacherPanelShellSource).toMatch(
      /--panel-border:\s*color-mix\(\s*in srgb,\s*var\(--journal-border,\s*var\(--color-border-default\)\) 74%,\s*transparent\s*\);/
    )
    expect(classTrendPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(classInsightsPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(classReviewPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(interventionPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(studentInsightPanelSource).toContain('--teacher-card-border:')
    expect(studentInsightPanelSource).toContain('--teacher-divider:')
    expect(studentInsightOverviewSectionSource).toMatch(
      /\.insight-overview-layout\s*\{[\s\S]*display:\s*grid;[\s\S]*gap:\s*var\(--space-6\);/s
    )
    expect(studentInsightOverviewSectionSource).not.toMatch(
      /\.insight-overview-layout\s*\{[\s\S]*border-top:/s
    )
    expect(studentInsightOverviewSectionSource).toContain('variant="teacher-flat"')
    expect(studentInsightOverviewSectionSource).toMatch(
      /\.insight-overview-card\s*\{[\s\S]*?--section-card-border-top-width:\s*0;/s
    )
    expect(studentInsightPanelSource).not.toMatch(/\.insight-rate-panel\s*\{[^}]*border-top:/s)
    expect(studentInsightCompositeSource).toMatch(
      /\.insight-kpi-value\s*\{[\s\S]*--metric-panel-value-size:\s*var\(--font-size-1-00\);/s
    )
    expect(studentInsightCompositeSource).toContain(
      'class="writeup-kpi-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(studentInsightWriteupsSource).toMatch(
      /\.writeup-kpi-grid\s*\{[\s\S]*--metric-panel-background:\s*transparent;[\s\S]*--metric-panel-shadow:\s*none;[\s\S]*--metric-panel-radius:\s*0;/s
    )
    expect(studentInsightCompositeSource).toContain(
      'class="insight-kpi-card writeup-kpi-card progress-card metric-panel-card"'
    )
    expect(studentInsightCompositeSource).toContain(
      'class="insight-kpi-label progress-card-label metric-panel-label"'
    )
    expect(studentInsightCompositeSource).toContain(
      'class="insight-kpi-value progress-card-value metric-panel-value"'
    )
    expect(studentInsightCompositeSource).toContain(
      'class="insight-kpi-hint progress-card-hint metric-panel-helper"'
    )
    expect(studentInsightCompositeSource).toContain(
      'class="insight-kpi-grid progress-strip metric-panel-grid metric-panel-default-surface md:grid-cols-3"'
    )
    expect(studentInsightPanelSource).toContain(':loading="loading"')
    expect(studentInsightPanelSource).not.toContain('insight-loading-shell')
    expect(studentInsightPrimarySectionsSource).toContain('insight-timeline-glass')
    expect(studentInsightManualReviewSource).toMatch(
      /\.insight-kpi-grid\s*\{[\s\S]*--metric-panel-background:\s*transparent;[\s\S]*--metric-panel-shadow:\s*none;[\s\S]*--metric-panel-radius:\s*0;/s
    )
    expect(studentInsightManualReviewSource).toContain('class="manual-review-detail-shell"')
    expect(studentInsightManualReviewSource).toMatch(
      /\.manual-review-detail-shell\s*\{[\s\S]*border-top:\s*1px solid color-mix\(in srgb,\s*var\(--teacher-divider\)\s*88%,\s*transparent\);[\s\S]*background:\s*transparent;/s
    )
    expect(studentInsightCompositeSource).toContain(
      'class="insight-kpi-card progress-card metric-panel-card"'
    )
    expect(studentInsightWriteupsSource).toMatch(
      /\.writeup-review-panel\s*\{[\s\S]*background:\s*transparent;/s
    )
    expect(studentInsightWriteupsSource).toMatch(
      /\.insight-manual-input\s*\{[\s\S]*background:\s*transparent;/s
    )
    expect(studentInsightManualReviewSource).toMatch(
      /\.insight-manual-input\s*\{[\s\S]*background:\s*transparent;/s
    )
    expect(studentInsightManualReviewSource).toContain(
      'class="ui-btn ui-btn--secondary insight-outline-action disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(studentInsightManualReviewSource).toContain(
      'class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(studentInsightPanelSource).not.toContain('challenge-btn-outline')
    expect(studentInsightPanelSource).not.toContain('challenge-btn-primary')
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-list workspace-directory-list"'
    )
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /<div[\s\S]*class="insight-recommendation-list workspace-directory-list"[\s\S]*<template v-if="loading">[\s\S]*<AppEmpty[\s\S]*v-else-if="recommendations\.length === 0"[\s\S]*<template v-else>/s
    )
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /\.insight-recommendation-list\s*\{[\s\S]*--workspace-directory-shell-border:\s*color-mix\(in srgb,\s*var\(--teacher-card-border\)\s*88%,\s*transparent\);[\s\S]*--workspace-directory-shell-background:[\s\S]*radial-gradient\([\s\S]*linear-gradient\([\s\S]*--workspace-directory-shell-radius:\s*var\(--workspace-radius-lg\);[\s\S]*overflow:\s*hidden;[\s\S]*box-shadow:\s*var\(--workspace-shadow-panel\);/s
    )
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /\.insight-recommendation-empty\s*\{[\s\S]*--app-empty-border-top:\s*0;[\s\S]*--app-empty-border-bottom:\s*0;[\s\S]*--app-empty-background:\s*transparent;/s
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      '<div class="insight-recommendation-skeleton-head">'
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-skeleton-row"'
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-skeleton-pills"'
    )
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /\.insight-recommendation-skeleton-row\s*\{[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s*auto\s*auto;[\s\S]*border-bottom:\s*1px solid var\(--workspace-directory-row-divider\);/s
    )
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /\.insight-recommendation-skeleton-kicker,[\s\S]*\.insight-recommendation-skeleton-action\s*\{[\s\S]*background-size:\s*220% 100%;[\s\S]*animation:\s*insightRecommendationSkeletonSweep 1\.55s ease-in-out infinite;/s
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      '@media (prefers-reduced-motion: reduce)'
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-row workspace-directory-grid-row"'
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="workspace-directory-row-btn insight-recommendation-action"'
    )
    expect(studentInsightPanelSource).not.toContain('variant="action"')
    expect(studentInsightPanelSource).not.toContain('--metric-panel-background')
    expect(studentInsightPanelSource).not.toContain('insight-kpi-card--primary')

    expect(classTrendPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
    expect(classInsightsPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
    expect(classReviewPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
  })

  it('teacher detail panels should inherit shared journal tokens instead of carrying local hex fallbacks', () => {
    for (const source of [
      classTrendPanelSource,
      classInsightsPanelSource,
      classReviewPanelSource,
      interventionPanelSource,
    ]) {
      expect(source).not.toContain('--panel-ink: var(--journal-ink, #0f172a);')
      expect(source).not.toContain('--panel-muted: var(--journal-muted, #64748b);')
      expect(source).not.toContain('--panel-accent: var(--journal-accent, #4f46e5);')
      expect(source).not.toContain('--panel-accent-strong: var(--journal-accent-strong, #4338ca);')
    }

    expect(teacherPanelShellSource).toContain('--panel-ink: var(--journal-ink);')
    expect(teacherPanelShellSource).toContain('--panel-muted: var(--journal-muted);')
    expect(teacherPanelShellSource).toContain('--panel-accent: var(--journal-accent);')
    expect(teacherPanelShellSource).toContain(
      '--panel-accent-strong: var(--journal-accent-strong);'
    )
  })

  it('teacher class workspace review insight intervention cards should reuse shared border tokens', () => {
    expect(classInsightsPanelSource).toContain('--showcase-panel-border: var(--panel-border);')
    expect(classInsightsPanelSource).toMatch(
      /\.teacher-subsection--bare\s*\{[\s\S]*border:\s*1px solid var\(--panel-border\);/s
    )
    expect(classInsightsPanelSource).not.toMatch(
      /\.teacher-subsection--bare:hover\s*\{[\s\S]*border-color:/s
    )

    expect(classReviewPanelSource).toMatch(
      /\.review-item\s*\{[\s\S]*border:\s*1px solid var\(--panel-border\);/s
    )
    expect(classReviewPanelSource).toMatch(
      /\.review-item__recommendation--premium\s*\{[\s\S]*border-top:\s*1px solid var\(--panel-divider\);/s
    )
    expect(classReviewPanelSource).not.toContain(
      'border: 1px solid color-mix(in srgb, var(--review-accent) 12%, var(--panel-border));'
    )
    expect(classReviewPanelSource).not.toContain(
      'border-color: color-mix(in srgb, var(--review-accent) 30%, var(--panel-border));'
    )

    expect(interventionPanelSource).toMatch(
      /\.intervention-item\s*\{[\s\S]*border:\s*1px solid var\(--panel-border\);/s
    )
    expect(interventionPanelSource).toMatch(
      /\.intervention-item__recommendation--premium\s*\{[\s\S]*border:\s*1px solid var\(--panel-border\);/s
    )
    expect(interventionPanelSource).not.toContain(
      'border: 1px solid color-mix(in srgb, var(--intervention-accent) 12%, var(--panel-border));'
    )
    expect(interventionPanelSource).not.toContain(
      'border-color: color-mix(in srgb, var(--intervention-accent) 28%, var(--panel-border));'
    )
  })
})
