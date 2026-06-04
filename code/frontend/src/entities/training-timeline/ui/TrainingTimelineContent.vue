<template>
  <div class="timeline-content">
    <div class="workspace-panel-header timeline-header">
      <div class="workspace-panel-header__intro">
        <div class="workspace-overline">
          Timeline
        </div>
        <h1 class="journal-page-title workspace-page-title journal-soft-page-title">
          训练记录总览
        </h1>
        <p class="workspace-page-copy max-w-2xl">
          按时间回看最近训练动作，看看节奏有没有断。
        </p>
      </div>
      <div
        class="workspace-panel-header__summary timeline-metric-grid progress-strip metric-panel-grid metric-panel-default-surface"
      >
        <article
          v-for="metric in timelineMetrics"
          :key="metric.key"
          class="timeline-metric-card progress-card metric-panel-card"
        >
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>{{ metric.label }}</span>
            <component :is="metric.icon" class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ metric.value }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            {{ metric.helper }}
          </div>
        </article>
      </div>
    </div>

    <div class="timeline-board mt-0 px-0 pt-0 md:px-0 md:pt-0">
      <section class="timeline-section workspace-directory-section">
        <section
          class="timeline-directory-shell teacher-directory-shell workspace-directory-list workspace-directory-list--catalog"
        >
          <header class="list-heading timeline-list-heading">
            <div class="timeline-list-heading__body">
              <div class="workspace-overline">Timeline Log</div>
              <h3 class="list-heading__title journal-soft-section-title">
                训练记录
              </h3>
            </div>
          </header>

          <div
            v-if="groupedTimeline.length === 0"
            class="journal-soft-empty-state workspace-directory-empty timeline-empty-state"
          >
            当前还没有训练动态。
          </div>

          <div
            v-else
            class="timeline-group-list"
          >
            <section
              v-for="group in groupedTimeline"
              :key="group.date"
              class="timeline-group"
            >
              <div class="timeline-group-date">
                {{ group.date }}
              </div>
              <div class="timeline-event-list">
                <article
                  v-for="event in group.events"
                  :key="event.id"
                  class="timeline-event-item"
                >
                  <div class="flex flex-wrap items-start justify-between gap-3">
                    <div class="flex items-start gap-3">
                      <span
                        class="status-dot mt-1.5 shrink-0"
                        :class="`status-dot-${event.type === 'solve' ? 'solved' : event.type.includes('instance') ? 'ready' : 'idle'}`"
                      />
                      <div>
                        <div class="journal-soft-body-title text-sm font-semibold">
                          {{ event.title }}
                        </div>
                        <div class="journal-soft-body-copy mt-1 text-sm leading-6">
                          {{ timelineSummary(event) }}
                        </div>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <span :class="timelineTypeTone(event)">
                        {{ timelineTypeLabel(event) }}
                      </span>
                      <span class="journal-soft-meta tech-font text-xs">{{
                        formatTime(event.created_at)
                      }}</span>
                    </div>
                  </div>
                </article>
              </div>
            </section>
          </div>

          <div
            v-if="totalTimelineCount > 0"
            class="timeline-pagination workspace-directory-pagination"
          >
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
      </section>
    </div>
  </div>
</template>

<style scoped>
.timeline-content {
  min-width: 0;
}

.timeline-metric-grid {
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.timeline-metric-grid.metric-panel-default-surface {
  --metric-panel-border: var(--journal-soft-border);
  --metric-panel-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 14%, transparent),
      transparent 44%
    ),
    linear-gradient(
      165deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base))
    );
  --metric-panel-shadow: 0 10px 20px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent);
}

.timeline-metric-card {
  min-height: 100%;
}

.timeline-board {
  border-top: 0;
}

.timeline-header + .timeline-board {
  margin-top: var(--workspace-panel-divider-gap, var(--space-workspace-panel-divider-gap));
}

.timeline-list-heading__body {
  display: grid;
  gap: var(--space-2);
}

.timeline-list-heading__copy {
  margin: 0;
}

.timeline-section {
  --workspace-directory-section-gap: 0;
}

.timeline-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: var(--journal-soft-border);
  --workspace-directory-row-divider: var(--journal-divider);
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 calc(var(--space-4) + var(--space-0-5)) calc(var(--space-8) + var(--space-0-5))
    color-mix(in srgb, var(--color-shadow-soft) 18%, transparent);
}

