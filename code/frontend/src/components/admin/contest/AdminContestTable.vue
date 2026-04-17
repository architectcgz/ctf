<script setup lang="ts">
import { computed } from 'vue'
import { Swords } from 'lucide-vue-next'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

const props = defineProps<{
  contests: ContestDetailData[]
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  edit: [contest: ContestDetailData]
  export: [contest: ContestDetailData]
  workbench: [contest: ContestDetailData]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

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
  return contest.mode === 'awd' && (contest.status === 'running' || contest.status === 'frozen')
}
</script>

<template>
  <div class="space-y-5">
    <div class="contest-directory workspace-directory-list">
      <div class="contest-directory-head" aria-hidden="true">
        <span>竞赛</span>
        <span>模式</span>
        <span>状态</span>
        <span>开始时间</span>
        <span>结束时间</span>
        <span class="contest-directory-head__actions">操作</span>
      </div>

      <article v-for="contest in contests" :key="contest.id" class="contest-row">
        <div class="contest-row__identity">
          <h3 class="contest-row__title" :title="contest.title">{{ contest.title }}</h3>
          <p class="contest-row__description">
            {{ contest.description || '当前未填写竞赛描述。' }}
          </p>
        </div>

        <div class="contest-row__mode">{{ getModeLabel(contest.mode) }}</div>

        <div class="contest-row__status">
          <span class="ui-badge contest-status-pill" :class="getStatusPillClass(contest.status)">
            {{ getStatusLabel(contest.status) }}
          </span>
        </div>

        <div class="contest-row__starts-at">
          <p>{{ formatTime(contest.starts_at) }}</p>
        </div>

        <div class="contest-row__ends-at">
          <p>{{ formatTime(contest.ends_at) }}</p>
        </div>

        <div class="ui-row-actions contest-row__actions" role="group" aria-label="竞赛操作">
          <button
            v-if="canEnterWorkbench(contest)"
            :id="`contest-open-workbench-${contest.id}`"
            type="button"
            class="ui-btn ui-btn--sm ui-btn--primary contest-action contest-action--workbench"
            @click="emit('workbench', contest)"
          >
            <Swords class="h-3.5 w-3.5" />
            进入 AWD 赛区
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--sm ui-btn--secondary contest-action"
            @click="emit('edit', contest)"
          >
            编辑
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--sm ui-btn--secondary contest-action"
            @click="emit('export', contest)"
          >
            导出结果
          </button>
        </div>
      </article>
    </div>

    <div
      class="admin-pagination workspace-directory-pagination text-sm text-[var(--color-text-muted)]"
    >
      <AdminPaginationControls
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
  --contest-directory-columns: minmax(17rem, 1.46fr) minmax(6rem, 0.54fr) minmax(7rem, 0.68fr)
    minmax(9.5rem, 0.78fr) minmax(9.5rem, 0.78fr) minmax(11rem, 11rem);
  display: grid;
  gap: 0;
}

.contest-directory-head {
  display: grid;
  grid-template-columns: var(--contest-directory-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-directory-head > span {
  min-width: 0;
}

.contest-directory-head__actions {
  text-align: right;
}

.contest-row {
  display: grid;
  grid-template-columns: var(--contest-directory-columns);
  gap: var(--space-4);
  align-items: start;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.contest-row > div {
  min-width: 0;
}

.contest-row__identity {
  display: grid;
  gap: var(--space-1-5);
}

.contest-row__title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-1-00);
  font-weight: 600;
  color: var(--journal-ink);
}

.contest-row__description {
  margin: 0;
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: var(--font-size-0-875);
  line-height: 1.55;
  color: var(--journal-muted);
}

.contest-row__mode,
.contest-row__starts-at,
.contest-row__ends-at {
  font-size: var(--font-size-0-90);
  color: var(--journal-muted);
}

.contest-row__starts-at p,
.contest-row__ends-at p {
  margin: 0;
  line-height: 1.45;
}

.contest-row__starts-at p {
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

.contest-row__ends-at p {
  color: color-mix(in srgb, var(--journal-muted) 88%, var(--journal-ink));
}

.contest-status-pill {
  --ui-badge-radius: 999px;
  --ui-badge-padding: 0.35rem 0.75rem;
  --ui-badge-size: var(--font-size-0-78);
  --ui-badge-spacing: 0.02em;
  line-height: 1;
}

.contest-status-pill--running {
  --ui-badge-border: color-mix(in srgb, #22d3ee 38%, transparent);
  --ui-badge-background: color-mix(in srgb, #22d3ee 16%, var(--journal-surface));
  --ui-badge-color: #67e8f9;
}

.contest-status-pill--registering {
  --ui-badge-border: color-mix(in srgb, #f59e0b 34%, transparent);
  --ui-badge-background: color-mix(in srgb, #f59e0b 15%, var(--journal-surface));
  --ui-badge-color: #fbbf24;
}

.contest-status-pill--draft {
  --ui-badge-border: color-mix(in srgb, #a78bfa 28%, transparent);
  --ui-badge-background: color-mix(in srgb, #a78bfa 12%, var(--journal-surface));
  --ui-badge-color: #c4b5fd;
}

.contest-status-pill--frozen {
  --ui-badge-border: color-mix(in srgb, #60a5fa 30%, transparent);
  --ui-badge-background: color-mix(in srgb, #60a5fa 13%, var(--journal-surface));
  --ui-badge-color: #93c5fd;
}

.contest-status-pill--ended {
  --ui-badge-border: color-mix(in srgb, #34d399 28%, transparent);
  --ui-badge-background: color-mix(in srgb, #34d399 12%, var(--journal-surface));
  --ui-badge-color: #6ee7b7;
}

.contest-status-pill--cancelled,
.contest-status-pill--neutral {
  --ui-badge-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --ui-badge-background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  --ui-badge-color: color-mix(in srgb, var(--journal-muted) 92%, var(--journal-ink));
}

.contest-row__actions {
  justify-content: flex-end;
  flex-wrap: wrap;
}

.contest-action {
  min-width: 5.25rem;
}

.contest-action--workbench {
  --ui-btn-primary-bg: color-mix(in srgb, var(--color-success) 78%, var(--journal-ink));
  --ui-btn-primary-border: color-mix(in srgb, var(--color-success) 56%, transparent);
  --ui-btn-primary-color: white;
  box-shadow: 0 10px 24px color-mix(in srgb, var(--color-success) 18%, transparent);
}

@media (max-width: 1023px) {
  .contest-directory-head {
    display: none;
  }

  .contest-row {
    grid-template-columns: 1fr;
    gap: var(--space-2-5);
    padding: var(--space-4) 0;
  }

  .contest-row__actions {
    justify-content: flex-start;
  }
}
</style>
