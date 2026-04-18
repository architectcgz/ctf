import { describe, expect, it } from 'vitest'

import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'
import instanceManagementSource from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue?raw'
import awdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'
import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'

describe('teacher root shell cleanup', () => {
  it.each([
    ['ClassManagementPage.vue', classManagementSource],
    ['StudentManagementPage.vue', studentManagementSource],
    ['TeacherInstanceManagementPage.vue', instanceManagementSource],
    ['TeacherAWDReviewIndex.vue', awdReviewIndexSource],
    ['TeacherAWDReviewDetail.vue', awdReviewDetailSource],
  ])('%s 应切到共享 workspace 根壳，不再手写教师页外层圆角', (_name, source) => {
    expect(source).toContain('workspace-shell')
    expect(source).toContain('teacher-management-shell')
    expect(source).not.toContain('rounded-[30px]')
  })
})
