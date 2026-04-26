import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import awdReviewIndexSource from '@/views/teacher/TeacherAWDReviewIndex.vue?raw'
import awdReviewDetailSource from '@/views/teacher/TeacherAWDReviewDetail.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

const teacherDirectoryPattern =
  /WorkspaceDataTable[\s\S]*workspace-directory-list[\s\S]*teacher-directory-cell-name[\s\S]*teacher-directory-row-cta/s
const teacherClassDirectoryPattern =
  /WorkspaceDataTable[\s\S]*workspace-directory-list[\s\S]*teacher-directory-cell-class-name[\s\S]*teacher-directory-state-chip/s
const teacherInstanceDataTablePattern =
  /WorkspaceDataTable[\s\S]*workspace-directory-list[\s\S]*teacher-instance-user-cell[\s\S]*instance-status-pill/s

describe('teacher dark surface alignment', () => {
  it('teacher management pages should use shared teacher surface classes', () => {
    expect(classManagementSource).toContain('teacher-surface')
    expect(studentManagementSource).toContain('teacher-surface')
    expect(instanceManagementSource).toContain('teacher-surface')
    expect(awdReviewIndexSource).toContain('teacher-surface')
    expect(awdReviewDetailSource).toContain('teacher-surface')
  })

  it('target pages should reuse shared journal and directory surface vocabulary instead of page-local skins', () => {
    expect(classManagementSource).toContain('workspace-overline')
    expect(classManagementSource).toContain('metric-panel-card')
    expect(classManagementSource).not.toContain('teacher-summary-item')
    expect(classManagementSource).toContain('WorkspaceDataTable')
    expect(classManagementSource).toContain('teacher-directory-row-cta')
    expect(studentManagementSource).toContain('workspace-overline')
    expect(studentManagementSource).toContain('teacher-actions')
    expect(studentManagementSource).toContain('metric-panel-card')
    expect(studentManagementSource).not.toContain('teacher-summary-item')
    expect(studentManagementSource).toContain('WorkspaceDataTable')
    expect(studentManagementSource).toContain('teacher-directory-row-cta')
    expect(instanceManagementSource).toContain('workspace-overline')
    expect(instanceManagementSource).toContain('teacher-actions')
    expect(instanceManagementSource).toContain('metric-panel-card')
    expect(instanceManagementSource).not.toContain('teacher-summary-item')
    expect(instanceManagementSource).toContain('WorkspaceDataTable')
    expect(instanceManagementSource).toContain('teacher-directory-row')
    expect(awdReviewIndexSource).toContain('workspace-overline')
    expect(awdReviewIndexSource).toContain('teacher-actions')
    expect(awdReviewIndexSource).toContain('metric-panel-card')
    expect(awdReviewIndexSource).not.toContain('teacher-summary-item')
    expect(awdReviewDetailSource).toContain('workspace-overline')
    expect(awdReviewDetailSource).toContain('teacher-actions')
    expect(awdReviewDetailSource).toContain('metric-panel-card')
    expect(awdReviewDetailSource).not.toContain('teacher-summary-item')
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
    expect(classManagementSource).toMatch(teacherClassDirectoryPattern)
    expect(studentManagementSource).toMatch(teacherDirectoryPattern)
    expect(instanceManagementSource).toMatch(teacherInstanceDataTablePattern)
    expect(classManagementSource).toContain('teacher-empty-state')
    expect(studentManagementSource).toContain('teacher-empty-state')
    expect(instanceManagementSource).toContain('teacher-empty-state')
  })

  it('student and instance pages should not keep darker or louder local skins than dashboard', () => {
    expect(studentManagementSource).not.toContain('teacher-kpi-card--primary')
    expect(studentManagementSource).not.toContain('teacher-kpi-card--success')
    expect(studentManagementSource).not.toContain('teacher-kpi-card--warning')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-hero')
    expect(instanceManagementSource).toContain(
      '--teacher-management-hero-border: var(--teacher-card-border);'
    )
    expect(instanceManagementSource).not.toMatch(/^\.teacher-hero\s*\{/m)
  })

  it('awd review pages should not keep page-local teacher token duplication or bright hardcoded surfaces', () => {
    expect(awdReviewIndexSource).not.toContain('--journal-ink: var(--color-text-primary);')
    expect(awdReviewIndexSource).not.toContain('#ffffff')
    expect(awdReviewIndexSource).not.toContain('#f8fafc')
    expect(awdReviewIndexSource).not.toContain('rgba(255, 255, 255')
    expect(awdReviewIndexSource).not.toContain('.teacher-btn {')

    expect(awdReviewDetailSource).not.toContain('--journal-ink: var(--color-text-primary);')
    expect(awdReviewDetailSource).not.toContain('#ffffff')
    expect(awdReviewDetailSource).not.toContain('#f8fafc')
    expect(awdReviewDetailSource).not.toContain('rgba(255, 255, 255')
    expect(awdReviewDetailSource).not.toContain('.teacher-btn {')
  })
})
