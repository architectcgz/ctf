<template>
  <UserProfileWorkspaceShell
    :loading="loading"
    :error="error"
    :profile="profile"
    :export-loading="exportLoading"
    :export-error="exportError"
    :report-format="reportFormat"
    :latest-report="latestReport"
    :latest-report-format="latestReportFormat"
    :latest-report-created-at="latestReportCreatedAt"
    :can-manage-personal-report="canManagePersonalReport"
    :page-copy="pageCopy"
    :profile-fields="profileFields"
    :report-task-meta="reportTaskMeta"
    :profile-summary-items="profileSummaryItems"
    @load-profile="loadProfile"
    @create-report="createReport"
    @download-report="handleDownload"
    @update-report-format="reportFormat = $event"
  />
</template>

<script setup lang="ts">
import UserProfileWorkspaceShell from '@/components/profile/UserProfileWorkspaceShell.vue'
import { useUserProfilePage } from '@/features/profile'

const {
  loading,
  error,
  profile,
  exportLoading,
  exportError,
  reportFormat,
  latestReport,
  latestReportFormat,
  latestReportCreatedAt,
  canManagePersonalReport,
  pageCopy,
  profileFields,
  reportTaskMeta,
  profileSummaryItems,
  loadProfile,
  createReport,
  handleDownload,
} = useUserProfilePage()
</script>
