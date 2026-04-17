import { describe, expect, it } from 'vitest'

import adminContestFormDialogSource from '@/components/admin/contest/AdminContestFormDialog.vue?raw'
import contestChallengeEditorDialogSource from '@/components/admin/contest/ContestChallengeEditorDialog.vue?raw'
import adminUserFormDialogSource from '@/components/admin/user/AdminUserFormDialog.vue?raw'
import awdRoundCreateDialogSource from '@/components/admin/contest/AWDRoundCreateDialog.vue?raw'
import awdServiceCheckDialogSource from '@/components/admin/contest/AWDServiceCheckDialog.vue?raw'
import awdAttackLogDialogSource from '@/components/admin/contest/AWDAttackLogDialog.vue?raw'
import awdReadinessOverrideDialogSource from '@/components/admin/contest/AWDReadinessOverrideDialog.vue?raw'
import adminNotificationPublishDrawerSource from '@/components/notifications/AdminNotificationPublishDrawer.vue?raw'
import teacherAwdReviewTeamDrawerSource from '@/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue?raw'
import imageManageSource from '@/views/admin/ImageManage.vue?raw'

describe('backoffice dialog adoption', () => {
  it('centered backoffice dialogs should adopt AdminSurfaceModal', () => {
    const centeredSources = [
      adminContestFormDialogSource,
      contestChallengeEditorDialogSource,
      adminUserFormDialogSource,
      awdRoundCreateDialogSource,
      awdServiceCheckDialogSource,
      awdAttackLogDialogSource,
      awdReadinessOverrideDialogSource,
      imageManageSource,
    ]

    centeredSources.forEach((source) => {
      expect(source).toContain("from '@/components/common/modal-templates/AdminSurfaceModal.vue'")
      expect(source).toContain('<AdminSurfaceModal')
    })
  })

  it('drawer-style backoffice panels should adopt AdminSurfaceDrawer', () => {
    const drawerSources = [adminNotificationPublishDrawerSource, teacherAwdReviewTeamDrawerSource]

    drawerSources.forEach((source) => {
      expect(source).toContain("from '@/components/common/modal-templates/AdminSurfaceDrawer.vue'")
      expect(source).toContain('<AdminSurfaceDrawer')
    })
  })

  it('admin user form dialog should adopt shared form and action primitives', () => {
    expect(adminUserFormDialogSource).toContain('<AdminSurfaceModal')
    expect(adminUserFormDialogSource).toContain('class="ui-field')
    expect(adminUserFormDialogSource).toContain('class="ui-control-wrap')
    expect(adminUserFormDialogSource).toContain('class="ui-control')
    expect(adminUserFormDialogSource).toContain('class="ui-btn ui-btn--secondary')
    expect(adminUserFormDialogSource).toContain('class="ui-btn ui-btn--primary')
    expect(adminUserFormDialogSource).not.toContain('rounded-xl border border-border bg-surface')
  })
})
