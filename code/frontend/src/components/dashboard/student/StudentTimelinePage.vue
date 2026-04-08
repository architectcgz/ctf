<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import type { TimelineEvent } from '@/api/contracts'
import { formatTime } from '@/utils/format'

import { timelineSummary, timelineTypeLabel, timelineTypeTone } from './utils'

const props = withDefaults(
  defineProps<{
    timeline: TimelineEvent[]
    embedded?: boolean
    pageSize?: number
  }>(),
  {
    embedded: false,
    pageSize: 10,
  }
)

const solveCount = computed(() => props.timeline.filter((item) => item.type === 'solve').length)
const submitCount = computed(() => props.timeline.filter((item) => item.type === 'submit').length)
const instanceCount = computed(
  () =>
    props.timeline.filter(
      (item) =>
        item.type === 'instance' ||
        item.type === 'instance_access' ||
        item.type === 'instance_proxy_request' ||
        item.type === 'instance_extend' ||
        (item.meta?.raw_type as string | undefined) === 'instance_access' ||
        (item.meta?.raw_type as string | undefined) === 'instance_proxy_request' ||
        (item.meta?.raw_type as string | undefined) === 'instance_extend'
    ).length
)
const sortedTimeline = computed(() =>
  [...props.timeline].sort(
    (left, right) => new Date(right.created_at).getTime() - new Date(left.created_at).getTime()
  )
)
const totalTimelineCount = computed(() => sortedTimeline.value.length)
const totalTimelinePages = computed(() =>
  Math.max(1, Math.ceil(totalTimelineCount.value / Math.max(1, props.pageSize)))
)
const timelinePage = ref(1)

watch(
  () => totalTimelinePages.value,
  (nextTotalPages) => {
    timelinePage.value = Math.min(timelinePage.value, nextTotalPages)
  },
  { immediate: true }
)

const pagedTimeline = computed(() => {
  const safePage = Math.max(1, Math.floor(timelinePage.value || 1))
  const safePageSize = Math.max(1, Math.floor(props.pageSize || 10))
  const start = (safePage - 1) * safePageSize
  return sortedTimeline.value.slice(start, start + safePageSize)
})

const groupedTimeline = computed(() => {
  const groups = new Map<string, TimelineEvent[]>()
  pagedTimeline.value.forEach((event) => {
    const key = new Date(event.created_at).toLocaleDateString('zh-CN')
    groups.set(key, [...(groups.get(key) || []), event])
  })
  return Array.from(groups.entries()).map(([date, events]) => ({ date, events }))
})

function changeTimelinePage(page: number): void {
  timelinePage.value = page
}
</script>

<template>
  <section
    class="journal-soft-surface space-y-6 flex min-h-full flex-1 flex-col"
    :class="
      embedded
        ? 'journal-shell-embedded'
        : 'journal-shell journal-hero timeline-shell-flat px-6 py-6 md:px-8'
    "
  >
    <div class="timeline-header">
      <div>
        <h1 class="journal-page-title workspace-tab-heading__title text-[var(--journal-ink)]">
          训练记录总览
        </h1>
        <p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          看最近训练记录和节奏变化。
        </p>
        <div class="timeline-metric-grid mt-5">
          <article class="timeline-metric-card teacher-surface-section">
            <div class="journal-note-label">成功解题</div>
            <div class="journal-note-value">{{ solveCount }} 次</div>
          </article>
          <article class="timeline-metric-card teacher-surface-section">
            <div class="journal-note-label">提交次数</div>
            <div class="journal-note-value">{{ submitCount }} 次</div>
          </article>
          <article class="timeline-metric-card teacher-surface-section">
            <div class="journal-note-label">实例操作</div>
            <div class="journal-note-value">{{ instanceCount }} 次</div>
          </article>
          <article class="timeline-metric-card teacher-surface-section">
            <div class="journal-note-label">总记录</div>
            <div class="journal-note-value">{{ totalTimelineCount }} 条</div>
          </article>
        </div>
      </div>
    </div>

    <div
      class="timeline-board mt-0 px-0 pt-0 md:px-0 md:pt-0"
      :class="{ 'timeline-board--embedded': embedded }"
    >
      <div class="journal-divider timeline-board-divider" aria-hidden="true" />
      <section class="timeline-section">
        <div class="journal-eyebrow journal-eyebrow-soft">Timeline Log</div>
        <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">训练记录</h3>
        <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
          按日期回看最近的提交、解题和实例操作。
        </p>

        <div
          v-if="groupedTimeline.length === 0"
          class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-shell-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
        >
          当前还没有训练动态。
        </div>

        <div v-else class="timeline-group-list mt-5">
          <section v-for="group in groupedTimeline" :key="group.date" class="timeline-group">
            <div class="timeline-group-date">{{ group.date }}</div>
            <div class="timeline-event-list">
              <article v-for="event in group.events" :key="event.id" class="timeline-event-item">
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div class="flex items-start gap-3">
                    <span
                      class="status-dot mt-1.5 shrink-0"
                      :class="`status-dot-${event.type === 'solve' ? 'solved' : event.type.includes('instance') ? 'ready' : 'idle'}`"
                    />
                    <div>
                      <div class="text-sm font-semibold text-[var(--journal-ink)]">
                        {{ event.title }}
                      </div>
                      <div class="mt-1 text-sm leading-6 text-[var(--journal-muted)]">
                        {{ timelineSummary(event) }}
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <span
                      class="rounded-full border px-2.5 py-1 text-xs font-medium"
                      :class="timelineTypeTone(event)"
                    >
                      {{ timelineTypeLabel(event) }}
                    </span>
                    <span class="tech-font text-xs text-[var(--journal-muted)]">{{
                      formatTime(event.created_at)
                    }}</span>
                  </div>
                </div>
              </article>
            </div>
          </section>
        </div>

        <div v-if="totalTimelineCount > 0" class="timeline-pagination mt-5">
          <PagePaginationControls
            :page="timelinePage"
            :total-pages="totalTimelinePages"
            :total="totalTimelineCount"
            total-label="训练记录总数"
            show-jump
            @change-page="changeTimelinePage"
          />
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.timeline-shell-flat.journal-shell.journal-hero {
  border: 0;
  border-radius: 0 !important;
  box-shadow: none;
  overflow: visible;
}