.teacher-directory-shell {
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  box-shadow: 0 calc(var(--space-4) + var(--space-0-5)) calc(var(--space-8) + var(--space-0-5))
    color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.timeline-empty-state {
  margin-top: 0;
  padding-inline: 0;
}

.timeline-group-list {
  min-width: 0;
  border-top: 1px solid var(--workspace-directory-row-divider);
  padding-top: var(--space-4);
}

.timeline-group {
  padding: var(--space-4) 0;
}

.timeline-group + .timeline-group {
  border-top: 1px solid var(--journal-divider);
}

.timeline-group-date {
  margin-bottom: var(--space-3-5);
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 90%, transparent);
  padding: var(--space-1) var(--space-3);
  font-size: var(--font-size-0-68);
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
  padding: var(--space-4) 0 var(--space-4) var(--space-4-5);
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
  border-top: 1px solid var(--workspace-directory-row-divider);
}

.timeline-pagination {
  padding-inline: 0;
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
  color: var(--color-success);
  border-color: color-mix(in srgb, var(--color-success) 20%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
}

.stat-icon--warning {
  color: var(--color-warning);
  border-color: color-mix(in srgb, var(--color-warning) 20%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
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

.timeline-type-pill {
  display: inline-flex;
  align-items: center;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.25rem 0.625rem;
  font-size: var(--font-size-0-72);
  font-weight: 600;
}

.timeline-type-pill--primary {
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 30%, transparent);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
}

.timeline-type-pill--danger {
  color: var(--color-danger);
  border-color: color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
}

.timeline-type-pill--success {
  color: var(--color-success);
  border-color: color-mix(in srgb, var(--color-success) 30%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
}

.timeline-type-pill--warning {
  color: var(--color-warning);
  border-color: color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
}

.timeline-type-pill--reverse {
  color: var(--color-cat-reverse);
  border-color: color-mix(in srgb, var(--color-cat-reverse) 30%, transparent);
  background: color-mix(in srgb, var(--color-cat-reverse) 10%, transparent);
}

.status-dot-solved {
  background: var(--color-success);
}

.status-dot-ready {
  background: color-mix(in srgb, var(--color-success) 92%, var(--journal-ink));
  animation: dot-pulse 1.8s infinite;
}

.status-dot-idle {
  background: var(--color-text-muted);
}

.tech-font {
  font-family: var(--font-family-mono);
}

@keyframes dot-pulse {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-success) 38%, transparent);
  }
  70% {
    box-shadow: 0 0 0 8px color-mix(in srgb, var(--color-success) 0%, transparent);
  }
  100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-success) 0%, transparent);
  }
}

@media (max-width: 1023px) {
  .timeline-metric-grid {
    --metric-panel-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .timeline-metric-grid {
    --metric-panel-columns: 1fr;
  }
}
</style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Activity, FileText, Server, Target } from 'lucide-vue-next'

import PagePaginationControls from '@/shared/ui/common/PagePaginationControls.vue'
import { SUBMISSION_RECORDS_PAGE_SIZE } from '@/utils/constants'
import { formatTime } from '@/utils/format'

import {
  timelineSummary,
  timelineTypeLabel,
  timelineTypeTone,
  type TimelineEvent,
} from '../model'

const props = withDefaults(
  defineProps<{
    timeline: TimelineEvent[]
    pageSize?: number
  }>(),
  {
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
const timelineMetrics = computed(() => [
  {
    key: 'solve',
    label: '成功解题',
    value: solveCount.value,
    icon: Activity,
    helper: '累计命中 Flag 的训练次数',
  },
  {
    key: 'submit',
    label: '提交次数',
    value: submitCount.value,
    icon: Target,
    helper: '最近训练周期内的提交总量',
  },
  {
    key: 'instance',
    label: '实例操作',
    value: instanceCount.value,
    icon: Server,
    helper: '启动、访问和续期等实例相关动作',
  },
  {
    key: 'total',
    label: '总记录',
    value: totalTimelineCount.value,
    icon: FileText,
    helper: '当前时间线中收录的训练事件数量',
  },
])

watch(
  () => totalTimelinePages.value,
  (nextTotalPages) => {
    timelinePage.value = Math.min(timelinePage.value, nextTotalPages)
  },
  { immediate: true }
)

const pagedTimeline = computed(() => {
  const safePage = Math.max(1, Math.floor(timelinePage.value || 1))
  const safePageSize = Math.max(1, Math.floor(props.pageSize || SUBMISSION_RECORDS_PAGE_SIZE))
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
