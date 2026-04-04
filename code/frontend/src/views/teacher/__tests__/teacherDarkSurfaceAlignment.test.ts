import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import reportExportSource from '@/views/teacher/ReportExport.vue?raw'
import teacherSurfaceSource from '@/assets/styles/teacher-surface.css?raw'

const tableWrapperPattern =
  /el-table__inner-wrapper[\s\S]*el-table__body-wrapper[\s\S]*el-table__header-wrapper[\s\S]*el-table__empty-block/s

describe('teacher dark surface alignment', () => {
  it('teacher management pages should use shared teacher surface classes', () => {
    expect(classManagementSource).toContain('teacher-surface')
    expect(studentManagementSource).toContain('teacher-surface')
    expect(instanceManagementSource).toContain('teacher-surface')
    expect(reportExportSource).toContain('teacher-surface')
  })

  it('target pages should reuse dashboard journal surface vocabulary instead of page-local skins', () => {
    expect(classManagementSource).toContain('journal-brief')
    expect(classManagementSource).toContain('journal-metric')
    expect(studentManagementSource).toContain('journal-eyebrow')
    expect(studentManagementSource).toContain('journal-metric')
    expect(instanceManagementSource).toContain('journal-brief')
    expect(instanceManagementSource).toContain('journal-metric')
    expect(reportExportSource).toContain('journal-eyebrow')
    expect(reportExportSource).toContain('journal-brief')
    expect(reportExportSource).toContain('journal-metric')
    expect(reportExportSource).not.toContain('report-kpi-card--success')
    expect(reportExportSource).not.toContain('report-kpi-card--warning')
    expect(reportExportSource).not.toContain('report-kpi-card--primary')
  })

  it('class management should not leak element-plus primary plain button chrome', () => {
    expect(classManagementSource).not.toContain('<ElButton type="primary" plain')
    expect(classManagementSource).toContain("class=\"teacher-btn teacher-surface-btn\"")
  })

  it('shared teacher surface should not remap base theme background tokens darker than dashboard', () => {
    expect(teacherSurfaceSource).not.toMatch(/--color-bg-surface:\s*var\(--journal-surface\);/)
    expect(teacherSurfaceSource).not.toMatch(/--color-bg-base:\s*var\(--theme-bg-base\);/)
    expect(teacherSurfaceSource).not.toMatch(/--color-border-default:\s*var\(--journal-border\);/)
    expect(classManagementSource).not.toMatch(/--color-bg-surface:\s*var\(--journal-surface\);/)
    expect(classManagementSource).not.toMatch(/--color-bg-base:\s*var\(--theme-bg-base\);/)
    expect(studentManagementSource).not.toMatch(/--color-bg-surface:\s*var\(--journal-surface\);/)
    expect(studentManagementSource).not.toMatch(/--color-bg-base:\s*var\(--theme-bg-base\);/)
    expect(instanceManagementSource).not.toMatch(/--color-bg-surface:\s*var\(--journal-surface\);/)
    expect(instanceManagementSource).not.toMatch(/--color-bg-base:\s*var\(--theme-bg-base\);/)
  })

  it('teacher tables should cover table wrappers and empty state layers', () => {
    expect(classManagementSource).toMatch(tableWrapperPattern)
    expect(studentManagementSource).toMatch(tableWrapperPattern)
    expect(instanceManagementSource).toMatch(tableWrapperPattern)
  })

  it('student and instance pages should not keep darker or louder local skins than dashboard', () => {
    expect(studentManagementSource).not.toContain('teacher-kpi-card--primary')
    expect(studentManagementSource).not.toContain('teacher-kpi-card--success')
    expect(studentManagementSource).not.toContain('teacher-kpi-card--warning')
    expect(instanceManagementSource).toContain('--teacher-card-border:')
    expect(instanceManagementSource).toMatch(
      /\.teacher-hero\s*\{[\s\S]*border-color:\s*var\(--teacher-card-border\);/s
    )
  })

  it('report export should not keep page-local teacher token duplication or bright hardcoded surfaces', () => {
    expect(reportExportSource).not.toContain('--journal-ink: var(--color-text-primary);')
    expect(reportExportSource).not.toContain('#ffffff')
    expect(reportExportSource).not.toContain('#f8fafc')
    expect(reportExportSource).not.toContain('rgba(255, 255, 255')
    expect(reportExportSource).not.toContain('.report-card--hero')
    expect(reportExportSource).not.toContain('.report-card--action')
    expect(reportExportSource).not.toContain('.report-card--metric')
    expect(reportExportSource).not.toContain('.report-btn {')
  })
})
