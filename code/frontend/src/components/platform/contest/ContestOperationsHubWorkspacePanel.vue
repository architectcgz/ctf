<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

import type { ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

defineProps<{
  loading: boolean
  loadError: string
  operableContests: ContestDetailData[]
}>()

const emit = defineEmits<{
  (event: 'retry'): void
  (event: 'back'): void
  (event: 'enter-operations', contestId: string): void
}>()

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const contestTableColumns = [
  { key: 'title', label: '赛事名称', widthClass: 'w-[30%] min-w-[15rem]' },
  { key: 'status', label: '状态', widthClass: 'w-[12%] min-w-[7rem]', align: 'center' as const },
  { key: 'mode', label: '模式', widthClass: 'w-[12%] min-w-[7rem]', align: 'center' as const },
  { key: 'starts_at', label: '开始时间', widthClass: 'w-[18%] min-w-[11rem]' },
  { key: 'ends_at', label: '结束时间', widthClass: 'w-[18%] min-w-[11rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[10rem]', align: 'right' as const },
]
</script>

<template>
  <section
    v-if="loading"
    class="workspace-directory-section contest-ops-section"
  >
    <AppLoading>正在同步赛事运维目录...</AppLoading>
  </section>

  <AppEmpty
    v-else-if="loadError"
    class="workspace-directory-section contest-ops-section"
    title="赛事运维目录暂时不可用"
    :description="loadError"
    icon="AlertTriangle"
  >
    <template #action>
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
        @click="emit('retry')"
      >
        重试加载
      </button>
    </template>
  </AppEmpty>

  <AppEmpty
    v-else-if="operableContests.length === 0"
    class="workspace-directory-section contest-ops-section"
    title="当前还没有可进入运维台的 AWD 赛事"
    description="请先在竞赛目录中创建 AWD 赛事，或将赛事推进到可运维状态。"
    icon="Trophy"
  >
    <template #action>
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
        @click="emit('back')"
      >
        返回竞赛目录
      </button>
    </template>
  </AppEmpty>

  <section
    v-else
    class="workspace-directory-section contest-ops-section contest-ops-directory"
  >
    <header class="list-heading">
      <div>
        <div class="journal-note-label">
          Contest Ops Directory
        </div>
        <h2 class="list-heading__title">
          竞赛列表
        </h2>
      </div>
    </header>

    <WorkspaceDataTable
      class="workspace-directory-list contest-ops-table"
      :columns="contestTableColumns"
      :rows="operableContests"
      row-key="id"
    >
      <template #cell-title="{ row }">
        <div class="contest-ops-table__contest">
          <div
            class="contest-ops-table__title"
            :title="String((row as ContestDetailData).title)"
          >
            {{ (row as ContestDetailData).title }}
          </div>
          <div
            class="contest-ops-table__description"
            :title="(row as ContestDetailData).description || '当前未填写赛事描述。'"
          >
            {{ (row as ContestDetailData).description || '当前未填写赛事描述。' }}
          </div>
        </div>
      </template>

      <template #cell-status="{ row }">
        <span class="contest-ops-table__badge">
          {{ getStatusLabel((row as ContestDetailData).status) }}
        </span>
      </template>

      <template #cell-mode="{ row }">
        <span class="contest-ops-table__badge contest-ops-table__badge--muted">
          {{ getModeLabel((row as ContestDetailData).mode) }}
        </span>
      </template>

      <template #cell-starts_at="{ row }">
        <span class="contest-ops-table__time">
          {{ formatDateTime((row as ContestDetailData).starts_at) }}
        </span>
      </template>

      <template #cell-ends_at="{ row }">
        <span class="contest-ops-table__time">
          {{ formatDateTime((row as ContestDetailData).ends_at) }}
        </span>
      </template>

      <template #cell-actions="{ row }">
        <div class="contest-ops-actions">
          <button
            :id="`contest-ops-enter-${(row as ContestDetailData).id}`"
            type="button"
            class="ui-btn ui-btn--primary ui-btn--sm"
            @click="emit('enter-operations', String((row as ContestDetailData).id))"
          >
            <ArrowRight class="h-4 w-4" />
            进入运维台
          </button>
        </div>
      </template>
    </WorkspaceDataTable>
  </section>
</template>

<style scoped>
.contest-ops-section {
  padding: 0;
}

.contest-ops-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.contest-ops-directory {
  display: grid;
  gap: var(--space-4);
}

.contest-ops-table {
  padding: 0;
}

.contest-ops-table :deep(.workspace-data-table__cell + .workspace-data-table__cell) {
  border-left: 1px solid var(--workspace-table-line);
}

.contest-ops-table__contest {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.contest-ops-table__title {
  overflow: hidden;
  color: var(--journal-ink);
  font-size: var(--font-size-15);
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contest-ops-table__description,
.contest-ops-table__time {
  overflow: hidden;
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contest-ops-table__badge {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: var(--space-1) var(--space-2-5);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
  font-size: var(--font-size-12);
  font-weight: 600;
}

.contest-ops-table__badge--muted {
  background: color-mix(in srgb, var(--journal-border) 14%, transparent);
  color: var(--color-text-secondary);
}

@media (max-width: 768px) {
  .contest-ops-table :deep(.workspace-data-table) {
    min-width: 52rem;
  }

  .contest-ops-actions {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
