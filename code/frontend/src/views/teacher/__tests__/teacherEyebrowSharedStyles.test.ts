import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import teacherClassInsightsSource from '@/components/teacher/TeacherClassInsightsPanel.vue?raw'
import teacherClassReviewSource from '@/components/teacher/TeacherClassReviewPanel.vue?raw'
import teacherClassTrendSource from '@/components/teacher/TeacherClassTrendPanel.vue?raw'
import teacherInterventionSource from '@/components/teacher/TeacherInterventionPanel.vue?raw'
import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import classStudentsSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import teacherInstanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import awdReviewIndexSource from '@/views/teacher/TeacherAWDReviewIndex.vue?raw'
import awdReviewDetailSource from '@/views/teacher/TeacherAWDReviewDetail.vue?raw'

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

describe('teacher eyebrow shared styles', () => {
  it('应该在 teacher surface 共享样式里声明 teacher 页面和 panel 的 eyebrow 规则', () => {
    expect(teacherSurfaceSource).toContain('.teacher-surface .journal-eyebrow')
    expect(teacherSurfaceSource).toContain('.teacher-panel .journal-eyebrow')
    expect(teacherSurfaceSource).toContain('--teacher-eyebrow-spacing, 0.08em')
  })

  it('teacher 页和 panel 不应继续本地重写整套 eyebrow 样式', () => {
    for (const source of [
      teacherClassInsightsSource,
      teacherClassReviewSource,
      teacherClassTrendSource,
      teacherInterventionSource,
      classManagementSource,
      classStudentsSource,
      teacherInstanceManagementSource,
      studentManagementSource,
      awdReviewIndexSource,
      awdReviewDetailSource,
    ]) {
      expect(source).not.toMatch(/^\.journal-eyebrow\s*\{/m)
    }
  })
})
