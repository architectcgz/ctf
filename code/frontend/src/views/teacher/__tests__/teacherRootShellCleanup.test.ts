import { describe, expect, it } from 'vitest'

import classManagementSource from '@/features/teacher-class-management/ui/ClassManagementPage.vue?raw'
import studentManagementSource from '@/features/teacher-student-management/ui/StudentManagementPage.vue?raw'
import instanceManagementSourceBase from '@/features/teacher-instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/components/teacher/instance-management/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/components/teacher/instance-management/TeacherInstanceHeroPanel.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'

const instanceManagementSource = [
  instanceManagementSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

describe('teacher root shell cleanup', () => {
  it.each([
    ['ClassManagementPage.vue', classManagementSource],
    ['StudentManagementPage.vue', studentManagementSource],
    ['TeacherInstanceManagementPage.vue', instanceManagementSource],
    ['AwdReviewSurfaceShell.vue', awdReviewSurfaceShellSource],
    ['AwdReviewIndexWorkspace.vue', awdReviewIndexWorkspaceSource],
    ['AwdReviewWorkspace.vue', awdReviewWorkspaceSource],
  ])('%s 应切到共享 workspace 根壳，不再手写教师页外层圆角', (_name, source) => {
    if (source === awdReviewSurfaceShellSource) {
      expect(source).toContain('workspace-shell')
      expect(source).toContain('teacher-management-shell')
    }
    expect(source).not.toContain('rounded-[30px]')
  })
})
