import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import teacherClassInsightsSource from '@/components/teacher/ClassInsightsPanel.vue?raw'
import teacherClassReviewSource from '@/components/teacher/ClassReviewPanel.vue?raw'
import teacherClassTrendSource from '@/components/teacher/ClassTrendPanel.vue?raw'
import interventionPanelSource from '@/components/teacher/InterventionPanel.vue?raw'
import classManagementSource from '@/features/teacher-class-management/ui/ClassManagementPage.vue?raw'
import classStudentsSourceBase from '@/features/class-students-workspace/ui/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/components/teacher/class-management/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/components/teacher/class-management/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/components/teacher/class-management/ClassStudentsDirectoryPanel.vue?raw'
import teacherInstanceManagementSourceBase from '@/features/teacher-instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/components/teacher/instance-management/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/components/teacher/instance-management/TeacherInstanceHeroPanel.vue?raw'
import studentManagementSource from '@/features/teacher-student-management/ui/StudentManagementPage.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewDetailSource from '@/views/teacher/TeacherAWDReviewDetail.vue?raw'

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
