import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewIndexWorkspace.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/teacher-awd-review/TeacherAWDReviewSurfaceShell.vue?raw'

describe('teacher root shell cleanup', () => {
  it.each([
    ['ClassManagementPage.vue', classManagementSource],
    ['StudentManagementPage.vue', studentManagementSource],
    ['TeacherInstanceManagementPage.vue', instanceManagementSource],
    ['TeacherAWDReviewSurfaceShell.vue', awdReviewSurfaceShellSource],
    ['TeacherAWDReviewIndexWorkspace.vue', awdReviewIndexWorkspaceSource],
    ['TeacherAWDReviewWorkspace.vue', awdReviewWorkspaceSource],
  ])('%s 应切到共享 workspace 根壳，不再手写教师页外层圆角', (_name, source) => {
    if (source === awdReviewSurfaceShellSource) {
      expect(source).toContain('workspace-shell')
      expect(source).toContain('teacher-management-shell')
    }
    expect(source).not.toContain('rounded-[30px]')
  })
})
