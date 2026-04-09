import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import classStudentsSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

function extractScopedStyle(source: string): string {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] ?? ''
}

describe('teacher shared directory styles', () => {
  it('班级学生页应通过 teacher-management-shell 承接目录与筛选共享样式', () => {
    expect(classStudentsSource).toMatch(/class="[^"]*\bteacher-management-shell\b[^"]*"/)
  })

  it('应在 teacher-surface.css 中声明教师端目录与筛选基础块共享样式', () => {
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-controls')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-controls-bar')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-controls-title')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-controls-copy')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-field-control')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-filter-control')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-directory-top')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-directory-title')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-directory-meta')
    expect(teacherSurfaceSource).toContain('.teacher-management-shell .teacher-directory-head')
  })

  it('教师端目录页不应继续在 scoped style 中重复声明这些共享基础块', () => {
    for (const source of [
      classManagementSource,
      studentManagementSource,
      classStudentsSource,
      instanceManagementSource,
    ]) {
      const style = extractScopedStyle(source)

      expect(style).not.toMatch(/^\.teacher-summary-grid\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-controls\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-controls-bar\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-controls-title\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-controls-copy\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-field\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-field-label\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-field-control\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-filter-control\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-directory-top\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-directory-title\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-directory-meta\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-directory-head\s*\{/m)
      expect(style).not.toMatch(/^\.teacher-directory-head-cell\s*\{/m)
    }
  })
})
