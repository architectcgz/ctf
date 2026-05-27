<script setup lang="ts">
import type { AdminAwdChallengeData, AdminAwdChallengeImportPreview } from '@/api/contracts'
import AwdChallengeLibrarySection from '@/components/platform/awd-service/AwdChallengeLibrarySection.vue'
import AwdChallengeWorkspaceHeader from '@/components/platform/awd-service/AwdChallengeWorkspaceHeader.vue'
import AwdChallengeImportSection from './AwdChallengeImportSection.vue'
import type { PlatformAwdChallengeImportUploadResult } from '../model'

type AwdServiceTypeFilter = AdminAwdChallengeData['service_type'] | ''
type AwdServiceStatusFilter = AdminAwdChallengeData['status'] | ''

const props = withDefaults(
  defineProps<{
    mode?: 'library' | 'import'
    list: AdminAwdChallengeData[]
    total: number
    page: number
    pageSize: number
    loading: boolean
    keyword: string
    serviceTypeFilter: AwdServiceTypeFilter
    statusFilter: AwdServiceStatusFilter
    uploading: boolean
    queueLoading: boolean
    importQueue: AdminAwdChallengeImportPreview[]
    uploadResults: PlatformAwdChallengeImportUploadResult[]
    selectedFileName?: string
  }>(),
  {
    mode: 'library',
  }
)

const emit = defineEmits<{
  refresh: []
  refreshImportQueue: []
  updateKeyword: [value: string]
  updateServiceTypeFilter: [value: AwdServiceTypeFilter]
  updateStatusFilter: [value: AwdServiceStatusFilter]
  selectImportPackages: [files: File[]]
  commitImport: [preview: AdminAwdChallengeImportPreview]
  openImportPage: []
  openEditDialog: [challenge: AdminAwdChallengeData]
  deleteChallenge: [challenge: AdminAwdChallengeData]
  changePage: [page: number]
}>()
</script>

<template>
  <div
    class="workspace-shell journal-shell journal-shell-admin journal-hero awd-challenge-library-shell"
  >
    <main class="content-pane awd-challenge-library-content">
      <AwdChallengeWorkspaceHeader
        :mode="mode"
        @refresh="emit('refresh')"
        @open-import-page="emit('openImportPage')"
        @refresh-import-queue="emit('refreshImportQueue')"
      />

      <AwdChallengeLibrarySection
        v-if="mode === 'library'"
        :list="list"
        :total="total"
        :page="page"
        :page-size="pageSize"
        :loading="loading"
        :keyword="keyword"
        :service-type-filter="serviceTypeFilter"
        :status-filter="statusFilter"
        @update-keyword="emit('updateKeyword', $event)"
        @update-service-type-filter="emit('updateServiceTypeFilter', $event)"
        @update-status-filter="emit('updateStatusFilter', $event)"
        @open-edit-dialog="emit('openEditDialog', $event)"
        @delete-challenge="emit('deleteChallenge', $event)"
        @change-page="emit('changePage', $event)"
      />

      <AwdChallengeImportSection
        v-else
        :uploading="uploading"
        :queue-loading="queueLoading"
        :import-queue="importQueue"
        :upload-results="uploadResults"
        :selected-file-name="selectedFileName"
        @select-import-packages="emit('selectImportPackages', $event)"
        @commit-import="emit('commitImport', $event)"
      />
    </main>
  </div>
</template>

<style scoped>
.awd-challenge-library-content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}
</style>
