import { describe, expect, it } from 'vitest'

import adminContestFormDialogSource from '@/components/platform/contest/PlatformContestFormDialog.vue?raw'
import contestChallengeEditorDialogSource from '@/components/platform/contest/ContestChallengeEditorDialog.vue?raw'
import adminUserFormDialogSource from '@/components/platform/user/PlatformUserFormDialog.vue?raw'
import awdRoundCreateDialogSource from '@/components/platform/contest/AWDRoundCreateDialog.vue?raw'
import awdServiceCheckDialogSource from '@/components/platform/contest/AWDServiceCheckDialog.vue?raw'
import awdAttackLogDialogSource from '@/components/platform/contest/AWDAttackLogDialog.vue?raw'
import awdReadinessOverrideDialogSource from '@/components/platform/contest/AWDReadinessOverrideDialog.vue?raw'
import adminNotificationPublishDrawerSource from '@/components/notifications/AdminNotificationPublishDrawer.vue?raw'
import teacherAwdReviewTeamDrawerSource from '@/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue?raw'
import imageCreateModalSource from '@/components/platform/images/ImageCreateModal.vue?raw'
import imageDetailModalSource from '@/components/platform/images/ImageDetailModal.vue?raw'

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
      imageCreateModalSource,
      imageDetailModalSource,
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

  it('admin notification drawer should adopt shared button primitives instead of drawer-local button classes', () => {
    expect(adminNotificationPublishDrawerSource).toContain('class="ui-btn ui-btn--sm ui-btn--secondary"')
    expect(adminNotificationPublishDrawerSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(adminNotificationPublishDrawerSource).toContain('class="ui-btn ui-btn--primary"')
    expect(adminNotificationPublishDrawerSource).not.toContain('publish-inline-btn')
    expect(adminNotificationPublishDrawerSource).not.toContain('publish-btn')
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
