<script setup lang="ts">
import { computed, ref } from 'vue'
import { MoreHorizontal } from 'lucide-vue-next'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import CActionMenu from '@/components/common/menus/CActionMenu.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

const props = defineProps<{
  contests: ContestDetailData[]
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  edit: [contest: ContestDetailData]
  announce: [contest: ContestDetailData]
  workbench: [contest: ContestDetailData]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const openActionMenuId = ref<string | null>(null)
const contestTableColumns = [
  { key: 'title', label: '竞赛', widthClass: 'w-[30%] min-w-[17rem]' },
  { key: 'mode', label: '模式', widthClass: 'w-[12%] min-w-[7rem]' },
  { key: 'status', label: '状态', widthClass: 'w-[12%] min-w-[7rem]', align: 'center' as const },
  { key: 'starts_at', label: '开始时间', widthClass: 'w-[16%] min-w-[10rem]' },
  { key: 'ends_at', label: '结束时间', widthClass: 'w-[16%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[14rem]', align: 'right' as const },
]

function formatTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getStatusPillClass(status: ContestStatus): string {
  if (status === 'running') return 'contest-status-pill--running'
  if (status === 'registering') return 'contest-status-pill--registering'
  if (status === 'draft' || status === 'published') return 'contest-status-pill--draft'
  if (status === 'frozen') return 'contest-status-pill--frozen'
  if (status === 'ended' || status === 'archived') return 'contest-status-pill--ended'
  if (status === 'cancelled') return 'contest-status-pill--cancelled'
  return 'contest-status-pill--neutral'
}

function canEnterWorkbench(contest: ContestDetailData): boolean {
  return (
    contest.mode === 'awd' &&
    (contest.status === 'running' || contest.status === 'frozen' || contest.status === 'ended')
  )
}

function canOpenActionMenu(contest: ContestDetailData): boolean {
  return contest.status !== 'ended'
}

function closeActionMenu(): void {
  openActionMenuId.value = null
}

function setActionMenuOpen(contestId: string, nextOpen: boolean): void {
  openActionMenuId.value = nextOpen ? contestId : null
}

function handleEdit(contest: ContestDetailData): void {
  closeActionMenu()
  emit('edit', contest)
}

function handleAnnounce(contest: ContestDetailData): void {
  closeActionMenu()
  emit('announce', contest)
}
</script>

<template>
  <div class="space-y-5">
    <WorkspaceDataTable
      class="contest-directory workspace-directory-list"
      :columns="contestTableColumns"
      :rows="contests"
      row-key="id"
    >
      <template #cell-title="{ row }">
        <div class="contest-table__identity">
          <h3
            class="contest-table__title"
            :title="(row as ContestDetailData).title"
          >
            {{ (row as ContestDetailData).title }}
          </h3>
          <p
            class="contest-table__description"
            :title="(row as ContestDetailData).description || '当前未填写竞赛描述。'"
          >
            {{ (row as ContestDetailData).description || '当前未填写竞赛描述。' }}
          </p>
        </div>
      </template>

      <template #cell-mode="{ row }">
        <span class="contest-table__muted">
          {{ getModeLabel((row as ContestDetailData).mode) }}
        </span>
      </template>

      <template #cell-status="{ row }">
        <span
          class="ui-badge contest-status-pill"
          :class="getStatusPillClass((row as ContestDetailData).status)"
        >
          {{ getStatusLabel((row as ContestDetailData).status) }}
        </span>
      </template>

      <template #cell-starts_at="{ row }">
        <span class="contest-table__time contest-table__time--start">
          {{ formatTime((row as ContestDetailData).starts_at) }}
        </span>
      </template>

      <template #cell-ends_at="{ row }">
        <span class="contest-table__time contest-table__time--end">
          {{ formatTime((row as ContestDetailData).ends_at) }}
        </span>
      </template>

      <template #cell-actions="{ row }">
        <div
          class="ui-row-actions contest-table__actions ui-row-actions--fixed"
          role="group"
          aria-label="竞赛操作"
        >
          <button
            v-if="canEnterWorkbench(row as ContestDetailData)"
            :id="`contest-open-workbench-${(row as ContestDetailData).id}`"
            type="button"
            class="ui-btn ui-btn--sm ui-btn--primary contest-action contest-action--workbench ui-row-action--main"
            @click="emit('workbench', row as ContestDetailData)"
          >
            运维
          </button>
          <button
            :id="`contest-row-edit-${(row as ContestDetailData).id}`"
            type="button"
            class="ui-btn ui-btn--sm ui-btn--secondary contest-action contest-action--edit ui-row-action--default"
            @click="handleEdit(row as ContestDetailData)"
          >
            编辑
          </button>
          <CActionMenu
            v-if="canOpenActionMenu(row as ContestDetailData)"
            :open="openActionMenuId === (row as ContestDetailData).id"
            title="Management"
            menu-label="更多竞赛操作"
            accent="var(--journal-accent, var(--color-primary))"
            class="ui-row-action--menu"
            @update:open="setActionMenuOpen((row as ContestDetailData).id, $event)"
          >
            <template #trigger="{ open, toggle, setTriggerRef }">
              <button
                :id="`contest-row-more-${(row as ContestDetailData).id}`"
                :ref="setTriggerRef"
                type="button"
                class="c-action-menu__trigger c-action-menu__trigger--icon"
                :aria-expanded="open ? 'true' : 'false'"
                aria-haspopup="menu"
                aria-label="更多竞赛操作"
                @click.stop="toggle"
              >
                <MoreHorizontal class="h-3.5 w-3.5" />
              </button>
            </template>

            <template #default>
              <button
                :id="`contest-row-menu-announce-${(row as ContestDetailData).id}`"
                type="button"
                class="c-action-menu__item"
                role="menuitem"
                @click="handleAnnounce(row as ContestDetailData)"
              >
                发布通知
              </button>
            </template>
          </CActionMenu>
        </div>
      </template>
    </WorkspaceDataTable>

    <div class="admin-pagination workspace-directory-pagination contest-pagination-tone text-sm">
      <PlatformPaginationControls
        :page="page"
        :total-pages="totalPages"
        :total="total"
        :total-label="`共 ${total} 场竞赛`"
        @change-page="emit('changePage', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.contest-directory {
  padding: 0;
}

