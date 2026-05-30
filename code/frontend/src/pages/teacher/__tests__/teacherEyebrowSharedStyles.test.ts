import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import teacherClassInsightsSource from '@/components/teacher/ClassInsightsPanel.vue?raw'
import teacherClassReviewSource from '@/components/teacher/ClassReviewPanel.vue?raw'
import teacherClassTrendSource from '@/components/teacher/ClassTrendPanel.vue?raw'
import interventionPanelSource from '@/features/teaching/student-analysis-review/ui/InterventionPanel.vue?raw'
import classManagementSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import classStudentsSourceBase from '@/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue?raw'
import teacherInstanceManagementSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'
import studentManagementSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewDetailSource from '@/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue?raw'

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

const teacherInstanceManagementSource = [
  teacherInstanceManagementSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

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
      interventionPanelSource,
      classManagementSource,
      classStudentsSource,
      teacherInstanceManagementSource,
      studentManagementSource,
      awdReviewIndexWorkspaceSource,
      awdReviewDetailSource,
    ]) {
      expect(source).not.toMatch(/^\.journal-eyebrow\s*\{/m)
    }
  })

  it('班级工作区的 trend review insight action panel 不应继续渲染 eyebrow 结构', () => {
    for (const source of [
      teacherClassInsightsSource,
      teacherClassReviewSource,
      teacherClassTrendSource,
      interventionPanelSource,
    ]) {
      expect(source).not.toContain('class="journal-eyebrow"')
    }
  })
})
