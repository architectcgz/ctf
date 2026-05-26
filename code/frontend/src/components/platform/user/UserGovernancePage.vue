<script setup lang="ts">
import { computed, ref, useTemplateRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { AdminUserImportData, AdminUserListItem, UserStatus } from '@/api/contracts'

import type { UserRole } from '@/utils/constants'
import UserGovernanceDetailModal from './UserGovernanceDetailModal.vue'
import UserGovernanceImportPanel from './UserGovernanceImportPanel.vue'
import UserGovernanceOverviewPanel from './UserGovernanceOverviewPanel.vue'

type UserFilterRole = UserRole | 'all'
type UserFilterStatus = UserStatus | 'all'
type UserPanelKey = 'overview' | 'import'

const props = defineProps<{
  list: AdminUserListItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  keyword: string
  studentNo: string
  teacherNo: string
  roleFilter: UserFilterRole
  statusFilter: UserFilterStatus
  importResult: AdminUserImportData | null
}>()

const emit = defineEmits<{
  refresh: []
  updateKeyword: [value: string]
  updateStudentNo: [value: string]
  updateTeacherNo: [value: string]
  updateRoleFilter: [value: UserFilterRole]
  updateStatusFilter: [value: UserFilterStatus]
  openCreateDialog: []
  openEditDialog: [user: AdminUserListItem]
  deleteUser: [userId: string]
  changePage: [page: number]
  importFile: [file: File]
}>()

const route = useRoute()
const router = useRouter()
const importInput = useTemplateRef<HTMLInputElement>('importInput')
const selectedUserId = ref<string | null>(null)

const activePanel = computed<UserPanelKey>(() => {
  const rawPanel = route.query.panel
  const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
  if (panel === 'import') {
    return 'import'
  }
  return 'overview'
})

const selectedUser = computed(() =>
  props.list.find((user) => user.id === selectedUserId.value) ?? null
)

function openUserDetail(user: AdminUserListItem): void {
  selectedUserId.value = user.id
}

function closeUserDetail(): void {
  selectedUserId.value = null
}

function handleDetailEdit(user: AdminUserListItem): void {
  closeUserDetail()
  emit('openEditDialog', user)
}

function handleDetailDelete(user: AdminUserListItem): void {
  closeUserDetail()
  emit('deleteUser', user.id)
}

async function switchPanel(panel: UserPanelKey): Promise<void> {
  if (activePanel.value === panel) return

  const nextQuery = { ...route.query }
  if (panel === 'overview') {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panel
  }

  await router.replace({ name: 'UserManage', query: nextQuery })
}

function triggerImport(): void {
  importInput.value?.click()
}

function handleImportChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  emit('importFile', file)
  input.value = ''
}
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <section
        v-show="activePanel === 'overview'"
        id="user-panel-overview"
        :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
      >
        <UserGovernanceOverviewPanel
          :list="list"
          :total="total"
          :page="page"
          :page-size="pageSize"
          :loading="loading"
          :keyword="keyword"
          :student-no="studentNo"
          :teacher-no="teacherNo"
          :role-filter="roleFilter"
          :status-filter="statusFilter"
          :import-result="importResult"
          @refresh="emit('refresh')"
          @open-import="switchPanel('import')"
          @open-create-dialog="emit('openCreateDialog')"
          @update-keyword="emit('updateKeyword', $event)"
          @update-student-no="emit('updateStudentNo', $event)"
          @update-teacher-no="emit('updateTeacherNo', $event)"
          @update-role-filter="emit('updateRoleFilter', $event)"
          @update-status-filter="emit('updateStatusFilter', $event)"
          @open-edit-dialog="emit('openEditDialog', $event)"
          @open-user-detail="openUserDetail"
          @change-page="emit('changePage', $event)"
        />
      </section>

      <UserGovernanceDetailModal
        :user="selectedUser"
        @close="closeUserDetail"
        @edit="handleDetailEdit"
        @delete="handleDetailDelete"
      />

      <section
        v-show="activePanel === 'import'"
        id="user-panel-import"
        :aria-hidden="activePanel === 'import' ? 'false' : 'true'"
      >
        <UserGovernanceImportPanel
          :import-result="importResult"
          @return-overview="switchPanel('overview')"
          @trigger-import="triggerImport"
        />
      </section>
    </main>

    <input
      ref="importInput"
      type="file"
      accept=".csv,text/csv"
      class="hidden"
      @change="handleImportChange"
    >
  </section>
</template>

<style scoped>
.journal-shell {
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --user-table-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --user-row-divider: color-mix(in srgb, var(--journal-border) 58%, transparent);
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --journal-note-label-weight: 600;
  --journal-note-label-spacing: 0.15em;
  --journal-note-label-color: var(--journal-muted);
  --journal-shell-dark-accent: var(--color-primary-hover);
}

@media (max-width: 767px) {
  .journal-hero {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
