import { describe, expect, it } from 'vitest'

import classStudentsSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import classTrendPanelSource from '@/components/teacher/TeacherClassTrendPanel.vue?raw'
import classInsightsPanelSource from '@/components/teacher/TeacherClassInsightsPanel.vue?raw'
import classReviewPanelSource from '@/components/teacher/TeacherClassReviewPanel.vue?raw'
import studentInsightPanelSource from '@/components/teacher/StudentInsightPanel.vue?raw'

describe('teacher detail surface alignment', () => {
  it('class students and student analysis pages should soften control and section borders', () => {
    expect(classStudentsSource).toContain('--teacher-card-border:')
    expect(classStudentsSource).toContain('--teacher-control-border:')
    expect(classStudentsSource).toContain('--teacher-divider:')
    expect(classStudentsSource).toMatch(/\.teacher-btn\s*\{[\s\S]*border:\s*1px solid var\(--teacher-control-border\);/s)
    expect(classStudentsSource).toMatch(/\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s)
    expect(classStudentsSource).toMatch(/\.teacher-table-shell\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s)

    expect(studentAnalysisSource).toContain('--teacher-card-border:')
    expect(studentAnalysisSource).toContain('--teacher-divider:')
    expect(studentAnalysisSource).toMatch(/:deep\(\.page-header\)\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s)
    expect(studentAnalysisSource).toMatch(/:deep\(\.section-card\)\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s)
    expect(studentAnalysisSource).toMatch(/\.analysis-note\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s)
    expect(studentAnalysisSource).toMatch(/:deep\(\.section-card__header\)\s*\{[\s\S]*border-bottom:\s*1px dashed var\(--teacher-divider\);/s)
  })

  it('teacher detail panels should use softened panel border fallbacks instead of bright rgba fallback lines', () => {
    expect(classTrendPanelSource).toMatch(/--panel-border:\s*color-mix\(in srgb,\s*var\(--journal-border,\s*var\(--color-border-default\)\) 74%, transparent\);/)
    expect(classInsightsPanelSource).toMatch(/--panel-border:\s*color-mix\(in srgb,\s*var\(--journal-border,\s*var\(--color-border-default\)\) 74%, transparent\);/)
    expect(classReviewPanelSource).toMatch(/--panel-border:\s*color-mix\(in srgb,\s*var\(--journal-border,\s*var\(--color-border-default\)\) 74%, transparent\);/)
    expect(studentInsightPanelSource).toContain('--teacher-card-border:')
    expect(studentInsightPanelSource).toContain('--teacher-divider:')
    expect(studentInsightPanelSource).toMatch(/:deep\(\.section-card\)\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s)
    expect(studentInsightPanelSource).toMatch(/\.insight-kpi-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s)

    expect(classTrendPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
    expect(classInsightsPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
    expect(classReviewPanelSource).not.toContain('rgba(226, 232, 240, 0.8)')
  })
})
