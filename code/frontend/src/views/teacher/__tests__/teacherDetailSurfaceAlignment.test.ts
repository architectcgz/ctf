import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classStudentsSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import classTrendPanelSource from '@/components/teacher/TeacherClassTrendPanel.vue?raw'
import classInsightsPanelSource from '@/components/teacher/TeacherClassInsightsPanel.vue?raw'
import classReviewPanelSource from '@/components/teacher/TeacherClassReviewPanel.vue?raw'
import interventionPanelSource from '@/components/teacher/TeacherInterventionPanel.vue?raw'
import studentInsightPanelSource from '@/components/teacher/StudentInsightPanel.vue?raw'
import studentInsightWriteupsSource from '@/components/teacher/student-insight/StudentInsightWriteupsSection.vue?raw'
import studentInsightManualReviewSource from '@/components/teacher/student-insight/StudentInsightManualReviewSection.vue?raw'
import reviewArchiveSource from '@/views/teacher/TeacherStudentReviewArchive.vue?raw'
import reviewArchiveWorkspaceSource from '@/widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue?raw'
import reviewArchiveSummarySectionSource from '@/widgets/teacher-review-archive/TeacherReviewArchiveSummarySection.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)
const styleSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')
const teacherPanelShellSource = readFileSync(
  `${process.cwd()}/src/components/teacher/teacher-panel-shell.css`,
  'utf-8'
)
const reviewArchiveCombinedSource = [
  reviewArchiveSource,
  reviewArchiveWorkspaceSource,
  reviewArchiveSummarySectionSource,
].join('\n')
const studentInsightCompositeSource = [
  studentInsightPanelSource,
  studentInsightWriteupsSource,
  studentInsightManualReviewSource,
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
    expect(studentAnalysisSource).toContain('class="workspace-shell journal-eyebrow-text"')
    expect(studentAnalysisSource).not.toContain('class="workspace-topbar"')
    expect(studentAnalysisSource).toContain('class="top-tabs"')
    expect(studentAnalysisSource).toContain(
      'class="teacher-title workspace-page-title student-analysis-title"'
    )
    expect(studentAnalysisSource).not.toContain(
      '查看当前学员的学习进度、推荐任务、题解与审核信息。'
    )
    expect(studentAnalysisSource).toMatch(
      /\.student-analysis-title\s*\{[\s\S]*--workspace-page-title-margin-top:\s*0;[\s\S]*max-width:\s*min\(100%,\s*38rem\);/s
    )
    expect(studentAnalysisSource).toMatch(
      /:deep\(\.section-card\)\s*\{[\s\S]*border-top:\s*1px solid color-mix\(in srgb,\s*var\(--teacher-divider\)\s*90%,\s*transparent\);/s
    )
    expect(studentAnalysisSource).toMatch(
      /\.content-pane\s*\{[\s\S]*padding-top:\s*var\(--workspace-tabs-panel-gap,\s*var\(--workspace-tab-panel-gap-top-tight\)\);/s
    )
    expect(studentAnalysisSource).toMatch(
      /\.summary-strip\s*\{[\s\S]*?margin:\s*0 0 var\(--space-5\);[\s\S]*?padding:\s*var\(--space-1\) 0 0;/s
    )
    expect(studentAnalysisSource).not.toMatch(/\.summary-strip\s*\{[^}]*border-bottom:/s)
    expect(studentAnalysisSource).toContain(
      'class="summary-card summary-card--solved progress-card metric-panel-card"'
    )
    expect(studentAnalysisSource).toContain(
      'class="summary-card summary-card--completion progress-card metric-panel-card"'
    )
    expect(studentAnalysisSource).toContain(
      'class="summary-card summary-card--weakness progress-card metric-panel-card"'
    )
    expect(studentAnalysisSource).toContain('--metric-panel-border:')
    expect(studentAnalysisSource).toContain('var(--teacher-card-border)')
    expect(studentAnalysisSource).toContain('--summary-card-accent: var(--workspace-brand);')
    expect(studentAnalysisSource).toContain('--summary-card-accent: var(--color-primary);')
    expect(studentAnalysisSource).toContain('--summary-card-accent: var(--color-success);')
    expect(studentAnalysisSource).not.toContain('--summary-card-accent: var(--color-warning);')
    expect(studentAnalysisSource).toContain('--metric-panel-value-color:')
    expect(studentAnalysisSource).toMatch(
      /:deep\(\.section-card__header\)\s*\{[\s\S]*border-bottom:\s*1px dashed color-mix\(in srgb,\s*var\(--teacher-divider\)\s*86%,\s*transparent\);/s
    )

    expect(reviewArchiveCombinedSource).toContain('--teacher-card-border:')
    expect(reviewArchiveCombinedSource).toContain('--teacher-divider:')
    expect(reviewArchiveCombinedSource).toContain('--journal-accent: var(--color-primary);')
    expect(reviewArchiveCombinedSource).toContain(
      '--journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));'
    )
    expect(reviewArchiveCombinedSource).toMatch(
      /:deep\(\.section-card\)\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(reviewArchiveCombinedSource).toMatch(
      /:deep\(\.section-card__header\)\s*\{[\s\S]*border-bottom:\s*1px dashed var\(--teacher-divider\);/s
    )
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
    expect(classTrendPanelSource).toContain("@import './teacher-panel-shell.css';")
    expect(classInsightsPanelSource).toContain("@import './teacher-panel-shell.css';")
    expect(classReviewPanelSource).toContain("@import './teacher-panel-shell.css';")
    expect(interventionPanelSource).toContain("@import './teacher-panel-shell.css';")
    expect(studentInsightPanelSource).toContain('--teacher-card-border:')
    expect(studentInsightPanelSource).toContain('--teacher-divider:')
    expect(studentInsightPanelSource).toMatch(
      /:deep\(\.section-card\)\s*\{[\s\S]*border-top:\s*1px solid color-mix\(in srgb,\s*var\(--teacher-divider\)\s*88%,\s*transparent\);/s
    )
    expect(studentInsightPanelSource).toMatch(
      /\.insight-overview-layout\s*\{[\s\S]*?border-top:\s*1px solid color-mix\(in srgb,\s*var\(--teacher-divider\)\s*88%,\s*transparent\);/s
    )
    expect(studentInsightPanelSource).toMatch(
      /\.insight-overview-layout\s*:deep\(\.section-card\)\s*\{[\s\S]*?border-top:\s*0;/s
    )
    expect(studentInsightPanelSource).not.toMatch(/\.insight-rate-panel\s*\{[^}]*border-top:/s)
    expect(studentInsightCompositeSource).toMatch(
      /\.insight-kpi-value\s*\{[\s\S]*--metric-panel-value-size:\s*var\(--font-size-1-00\);/s
    )
    expect(studentInsightCompositeSource).toContain(
      'class="writeup-kpi-grid progress-strip metric-panel-grid metric-panel-default-surface"'
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
    expect(studentInsightCompositeSource).toContain(
      'class="insight-kpi-card progress-card metric-panel-card"'
    )
    expect(studentInsightManualReviewSource).toContain(
      'class="ui-btn ui-btn--secondary insight-outline-action disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(studentInsightManualReviewSource).toContain(
      'class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"'
    )
    expect(studentInsightPanelSource).not.toContain('challenge-btn-outline')
    expect(studentInsightPanelSource).not.toContain('challenge-btn-primary')
    expect(studentInsightPanelSource).toContain(
      'class="insight-recommendation-list workspace-directory-list"'
    )
    expect(studentInsightPanelSource).toContain(
      'class="insight-recommendation-row workspace-directory-grid-row"'
    )
    expect(studentInsightPanelSource).toContain(
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
