<script setup lang="ts">
import { Trash2 } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'

interface InstanceManageTableRow {
  id: string
  challenge: string
  user: string
  user_meta: string
  ip_address: string
  status: string
  status_label: string
  created_at: string
  actions: string
}

defineProps<{
  loading: boolean
  hasInstances: boolean
  rows: InstanceManageTableRow[]
  page: number
  totalPages: number
  total: number
  destroyingId: string
  error: string | null
}>()

const emit = defineEmits<{
  (event: 'destroy-instance', id: string): void
  (event: 'change-page', page: number): void
}>()

const columns = [
  { key: 'id', label: '实例 ID', widthClass: 'w-[20%] min-w-[12rem]' },
  { key: 'challenge', label: '关联题目', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'user', label: '所属用户', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'ip_address', label: '访问地址', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'status', label: '状态', widthClass: 'w-[10%] min-w-[6rem]', align: 'center' as const },
  { key: 'created_at', label: '创建时间', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[8rem]', align: 'right' as const },
]
</script>

<template>
  <div class="admin-instance-manage-shell__content">
    <section class="workspace-directory-section admin-instance-manage-directory">
      <header class="list-heading">
        <div>
          <div class="workspace-overline">
            Active Instances
          </div>
          <h2 class="list-heading__title">
            实时实例列表
          </h2>
        </div>
      </header>

      <div
        v-if="loading && !hasInstances"
        class="py-12 flex justify-center"
      >
        <AppLoading>同步实例状态...</AppLoading>
      </div>

      <template v-else>
        <AppEmpty
          v-if="!hasInstances"
          class="workspace-directory-empty"
          icon="Server"
          title="暂无运行中的实例"
          description="当前平台上没有任何用户开启题目环境。"
        />

        <WorkspaceDataTable
          v-else
          class="workspace-directory-list admin-instance-manage-table"
          :columns="columns"
          :rows="rows"
          row-key="id"
        >
          <template #cell-id="{ row }">
            <span class="font-mono text-xs">{{ (row as InstanceManageTableRow).id }}</span>
          </template>
          <template #cell-user="{ row }">
            <div class="flex flex-col items-start gap-1">
              <span>{{ (row as InstanceManageTableRow).user }}</span>
              <span class="font-mono text-[11px] text-[var(--journal-muted)]">
                {{ (row as InstanceManageTableRow).user_meta }}
              </span>
            </div>
          </template>
          <template #cell-status="{ row }">
            <span
              class="instance-status-pill"
              :class="(row as InstanceManageTableRow).status === 'running' ? 'instance-status-pill--running' : 'instance-status-pill--inactive'"
            >
              {{ (row as InstanceManageTableRow).status_label }}
            </span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex justify-end gap-2">
              <button
                type="button"
                class="ui-btn ui-btn--ghost ui-btn--xs"
                :disabled="destroyingId === (row as InstanceManageTableRow).id"
                @click="emit('destroy-instance', (row as InstanceManageTableRow).id)"
              >
                <Trash2 class="h-3 w-3 mr-1" />
                {{ destroyingId === (row as InstanceManageTableRow).id ? '销毁中' : '销毁' }}
              </button>
            </div>
          </template>
        </WorkspaceDataTable>

        <div class="workspace-directory-pagination">
          <WorkspaceDirectoryPagination
            :page="page"
            :total-pages="totalPages"
            :total="total"
            total-label="个实例"
            @change-page="emit('change-page', $event)"
          />
        </div>
      </template>
    </section>

    <div
      v-if="error"
      class="teacher-surface-error"
    >
      {{ error }}
    </div>
  </div>
</template>

<style scoped>
.admin-instance-manage-shell__content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap);
  margin-top: var(--space-10);
}

.instance-status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.4rem;
  padding: 0 0.5rem;
  border-radius: 999px;
  border: 1px solid transparent;
  font-size: var(--font-size-10);
  font-weight: 700;
  text-transform: uppercase;
}

.instance-status-pill--running {
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-success) 24%, transparent);
  color: color-mix(in srgb, var(--color-success) 82%, var(--color-text-primary));
}

.instance-status-pill--inactive {
  background: color-mix(in srgb, var(--color-text-muted) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  color: var(--color-text-secondary);
}
</style>
