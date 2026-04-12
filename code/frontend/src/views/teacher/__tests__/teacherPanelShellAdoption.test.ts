import { existsSync, readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classTrendPanelSource from '@/components/teacher/TeacherClassTrendPanel.vue?raw'
import classInsightsPanelSource from '@/components/teacher/TeacherClassInsightsPanel.vue?raw'
import classReviewPanelSource from '@/components/teacher/TeacherClassReviewPanel.vue?raw'
import interventionPanelSource from '@/components/teacher/TeacherInterventionPanel.vue?raw'

const teacherPanelShellPath = `${process.cwd()}/src/components/teacher/teacher-panel-shell.css`

describe('teacher panel shell adoption', () => {
  it('teacher detail panel 应统一复用共享壳层样式，而不是继续各自维护基础 panel 壳子', () => {
    expect(existsSync(teacherPanelShellPath)).toBe(true)

    const teacherPanelShellSource = readFileSync(teacherPanelShellPath, 'utf-8')
    expect(teacherPanelShellSource).toContain('.teacher-panel {')
    expect(teacherPanelShellSource).toContain('.teacher-panel__title {')
    expect(teacherPanelShellSource).toContain('.teacher-panel__subtitle {')

    for (const source of [
      classTrendPanelSource,
      classInsightsPanelSource,
      classReviewPanelSource,
      interventionPanelSource,
    ]) {
      expect(source).toContain("@import './teacher-panel-shell.css';")
      expect(source).not.toMatch(/\.teacher-panel\s*\{/s)
      expect(source).not.toMatch(/\.teacher-panel__header\s*\{/s)
      expect(source).not.toMatch(/\.teacher-panel__title\s*\{/s)
      expect(source).not.toMatch(/\.teacher-panel__subtitle\s*\{/s)
    }
  })
})
