import { existsSync, readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classStudentsSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import teacherDashboardSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'

const teacherWorkspaceSubpanelPath = `${process.cwd()}/src/components/teacher/teacher-workspace-subpanel.css`

describe('teacher workspace subpanel adoption', () => {
  it('teacher workspace 页面应统一复用共享 subpanel 壳层样式，而不是继续各自维护深选择器块', () => {
    expect(existsSync(teacherWorkspaceSubpanelPath)).toBe(true)

    const teacherWorkspaceSubpanelSource = readFileSync(teacherWorkspaceSubpanelPath, 'utf-8')
    expect(teacherWorkspaceSubpanelSource).toContain('.workspace-subpanel {')
    expect(teacherWorkspaceSubpanelSource).toContain(
      '--teacher-workspace-panel-border: var(--teacher-card-border, var(--panel-border));'
    )
    expect(teacherWorkspaceSubpanelSource).toContain(
      '--teacher-workspace-line-soft: color-mix('
    )
    expect(teacherWorkspaceSubpanelSource).toContain(
      '--teacher-workspace-review-background: linear-gradient('
    )
    expect(teacherWorkspaceSubpanelSource).toContain('.workspace-subpanel :deep(.teacher-panel) {')
    expect(teacherWorkspaceSubpanelSource).toContain(
      '.workspace-subpanel--flat :deep(.teacher-panel) {'
    )
    expect(teacherWorkspaceSubpanelSource).not.toContain(
      '.workspace-subpanel :deep(.journal-eyebrow) {'
    )
    expect(teacherWorkspaceSubpanelSource).toContain(
      '.workspace-subpanel :deep(.teacher-panel__header > .teacher-panel__title:first-child),'
    )

    for (const source of [classStudentsSource, teacherDashboardSource]) {
      expect(source).toContain("@import '../teacher-workspace-subpanel.css';")
      expect(source).not.toMatch(/\.workspace-subpanel\s*:deep\(\.teacher-panel\)\s*\{/s)
      expect(source).not.toMatch(/\.workspace-subpanel\s*:deep\(\.journal-eyebrow\)\s*\{/s)
      expect(source).not.toMatch(/\.workspace-subpanel--flat\s*:deep\(\.teacher-panel\)\s*\{/s)
    }
  })
})
