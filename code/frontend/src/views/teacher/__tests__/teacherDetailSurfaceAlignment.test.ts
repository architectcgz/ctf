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
    expect(teacherSurfaceSource).toMatch(
      /\.teacher-btn\s*\{[\s\S]*border:\s*1px solid var\(--teacher-control-border\);/s
    )

    expect(classStudentsSource).toContain('--teacher-card-border:')
    expect(classStudentsSource).toContain('--teacher-control-border:')
    expect(classStudentsSource).toContain('--teacher-divider:')
    expect(classStudentsSource).not.toContain('.teacher-btn {')
    expect(classStudentsSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(classStudentsSource).toMatch(
      /\.teacher-table-shell\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )

    expect(studentAnalysisSource).toContain('--teacher-card-border:')
    expect(studentAnalysisSource).toContain('--teacher-divider:')
    expect(studentAnalysisSource).toContain('class="workspace-shell journal-eyebrow-text"')
    expect(studentAnalysisSource).not.toContain('class="workspace-topbar"')
    expect(studentAnalysisSource).toContain('class="top-tabs"')
    expect(studentAnalysisSource).toMatch(
      /:deep\(\.section-card\)\s*\{[\s\S]*border-top:\s*1px solid color-mix\(in srgb,\s*var\(--teacher-divider\)\s*90%,\s*transparent\);/s
    )
    expect(studentAnalysisSource).toMatch(
      /\.summary-strip\s*\{[\s\S]*?margin:\s*0 0 var\(--space-5\);[\s\S]*?padding:\s*var\(--space-1\) 0 0;/s
    )
    expect(studentAnalysisSource).not.toMatch(/\.summary-strip\s*\{[^}]*border-bottom:/s)
    expect(studentAnalysisSource).toContain('class="summary-card progress-card metric-panel-card"')
    expect(studentAnalysisSource).toContain('--metric-panel-border: var(--teacher-card-border);')
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
})
