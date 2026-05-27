import { existsSync, readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classStudentsSourceBase from '@/features/class-students-workspace/ui/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/components/teacher/class-management/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/components/teacher/class-management/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/components/teacher/class-management/ClassStudentsDirectoryPanel.vue?raw'
import teacherDashboardSourceBase from '@/features/teacher-dashboard/ui/TeacherDashboardPage.vue?raw'
import teacherDashboardPortraitPanelSource from '@/components/teacher/dashboard/TeacherDashboardPortraitPanel.vue?raw'
import teacherDashboardStudentInsightPanelSource from '@/components/teacher/dashboard/TeacherDashboardStudentInsightPanel.vue?raw'
import teacherDashboardTrendPanelSource from '@/components/teacher/dashboard/TeacherDashboardTrendPanel.vue?raw'
import teacherDashboardReviewPanelSource from '@/components/teacher/dashboard/TeacherDashboardReviewPanel.vue?raw'
import teacherDashboardInterventionPanelSource from '@/components/teacher/dashboard/TeacherDashboardInterventionPanel.vue?raw'

const teacherWorkspaceSubpanelPath = `${process.cwd()}/src/components/teacher/teacher-workspace-subpanel.css`
const classStudentsSource = [
  classStudentsSourceBase,
  classStudentsOverviewPanelSource,
  classStudentsInsightWindowPanelSource,
  classStudentsDirectoryPanelSource,
].join('\n')
const teacherDashboardSource = [
  teacherDashboardSourceBase,
  teacherDashboardPortraitPanelSource,
  teacherDashboardStudentInsightPanelSource,
  teacherDashboardTrendPanelSource,
  teacherDashboardReviewPanelSource,
  teacherDashboardInterventionPanelSource,
].join('\n')

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

    expect(classStudentsSource).toContain("@import '../teacher-workspace-subpanel.css';")
    expect(teacherDashboardSource).toContain(
      "@import '@/components/teacher/teacher-workspace-subpanel.css';"
    )

    for (const source of [classStudentsSource, teacherDashboardSource]) {
      expect(source).not.toMatch(/\.workspace-subpanel\s*:deep\(\.teacher-panel\)\s*\{/s)
      expect(source).not.toMatch(/\.workspace-subpanel\s*:deep\(\.journal-eyebrow\)\s*\{/s)
      expect(source).not.toMatch(/\.workspace-subpanel--flat\s*:deep\(\.teacher-panel\)\s*\{/s)
    }
  })
})
