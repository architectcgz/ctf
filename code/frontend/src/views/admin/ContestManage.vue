<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import AdminContestFormDialog from '@/components/admin/contest/AdminContestFormDialog.vue'
import AWDOperationsPanel from '@/components/admin/contest/AWDOperationsPanel.vue'
import ContestOrchestrationPage from '@/components/admin/contest/ContestOrchestrationPage.vue'
import { useAdminContests } from '@/composables/useAdminContests'

const AWD_SELECTED_CONTEST_STORAGE_KEY = 'ctf_admin_awd_selected_contest'

function loadStoredSelectedAwdContestId(): string | null {
  if (typeof window === 'undefined') {
    return null
  }
  const value = window.sessionStorage.getItem(AWD_SELECTED_CONTEST_STORAGE_KEY)
  return value?.trim() || null
}

function persistSelectedAwdContestId(value: string | null): void {
  if (typeof window === 'undefined') {
    return
  }
  if (value) {
    window.sessionStorage.setItem(AWD_SELECTED_CONTEST_STORAGE_KEY, value)
    return
  }
  window.sessionStorage.removeItem(AWD_SELECTED_CONTEST_STORAGE_KEY)
}

const {
  list,
  total,
  page,
  pageSize,
  loading,
  refresh,
  changePage,
  statusFilter,
  dialogOpen,
  mode,
  saving,
  formDraft,
  fieldLocks,
  statusOptions,
  openCreateDialog,
  openEditDialog,
  closeDialog,
  saveContest,
} = useAdminContests()

const selectedAwdContestId = ref<string | null>(loadStoredSelectedAwdContestId())
const awdContests = computed(() => list.value.filter((item) => item.mode === 'awd'))

onMounted(() => {
  void refresh()
})

watch(
  awdContests,
  (nextContests) => {
    if (nextContests.length === 0) {
      return
    }

    const stillExists = nextContests.some((item) => item.id === selectedAwdContestId.value)
    if (!stillExists) {
      selectedAwdContestId.value = nextContests[0].id
    }
  },
  { immediate: true }
)

watch(
  () => selectedAwdContestId.value,
  (value) => {
    persistSelectedAwdContestId(value)
  },
  { immediate: true }
)

function updateStatusFilter(value: typeof statusFilter.value) {
  statusFilter.value = value
}

function updateSelectedAwdContestId(value: string) {
  selectedAwdContestId.value = value
}

function handleDialogOpenChange(value: boolean) {
  if (!value) {
    closeDialog()
  }
}
</script>

<template>
  <div class="space-y-6">
    <ContestOrchestrationPage
      :list="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :status-filter="statusFilter"
      @refresh="refresh"
      @open-create-dialog="openCreateDialog"
      @update-status-filter="updateStatusFilter"
      @open-edit-dialog="openEditDialog"
      @change-page="changePage"
    />

    <AWDOperationsPanel
      :contests="awdContests"
      :selected-contest-id="selectedAwdContestId"
      @update:selected-contest-id="updateSelectedAwdContestId"
    />

    <AdminContestFormDialog
      :open="dialogOpen"
      :mode="mode"
      :draft="formDraft"
      :saving="saving"
      :status-options="statusOptions"
      :field-locks="fieldLocks"
      @update:open="handleDialogOpenChange"
      @save="saveContest"
    />
  </div>
</template>
