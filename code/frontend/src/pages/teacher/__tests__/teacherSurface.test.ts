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
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'

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

const teacherManagementSources = [
  ['ClassManagementPage.vue', classManagementSource],
  ['StudentManagementPage.vue', studentManagementSource],
  ['TeacherInstanceManagementPage.vue', instanceManagementSource],
  ['AwdReviewSurfaceShell.vue', awdReviewSurfaceShellSource],
] as const

describe('teacher surface source regression', () => {
  it.each(teacherManagementSources)(
    '%s 应通过共享 teacher-management-shell 承接教师端 surface token',
    (_name, source) => {
      expect(source).toContain('teacher-management-shell')
      expect(source).toContain('workspace-shell')
      expect(source).not.toContain('--journal-ink: var(--color-text-primary);')
    }
  )

  it('AWD 复盘 widgets 应通过共享 surface shell 承接教师端外层壳', () => {
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewSurfaceShell')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewSurfaceShell')
  })
})
