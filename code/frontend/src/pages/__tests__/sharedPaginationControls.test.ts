import { describe, expect, it } from 'vitest'

import challengeDirectoryPanelSource from '@/features/challenge-list/ui/ChallengeDirectoryPanel.vue?raw'
import notificationListSource from '@/pages/notifications/NotificationListRoutePage.vue?raw'
import classManagementSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import studentManagementSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'

describe('shared pagination controls usage', () => {
  it('学生与教师目录页应接入共享分页组件，而不是继续本地拼分页按钮结构', () => {
    for (const source of [
      challengeDirectoryPanelSource,
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