.contest-directory :deep(.workspace-data-table__cell + .workspace-data-table__cell) {
  border-left: 1px solid var(--workspace-table-line);
}

.contest-table__identity {
  display: grid;
  gap: var(--space-1-5);
  min-width: 0;
}

.contest-table__title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-14);
  font-weight: 600;
  color: var(--journal-ink);
}

.contest-table__description {
  margin: 0;
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: var(--font-size-13);
  line-height: 1.55;
  color: var(--journal-muted);
}

.contest-table__muted,
.contest-table__time {
  display: block;
  overflow: hidden;
  font-size: var(--font-size-13);
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--journal-muted);
}

.contest-table__time--start {
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

.contest-table__time--end {
  color: color-mix(in srgb, var(--journal-muted) 88%, var(--journal-ink));
}

.contest-status-pill {
  --ui-badge-radius: 999px;
  --ui-badge-padding: 0.35rem 0.75rem;
  --ui-badge-size: var(--font-size-11);
  --ui-badge-spacing: 0.02em;
  line-height: 1;
}

.contest-status-pill--running {
  --ui-badge-border: color-mix(in srgb, var(--color-brand-swatch-cyan) 38%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--color-brand-swatch-cyan) 16%, var(--color-bg-surface));
  --ui-badge-color: color-mix(in srgb, var(--color-brand-swatch-cyan) 85%, var(--color-text-primary));
}

.contest-status-pill--registering {
  --ui-badge-border: color-mix(in srgb, var(--color-brand-swatch-orange) 34%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--color-brand-swatch-orange) 15%, var(--color-bg-surface));
  --ui-badge-color: var(--color-warning);
}

.contest-status-pill--draft {
  --ui-badge-border: color-mix(in srgb, var(--color-primary) 28%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface));
  --ui-badge-color: var(--color-primary);
}

.contest-status-pill--frozen {
  --ui-badge-border: color-mix(in srgb, var(--color-brand-swatch-blue) 30%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--color-brand-swatch-blue) 13%, var(--color-bg-surface));
  --ui-badge-color: color-mix(in srgb, var(--color-brand-swatch-blue) 85%, var(--color-text-primary));
}

.contest-status-pill--ended {
  --ui-badge-border: color-mix(in srgb, var(--color-success) 28%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--color-success) 12%, var(--color-bg-surface));
  --ui-badge-color: var(--color-success);
}

.contest-status-pill--cancelled,
.contest-status-pill--neutral {
  --ui-badge-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  --ui-badge-color: color-mix(in srgb, var(--journal-muted) 92%, var(--journal-ink));
}

.contest-table__actions {
  --ui-row-action-main-width: 4.25rem;
  justify-content: flex-end;
}

.contest-pagination-tone {
  color: var(--color-text-muted);
}

@media (max-width: 1023px) {
  .contest-directory :deep(.workspace-data-table) {
    min-width: 58rem;
  }

  .contest-table__actions {
    justify-content: flex-start;
  }
}
</style>
