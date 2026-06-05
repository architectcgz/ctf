import { describe, expect, it } from 'vitest'

import classManagementSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import studentManagementSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'
import instanceManagementSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'
import teacherClassManagementHeaderActionsSource from '@/features/teacher/class-management/ui/TeacherClassManagementHeaderActions.vue?raw'

const instanceManagementSource = [
  instanceManagementSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

describe('teacher dark surface alignment', () => {
  it('teacher management pages should use shared teacher surface classes', () => {
    expect(classManagementSource).toContain('teacher-surface')
    expect(studentManagementSource).toContain('teacher-surface')
    expect(instanceManagementSource).toContain('teacher-surface')
    expect(awdReviewSurfaceShellSource).toContain('teacher-surface')
  })

  it('class management should not leak element-plus primary plain button chrome', () => {
    expect(classManagementSource).not.toContain('<ElButton type="primary" plain')
    expect(teacherClassManagementHeaderActionsSource).toContain(
      'class="header-btn header-btn--primary"'
    )
    expect(teacherClassManagementHeaderActionsSource).toContain(
      'class="header-btn header-btn--ghost"'
    )
  })
})
