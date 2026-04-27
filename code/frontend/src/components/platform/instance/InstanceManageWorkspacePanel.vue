<script setup lang="ts">
import { Trash2 } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'

interface InstanceManageTableRow {
  id: string
  challenge: string
  student_id: string
  user: string
  username: string
  class_name: string
  ip_address: string
  status: string
  status_label: string
  created_at: string
  actions: string
}

type InstanceStatusFilter = 'running' | 'creating' | 'expired' | 'failed' | 'inactive' | ''

defineProps<{
  loading: boolean
  hasInstances: boolean
  rows: InstanceManageTableRow[]
  keyword: string
  statusFilter: InstanceStatusFilter
  page: number
  totalPages: number
  total: number
  destroyingId: string
  error: string | null
}>()

const emit = defineEmits<{
  (event: 'update:keyword', value: string): void
  (event: 'change:status-filter', value: InstanceStatusFilter): void
  (event: 'reset-filters'): void
  (event: 'open-student', studentId: string, className: string): void
  (event: 'destroy-instance', id: string): void
  (event: 'change-page', page: number): void
}>()

const columns = [
  { key: 'id', label: '实例 ID', widthClass: 'w-[20%] min-w-[12rem]' },
  { key: 'challenge', label: '关联题目', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'user', label: '所属用户', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'class_name', label: '班级', widthClass: 'w-[12%] min-w-[8rem]' },
  { key: 'ip_address', label: '访问地址', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'status', label: '状态', widthClass: 'w-[10%] min-w-[6rem]', align: 'center' as const },
  { key: 'created_at', label: '创建时间', widthClass: 'w-[13%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[8rem]', align: 'right' as const },
]

function handleStatusFilterChange(event: Event): void {
  const target = event.target
  emit('change:status-filter', target instanceof HTMLSelectElement ? target.value as InstanceStatusFilter : '')
}
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

      <WorkspaceDirectoryToolbar
        :model-value="keyword"
        :total="total"
        selected-sort-label=""
        :sort-options="[]"
        search-placeholder="检索实例、题目、用户或访问地址..."
        total-suffix="个实例"
        filter-panel-title="实例筛选"
        reset-label="清空筛选"
        :reset-disabled="!keyword && !statusFilter"
        @update:model-value="emit('update:keyword', $event)"
        @reset-filters="emit('reset-filters')"
      >
        <template #filter-panel>
          <label class="admin-instance-manage-filter-field">
            <span class="workspace-overline">实例状态</span>
            <select
              :value="statusFilter"
              class="admin-input admin-instance-manage-filter-control"
              @change="handleStatusFilterChange"
            >
              <option value="">
                全部状态
              </option>
              <option value="running">
                运行中
              </option>
              <option value="creating">
                创建中
              </option>
              <option value="expired">
                已过期
              </option>
              <option value="failed">
                异常
              </option>
              <option value="inactive">
                其他状态
              </option>
            </select>
          </label>
        </template>
      </WorkspaceDirectoryToolbar>

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

        <AppEmpty
          v-else-if="rows.length === 0"
          class="workspace-directory-empty"
          icon="Search"
          title="没有匹配实例"
          description="调整搜索关键词或筛选条件后再试。"
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
            <div class="instance-user-cell">
              <button
                type="button"
                class="instance-user-link"
                @click="emit('open-student', (row as InstanceManageTableRow).student_id, (row as InstanceManageTableRow).class_name)"
              >
                {{ (row as InstanceManageTableRow).user }}
              </button>
            </div>
          </template>
          <template #cell-class_name="{ row }">
            <span class="instance-class-cell">
              {{ (row as InstanceManageTableRow).class_name }}
            </span>
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
                class="ui-btn ui-btn--danger ui-btn--xs"
                :disabled="destroyingId === (row as InstanceManageTableRow).id"
                @click="emit('destroy-instance', (row as InstanceManageTableRow).id)"
              >
                <Trash2 class="h-3 w-3 mr-1" />
                {{ destroyingId === (row as InstanceManageTableRow).id ? '销毁中' : '销毁' }}
              </button>
            </div>
          </template>
        </WorkspaceDataTable>

        <WorkspaceDirectoryPagination
          v-if="total > 0 && rows.length > 0"
          :page="page"
          :total-pages="totalPages"
          :total="total"
          total-label="个实例"
          @change-page="emit('change-page', $event)"
        />
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

.admin-instance-manage-filter-field {
  display: grid;
  gap: var(--space-2);
}

.admin-instance-manage-filter-control {
  min-height: 2.5rem;
  width: 100%;
}

.instance-user-cell {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.instance-user-link {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  border: 0;
  background: transparent;
  padding: 0;
  text-align: left;
  cursor: pointer;
  color: var(--color-primary);
  font-size: var(--font-size-14);
  font-weight: 700;
  line-height: 1.45;
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.18em;
  transition: all 150ms ease;
}

.instance-user-link:hover {
  color: var(--color-primary-hover);
}

.instance-class-cell {
  font-size: var(--font-size-13);
  line-height: 1.5;
  color: var(--color-text-muted);
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
