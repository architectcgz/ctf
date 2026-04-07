import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import dashboardSource from '@/components/teacher/dashboard/TeacherDashboardPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import reportExportSource from '@/views/teacher/ReportExport.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

describe('teacher base surface alignment', () => {
  it('teacher base pages should soften control, card, and divider borders', () => {
    expect(teacherSurfaceSource).toMatch(
      /\.teacher-btn\s*\{[\s\S]*border:\s*1px solid var\(--teacher-control-border\);/s
    )
    expect(teacherSurfaceSource).toContain('.teacher-management-shell {')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-hero')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-summary')

    expect(classManagementSource).toContain('teacher-management-shell')
    expect(classManagementSource).not.toContain('.teacher-btn {')
    expect(classManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(classManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)
    expect(classManagementSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(classManagementSource).toMatch(
      /\.teacher-tip-block\s*\{[\s\S]*border-top:\s*1px dashed var\(--teacher-divider\);/s
    )

    expect(studentManagementSource).toContain('teacher-management-shell')
    expect(studentManagementSource).not.toContain('.teacher-btn {')
    expect(studentManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(studentManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)
    expect(studentManagementSource).toMatch(
      /\.teacher-hero-divider\s*\{[\s\S]*border-top:\s*1px dashed var\(--teacher-divider\);/s
    )

    expect(dashboardSource).toContain('--teacher-card-border:')
    expect(dashboardSource).toContain('--teacher-control-border:')
    expect(dashboardSource).toContain('--teacher-divider:')
    expect(dashboardSource).toMatch(
      /\.teacher-btn\s*\{[\s\S]*border:\s*1px solid var\(--teacher-control-border\);/s
    )
    expect(dashboardSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(dashboardSource).toMatch(
      /\.teacher-tip-block\s*\{[\s\S]*border-top:\s*1px dashed var\(--teacher-divider\);/s
    )

    expect(instanceManagementSource).toContain('teacher-management-shell')
    expect(instanceManagementSource).not.toContain('.teacher-btn {')
    expect(instanceManagementSource).toContain('--teacher-management-hero-border: var(--teacher-card-border);')
    expect(instanceManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
    expect(instanceManagementSource).not.toMatch(/^\.teacher-summary\s*\{/m)
    expect(instanceManagementSource).toMatch(
      /\.teacher-badge-card\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(instanceManagementSource).toMatch(
      /\.teacher-tip-block\s*\{[\s\S]*border-top:\s*1px dashed var\(--teacher-divider\);/s
    )
  })

  it('report export should soften page header, cards, divider, and dialog borders', () => {
    expect(reportExportSource).toContain('--report-card-border:')
    expect(reportExportSource).toContain('--report-divider:')
    expect(reportExportSource).not.toMatch(/^\.report-hero\s*\{/m)
    expect(reportExportSource).not.toMatch(/^\.report-summary\s*\{/m)
    expect(reportExportSource).toMatch(
      /:deep\(\.page-header\)\s*\{[\s\S]*border:\s*1px solid var\(--report-card-border\);/s
    )
    expect(reportExportSource).toMatch(
      /\.report-note\s*\{[\s\S]*border:\s*1px solid var\(--report-card-border\);/s
    )
    expect(reportExportSource).toMatch(
      /\.report-hero-divider\s*\{[\s\S]*border-top:\s*1px dashed var\(--report-divider\);/s
    )
    expect(reportExportSource).toMatch(
      /:deep\(\.report-preview-dialog \.el-dialog\)\s*\{[\s\S]*border:\s*1px solid var\(--report-card-border\);/s
    )
    expect(reportExportSource).toMatch(
      /\.report-kpi-card\s*\{[\s\S]*border:\s*1px solid var\(--report-card-border\);/s
    )
  })
})
