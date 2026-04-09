<script setup lang="ts">
import { computed } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import { getModeLabel, getStatusBadgeClass, getStatusLabel } from '@/utils/contest'

const props = defineProps<{
  contests: ContestDetailData[]
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  edit: [contest: ContestDetailData]
  export: [contest: ContestDetailData]
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
</script>

<template>
  <div class="space-y-5">
    <div class="contest-directory workspace-directory-list">
      <div class="contest-directory-head" aria-hidden="true">
        <span>竞赛</span>
        <span>模式</span>
        <span>状态</span>
        <span>时间窗口</span>
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
          <span
            class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
            :class="getStatusBadgeClass(contest.status)"
          >
            {{ getStatusLabel(contest.status) }}
          </span>
        </div>

        <div class="contest-row__window">
          <p>{{ formatTime(contest.starts_at) }}</p>
          <p class="contest-row__window-end">至 {{ formatTime(contest.ends_at) }}</p>
        </div>

        <div class="contest-row__actions" role="group" aria-label="竞赛操作">
          <button
            type="button"
            class="contest-action contest-action--primary"
            @click="emit('edit', contest)"
          >
            编辑
          </button>
          <button
            type="button"
            class="contest-action contest-action--ghost"
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
  --contest-directory-columns: minmax(18rem, 1.58fr) minmax(6rem, 0.56fr) minmax(6.5rem, 0.62fr)
    minmax(14rem, 1fr) minmax(11rem, 11rem);
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
.contest-row__window {
  font-size: var(--font-size-0-90);
  color: var(--journal-muted);
}

.contest-row__window {
  display: grid;
  gap: var(--space-1);
}

.contest-row__window p {
  margin: 0;
}

.contest-row__window-end {
  color: color-mix(in srgb, var(--journal-muted) 88%, var(--journal-ink));
}

.contest-row__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.contest-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 34px;
  border-radius: 10px;
  border: 1px solid transparent;
  padding: var(--space-1-5) var(--space-3);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  line-height: 1;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    color 150ms ease;
}

.contest-action:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--journal-accent) 45%, transparent);
  outline-offset: 2px;
}

.contest-action--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 32%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 76%, var(--journal-ink));
}

.contest-action--primary:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 18%, var(--journal-surface));
}

.contest-action--ghost {
  border-color: color-mix(in srgb, var(--journal-border) 80%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.contest-action--ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 32%, transparent);
  color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
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
