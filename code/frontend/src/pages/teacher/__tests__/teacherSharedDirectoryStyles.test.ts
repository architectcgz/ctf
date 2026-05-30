import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import classStudentsSourceBase from '@/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue?raw'
import studentManagementSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'
import instanceManagementSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)
const classStudentsSource = [
  classStudentsSourceBase,
  classStudentsOverviewPanelSource,
  classStudentsInsightWindowPanelSource,
  classStudentsDirectoryPanelSource,
].join('\n')

const instanceManagementSource = [
  instanceManagementSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

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
