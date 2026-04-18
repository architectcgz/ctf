import { describe, expect, it } from 'vitest'

import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import contestListSource from '@/views/contests/ContestList.vue?raw'
import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'
import instanceListSource from '@/views/instances/InstanceList.vue?raw'
import notificationListSource from '@/views/notifications/NotificationList.vue?raw'
import notificationDetailSource from '@/views/notifications/NotificationDetail.vue?raw'
import userProfileSource from '@/views/profile/UserProfile.vue?raw'
import skillProfileSource from '@/views/profile/SkillProfile.vue?raw'
import securitySettingsSource from '@/views/profile/SecuritySettings.vue?raw'
import scoreboardViewSource from '@/views/scoreboard/ScoreboardView.vue?raw'
import studentDifficultySource from '@/components/dashboard/student/StudentDifficultyPage.vue?raw'
import studentCategoryProgressSource from '@/components/dashboard/student/StudentCategoryProgressPage.vue?raw'
import studentOverviewSource from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue?raw'
import studentRecommendationSource from '@/components/dashboard/student/StudentRecommendationPage.vue?raw'

describe('student root shell cleanup', () => {
  it.each([
    ['ChallengeList.vue', challengeListSource],
    ['ContestList.vue', contestListSource],
    ['ContestDetail.vue', contestDetailSource],
    ['InstanceList.vue', instanceListSource],
    ['NotificationList.vue', notificationListSource],
    ['NotificationDetail.vue', notificationDetailSource],
    ['UserProfile.vue', userProfileSource],
    ['SkillProfile.vue', skillProfileSource],
    ['SecuritySettings.vue', securitySettingsSource],
    ['ScoreboardView.vue', scoreboardViewSource],
  ])('%s 应切到共享学生根壳，不再手写外层圆角', (_name, source) => {
    expect(source).toContain('workspace-shell')
    expect(source).toContain('journal-shell-user')
    expect(source).not.toContain('rounded-[30px]')
  })

  it.each([
    ['StudentDifficultyPage.vue', studentDifficultySource],
    ['StudentCategoryProgressPage.vue', studentCategoryProgressSource],
    ['StudentOverviewStyleEditorial.vue', studentOverviewSource],
    ['StudentRecommendationPage.vue', studentRecommendationSource],
  ])('%s 独立展示时也应切到共享学生根壳', (_name, source) => {
    expect(source).toContain('workspace-shell')
    expect(source).toContain('journal-shell-user')
    expect(source).not.toContain('rounded-[30px]')
  })
})