.timeline-header {
  display: grid;
  gap: 1rem;
}

.timeline-metric-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.timeline-metric-card {
  min-height: 100%;
  padding: 0.82rem 0.95rem 0.78rem;
}

.timeline-metric-card.teacher-surface-section {
  background: linear-gradient(
    165deg,
    color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base))
  );
  box-shadow: 0 10px 20px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent);
}

.timeline-metric-card .journal-note-value {
  margin-top: 0.5rem;
  font-size: 1.05rem;
  font-weight: 700;
}

.timeline-board {
  border-top: 0;
}

.timeline-header + .timeline-board {
  margin-top: var(--space-1) !important;
}

.timeline-board-divider {
  --journal-divider-margin-block: 0 var(--space-1-5);
}

.timeline-board--embedded {
  margin-top: 1.25rem;
}

.timeline-section + .timeline-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--journal-divider);
}

.timeline-group-list {
  border-radius: 0;
  border: 0;
  background: transparent;
}

.timeline-group {
  padding: 1rem 0;
}

.timeline-group + .timeline-group {
  border-top: 1px solid var(--journal-divider);
}

.timeline-group-date {
  margin-bottom: 0.85rem;
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 90%, transparent);
  padding: 0.28rem 0.72rem;
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.timeline-event-list {
  position: relative;
}

.timeline-event-item {
  position: relative;
  padding: 0.95rem 0 0.95rem 1.1rem;
}

.timeline-event-item::before {
  content: '';
  position: absolute;
  left: 0.2rem;
  top: 0;
  bottom: 0;
  border-left: 1px solid var(--journal-divider);
}

.timeline-event-item:first-child {
  padding-top: 0.25rem;
}

.timeline-event-item:first-child::before {
  top: 0.65rem;
}

.timeline-event-item:last-child {
  padding-bottom: 0.2rem;
}

.timeline-event-item:last-child::before {
  bottom: 0.55rem;
}

.timeline-event-item + .timeline-event-item {
  border-top: 1px solid var(--journal-divider);
}

.timeline-pagination {
  border-top: 1px solid var(--journal-divider);
  padding-top: 0.35rem;
}

.stat-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid var(--journal-soft-border);
  background: var(--journal-surface-subtle);
}

.stat-icon--success {
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.2);
  background: rgba(16, 185, 129, 0.08);
}

.stat-icon--warning {
  color: #f59e0b;
  border-color: rgba(245, 158, 11, 0.2);
  background: rgba(245, 158, 11, 0.08);
}

.stat-icon--primary {
  color: var(--journal-accent);
  border-color: color-mix(in srgb, var(--journal-accent) 20%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 999px;
}

.status-dot-solved {
  background: #22c55e;
}

.status-dot-ready {
  background: #10b981;
  animation: dot-pulse 1.8s infinite;
}

.status-dot-idle {
  background: #94a3b8;
}

.tech-font {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

@keyframes dot-pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.38);
  }
  70% {
    box-shadow: 0 0 0 8px rgba(16, 185, 129, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}

@media (max-width: 1023px) {
  .timeline-metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .timeline-metric-grid {
    grid-template-columns: 1fr;
  }
}

:global([data-theme='dark']) .timeline-group-list {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}
</style>
