import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classStudentsSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import classTrendPanelSource from '@/components/teacher/TeacherClassTrendPanel.vue?raw'
import classInsightsPanelSource from '@/components/teacher/TeacherClassInsightsPanel.vue?raw'
import classReviewPanelSource from '@/components/teacher/TeacherClassReviewPanel.vue?raw'
import studentInsightPanelSource from '@/components/teacher/StudentInsightPanel.vue?raw'
import reviewArchiveSource from '@/views/teacher/TeacherStudentReviewArchive.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

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
    expect(studentAnalysisSource).toContain('class="workspace-topbar"')
    expect(studentAnalysisSource).toContain('class="top-tabs"')
    expect(studentAnalysisSource).toMatch(
      /:deep\(\.section-card\)\s*\{[\s\S]*border-top:\s*1px solid color-mix\(in srgb,\s*var\(--teacher-divider\)\s*90%,\s*transparent\);/s
    )
    expect(studentAnalysisSource).toMatch(
      /\.summary-strip\s*\{[\s\S]*?margin:\s*0 0 var\(--space-5\);[\s\S]*?padding:\s*var\(--space-1\) 0 0;/s
    )
    expect(studentAnalysisSource).not.toMatch(
      /\.summary-strip\s*\{[^}]*border-bottom:/s
    )
    expect(studentAnalysisSource).toContain('class="summary-card metric-panel-card"')
    expect(studentAnalysisSource).toContain('--metric-panel-border: var(--teacher-card-border);')
    expect(studentAnalysisSource).toMatch(
      /:deep\(\.section-card__header\)\s*\{[\s\S]*border-bottom:\s*1px dashed color-mix\(in srgb,\s*var\(--teacher-divider\)\s*86%,\s*transparent\);/s
    )

    expect(reviewArchiveSource).toContain('--teacher-card-border:')
    expect(reviewArchiveSource).toContain('--teacher-divider:')
    expect(reviewArchiveSource).toMatch(
      /:deep\(\.section-card\)\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(reviewArchiveSource).toMatch(
      /:deep\(\.section-card__header\)\s*\{[\s\S]*border-bottom:\s*1px dashed var\(--teacher-divider\);/s
    )
    expect(reviewArchiveSource).toContain('metric-panel-card')
    expect(reviewArchiveSource).toContain('--metric-panel-border: var(--teacher-card-border);')
    expect(reviewArchiveSource).toContain('class="summary-grid metric-panel-grid"')
    expect(reviewArchiveSource).toContain('class="summary-card summary-card--primary metric-panel-card"')
    expect(reviewArchiveSource).toContain('class="summary-card__label metric-panel-label"')
    expect(reviewArchiveSource).toContain('class="summary-card__value metric-panel-value"')
    expect(reviewArchiveSource).toContain('class="summary-card__hint metric-panel-helper"')
  })

  it('teacher detail panels should use softened panel border fallbacks instead of bright rgba fallback lines', () => {
    expect(classTrendPanelSource).toMatch(
      /--panel-border:\s*color-mix\(\s*in srgb,\s*var\(--journal-border,\s*var\(--color-border-default\)\) 74%,\s*transparent\s*\);/
    )
    expect(classInsightsPanelSource).toMatch(
      /--panel-border:\s*color-mix\(\s*in srgb,\s*var\(--journal-border,\s*var\(--color-border-default\)\) 74%,\s*transparent\s*\);/
    )
    expect(classReviewPanelSource).toMatch(
      /--panel-border:\s*color-mix\(\s*in srgb,\s*var\(--journal-border,\s*var\(--color-border-default\)\) 74%,\s*transparent\s*\);/
    )
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
    expect(studentInsightPanelSource).not.toMatch(
      /\.insight-rate-panel\s*\{[^}]*border-top:/s
    )
    expect(studentInsightPanelSource).toContain('class="writeup-kpi-grid metric-panel-grid"')
    expect(studentInsightPanelSource).toContain(
      'class="insight-kpi-card writeup-kpi-card insight-kpi-card--primary metric-panel-card"'
    )
    expect(studentInsightPanelSource).toContain('class="insight-kpi-label metric-panel-label"')
    expect(studentInsightPanelSource).toContain('class="insight-kpi-value metric-panel-value"')
    expect(studentInsightPanelSource).toContain('class="insight-kpi-hint metric-panel-helper"')
    expect(studentInsightPanelSource).toContain('class="insight-kpi-grid metric-panel-grid')
    expect(studentInsightPanelSource).toMatch(
      /\.insight-kpi-card\s*\{[\s\S]*--metric-panel-border:\s*color-mix\(in srgb,\s*var\(--teacher-card-border\)\s*88%,\s*transparent\);/s
    )
    expect(studentInsightPanelSource).toMatch(
      /\.insight-kpi-card\s*\{[\s\S]*--metric-panel-radius:\s*16px;/s
    )
    expect(studentInsightPanelSource).toMatch(
      /\.insight-kpi-value\s*\{[\s\S]*--metric-panel-value-size:\s*var\(--font-size-1-00\);/s
    )

    expect(classTrendPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
    expect(classInsightsPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
    expect(classReviewPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
  })
})
