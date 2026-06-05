import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import studentManagementSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'
import dashboardSourceBase from '@/features/teacher/dashboard/ui/TeacherDashboardPage.vue?raw'
import teacherDashboardPortraitPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue?raw'
import teacherDashboardStudentInsightPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue?raw'
import teacherDashboardTrendPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue?raw'
import teacherDashboardReviewPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue?raw'
import teacherDashboardInterventionPanelSource from '@/features/teacher/dashboard/ui/TeacherDashboardInterventionPanel.vue?raw'
import instanceManagementSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'
import awdReviewSummaryPanelSource from '@/widgets/awd-review-workspace/AwdReviewSummaryPanel.vue?raw'

const dashboardSource = [
  dashboardSourceBase,
  teacherDashboardPortraitPanelSource,
  teacherDashboardStudentInsightPanelSource,
  teacherDashboardTrendPanelSource,
  teacherDashboardReviewPanelSource,
  teacherDashboardInterventionPanelSource,
].join('\n')

const instanceManagementSource = [
  instanceManagementSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

describe('teacher base surface alignment', () => {
  it('teacher base pages should keep hero and summary shell owner on shared teacher surface CSS', () => {
    expect(teacherSurfaceSource).toContain('.teacher-management-shell {')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-hero')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-summary')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-summary-grid--header')

    expect(classManagementSource).toContain('teacher-management-shell')
    expect(classManagementSource).not.toMatch(/\.teacher-(?:btn)\s*\{/)
    expect(classManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(classManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)

    expect(studentManagementSource).toContain('teacher-management-shell')
    expect(studentManagementSource).not.toMatch(/\.teacher-(?:btn)\s*\{/)
    expect(studentManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(studentManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)

    expect(dashboardSource).toContain('--teacher-card-border:')
    expect(dashboardSource).toContain('--teacher-control-border:')
    expect(dashboardSource).toContain('--teacher-divider:')
    expect(dashboardSource).not.toMatch(/\.teacher-(?:btn)\s*\{/)

    expect(instanceManagementSource).toContain('teacher-management-shell')
    expect(instanceManagementSource).not.toMatch(/\.teacher-(?:btn)\s*\{/)
    expect(instanceManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(instanceManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)
  })

  it('awd review summary owner should stay on the shared summary panel component', () => {
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewSummaryPanel')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewSummaryPanel')
    expect(awdReviewSummaryPanelSource).toContain('metric-panel-default-surface')
  })

  it('awd review pages should keep page shell owner on the shared teacher surface shell', () => {
    expect(awdReviewSurfaceShellSource).toContain('teacher-management-shell')
    expect(awdReviewSurfaceShellSource).toContain('workspace-shell')
    expect(awdReviewIndexWorkspaceSource).not.toMatch(/\.teacher-(?:btn)\s*\{/m)
    expect(awdReviewIndexWorkspaceSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(awdReviewIndexWorkspaceSource).not.toMatch(/^\.teacher-summary\s*\{/m)
    expect(awdReviewWorkspaceSource).not.toMatch(/\.teacher-(?:btn)\s*\{/m)
    expect(awdReviewWorkspaceSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(awdReviewWorkspaceSource).not.toMatch(/^\.teacher-summary\s*\{/m)
  })
})
