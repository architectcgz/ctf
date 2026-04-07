import { describe, expect, it } from 'vitest'

import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import classManagementSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'
import studentManagementSource from '@/components/teacher/student-management/StudentManagementPage.vue?raw'

describe('shared pagination controls usage', () => {
  it('学生与教师列表页应接入共享分页组件，而不是继续本地拼分页按钮结构', () => {
    for (const source of [
      challengeListSource,
      notificationListSource,
      classManagementSource,
      studentManagementSource,
    ]) {
      expect(source).toContain('PagePaginationControls')
      expect(source).not.toContain('challenge-pagination-actions')
      expect(source).not.toContain('notification-pagination-actions')
      expect(source).not.toContain('teacher-directory-pagination-actions')
      expect(source).not.toContain('teacher-directory-pagination-button')
    }
  })
})
