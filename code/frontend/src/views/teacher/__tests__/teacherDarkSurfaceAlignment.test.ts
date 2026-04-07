import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import reportExportSource from '@/views/teacher/ReportExport.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

const teacherDirectoryPattern =
  /teacher-directory-head[\s\S]*teacher-directory-row[\s\S]*(teacher-directory-row-main|teacher-directory-cell)[\s\S]*teacher-directory-row-tags/s

describe('teacher dark surface alignment', () => {
  it('teacher management pages should use shared teacher surface classes', () => {
    expect(classManagementSource).toContain('teacher-surface')
    expect(studentManagementSource).toContain('teacher-surface')
    expect(instanceManagementSource).toContain('teacher-surface')
    expect(reportExportSource).toContain('teacher-surface')
  })

  it('target pages should reuse shared journal and directory surface vocabulary instead of page-local skins', () => {
    expect(classManagementSource).toContain('journal-eyebrow')
    expect(classManagementSource).toContain('teacher-summary-item')
    expect(classManagementSource).toContain('teacher-directory-head')
    expect(classManagementSource).toContain('teacher-directory-row')
    expect(studentManagementSource).toContain('journal-eyebrow')
    expect(studentManagementSource).toContain('teacher-controls')
    expect(studentManagementSource).toContain('teacher-directory-head')
    expect(studentManagementSource).toContain('teacher-directory-row')
    expect(instanceManagementSource).toContain('journal-eyebrow')
    expect(instanceManagementSource).toContain('teacher-controls')
    expect(instanceManagementSource).toContain('teacher-directory-head')
    expect(instanceManagementSource).toContain('teacher-directory-row')
    expect(reportExportSource).toContain('journal-eyebrow')
    expect(reportExportSource).toContain('journal-brief')
    expect(reportExportSource).toContain('journal-metric')
    expect(reportExportSource).not.toContain('report-kpi-card--success')
    expect(reportExportSource).not.toContain('report-kpi-card--warning')
    expect(reportExportSource).not.toContain('report-kpi-card--primary')
  })

  it('class management should not leak element-plus primary plain button chrome', () => {
    expect(classManagementSource).not.toContain('<ElButton type="primary" plain')
    expect(classManagementSource).toContain('class="teacher-btn teacher-btn--primary"')
    expect(classManagementSource).toContain('class="teacher-btn teacher-btn--ghost"')
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

  it('teacher management list pages should render shared directory shells for rows and empty states', () => {
    expect(classManagementSource).toMatch(teacherDirectoryPattern)
    expect(studentManagementSource).toMatch(teacherDirectoryPattern)
    expect(instanceManagementSource).toMatch(teacherDirectoryPattern)
    expect(classManagementSource).toContain('teacher-empty-state')
    expect(studentManagementSource).toContain('teacher-empty-state')
    expect(instanceManagementSource).toContain('teacher-empty-state')
  })

  it('student and instance pages should not keep darker or louder local skins than dashboard', () => {
    expect(studentManagementSource).not.toContain('teacher-kpi-card--primary')
    expect(studentManagementSource).not.toContain('teacher-kpi-card--success')
    expect(studentManagementSource).not.toContain('teacher-kpi-card--warning')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-hero')
    expect(instanceManagementSource).toContain('--teacher-management-hero-border: var(--teacher-card-border);')
    expect(instanceManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
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
