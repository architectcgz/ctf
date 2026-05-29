import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import adminContestFormDialogSource from '@/features/platform-contests/ui/PlatformContestFormDialog.vue?raw'
import contestChallengeEditorDialogSource from '@/features/contest-workbench/ui/ContestChallengeEditorDialog.vue?raw'
import adminUserFormDialogSource from '@/features/platform-user-management/ui/PlatformUserFormDialog.vue?raw'
import awdAttackLogDetailsSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogDetailsSection.vue?raw'
import awdAttackLogDialogSourceBase from '@/features/contest-awd-admin/ui/AWDAttackLogDialog.vue?raw'
import awdAttackLogTargetSectionSource from '@/features/contest-awd-admin/ui/AWDAttackLogTargetSection.vue?raw'
import awdOperationsDialogFooterSource from '@/features/contest-awd-admin/ui/AWDOperationsDialogFooter.vue?raw'
import awdRoundCreateDialogSourceBase from '@/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue?raw'
import awdRoundCreateScoreSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateScoreSection.vue?raw'
import awdRoundCreateSettingsSectionSource from '@/features/contest-awd-admin/ui/AWDRoundCreateSettingsSection.vue?raw'
import awdServiceCheckDialogSourceBase from '@/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue?raw'
import awdServiceCheckResultSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckResultSection.vue?raw'
import awdServiceCheckTargetSectionSource from '@/features/contest-awd-admin/ui/AWDServiceCheckTargetSection.vue?raw'
import awdReadinessOverrideDialogSource from '@/features/awd-readiness/ui/AWDReadinessOverrideDialog.vue?raw'
import adminNotificationPublishDrawerSource from '@/features/admin-notification-publisher/ui/AdminNotificationPublishDrawer.vue?raw'
import awdReviewTeamDrawerSource from '@/components/teacher/awd-review/AwdReviewTeamDrawer.vue?raw'
import imageCreateModalSource from '@/components/platform/images/ImageCreateModal.vue?raw'
import imageDetailModalSource from '@/components/platform/images/ImageDetailModal.vue?raw'

const awdAttackLogDialogSource = [
  awdAttackLogDialogSourceBase,
  awdAttackLogTargetSectionSource,
  awdAttackLogDetailsSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')

const awdRoundCreateDialogSource = [
  awdRoundCreateDialogSourceBase,
  awdRoundCreateSettingsSectionSource,
  awdRoundCreateScoreSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')

const awdServiceCheckDialogSource = [
  awdServiceCheckDialogSourceBase,
  awdServiceCheckTargetSectionSource,
  awdServiceCheckResultSectionSource,
  awdOperationsDialogFooterSource,
  readFileSync(resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsDialogs.css'), 'utf8'),
].join('\n')

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
    const drawerSources = [adminNotificationPublishDrawerSource, awdReviewTeamDrawerSource]

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
