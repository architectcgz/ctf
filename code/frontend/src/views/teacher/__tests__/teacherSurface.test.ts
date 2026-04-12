import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import classStudentsSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import studentAnalysisSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import awdReviewIndexSource from '@/views/teacher/TeacherAWDReviewIndex.vue?raw'
import awdReviewDetailSource from '@/views/teacher/TeacherAWDReviewDetail.vue?raw'
import reviewArchiveSource from '@/views/teacher/TeacherStudentReviewArchive.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

const teacherSurfacePattern =
  /--journal-ink:\s*var\(--color-text-primary\);[\s\S]*--journal-surface:\s*color-mix\(in srgb, var\(--color-bg-surface\) 88%, var\(--color-bg-base\)\);/s

const forbiddenTeacherSurfaceLiterals = ['rgba(255, 255, 255, 0.98)', '#ffffff', '#f8fafc']

const teacherSurfaceSources = [
  ['teacher-surface.css', teacherSurfaceSource],
  ['ClassStudentsPage.vue', classStudentsSource],
  ['StudentAnalysisPage.vue', studentAnalysisSource],
  ['TeacherStudentReviewArchive.vue', reviewArchiveSource],
] as const

const teacherManagementSources = [
  ['ClassManagementPage.vue', classManagementSource],
  ['StudentManagementPage.vue', studentManagementSource],
  ['TeacherInstanceManagementPage.vue', instanceManagementSource],
  ['TeacherAWDReviewIndex.vue', awdReviewIndexSource],
  ['TeacherAWDReviewDetail.vue', awdReviewDetailSource],
] as const

const teacherSurfaceForbiddenLiteralCases = teacherSurfaceSources.flatMap(([sourceName, source]) =>
  forbiddenTeacherSurfaceLiterals.map(
    (forbiddenLiteral) => [sourceName, forbiddenLiteral, source] as const
  )
)

describe('teacher surface source regression', () => {
  it.each(teacherSurfaceSources)('%s 应命中教师端 surface 主题模式', (_name, source) => {
    expect(teacherSurfacePattern.test(source)).toBe(true)
  })

  it.each(teacherManagementSources)(
    '%s 应通过共享 teacher-management-shell 承接教师端 surface token',
    (_name, source) => {
      expect(source).toContain('teacher-management-shell')
      expect(source).not.toContain('--journal-ink: var(--color-text-primary);')
    }
  )

  it.each(teacherSurfaceForbiddenLiteralCases)(
    '%s 不应包含教师端高对比亮色 surface 硬编码: %s',
    (_name, forbiddenLiteral, source) => {
      expect(source.includes(forbiddenLiteral)).toBe(false)
    }
  )
})
