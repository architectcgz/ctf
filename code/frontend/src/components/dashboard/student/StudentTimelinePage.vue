<script setup lang="ts">
import { computed } from 'vue'
import { CalendarClock, CircleCheckBig, Play, Send } from 'lucide-vue-next'

import type { TimelineEvent } from '@/api/contracts'
import { formatTime } from '@/utils/format'

import { timelineSummary, timelineTypeLabel, timelineTypeTone } from './utils'

const props = withDefaults(
  defineProps<{
    timeline: TimelineEvent[]
    embedded?: boolean
  }>(),
  {
    embedded: false,
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
const groupedTimeline = computed(() => {
  const groups = new Map<string, TimelineEvent[]>()
  props.timeline.forEach((event) => {
    const key = new Date(event.created_at).toLocaleDateString('zh-CN')
    groups.set(key, [...(groups.get(key) || []), event])
  })
  return Array.from(groups.entries()).map(([date, events]) => ({ date, events }))
})
</script>

<template>
  <section
    class="journal-soft-surface space-y-6 flex min-h-full flex-1 flex-col"
    :class="
      embedded
        ? 'journal-shell-embedded'
        : 'journal-shell journal-hero rounded-[30px] border px-6 py-6 md:px-8'
    "
  >
    <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
      <div>
        <div class="journal-eyebrow">Training Timeline</div>
        <h1 class="journal-page-title mt-3 text-[var(--journal-ink)]">训练节奏总览</h1>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          看最近训练记录和节奏变化。
        </p>
      </div>

      <article class="journal-brief rounded-[24px] border px-5 py-5">
        <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
          <CalendarClock class="h-5 w-5 text-[var(--journal-accent)]" />
          节奏快照
        </div>
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
          <div class="journal-note">
            <div class="journal-note-label">成功解题</div>
            <div class="journal-note-value">{{ solveCount }} 次</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">提交次数</div>
            <div class="journal-note-value">{{ submitCount }} 次</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">实例操作</div>
            <div class="journal-note-value">{{ instanceCount }} 次</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">总记录</div>
            <div class="journal-note-value">{{ timeline.length }} 条</div>
          </div>
        </div>
      </article>
    </div>

    <div
      class="timeline-board mt-6 px-1 pt-5 md:px-2 md:pt-6"
      :class="{ 'timeline-board--embedded': embedded }"
    >
      <section class="timeline-section">
        <div class="journal-eyebrow journal-eyebrow-soft">Rhythm Signals</div>
        <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">节奏信号</h3>
        <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
          先看整体节奏，再回到下方时间线定位具体动作。
        </p>

        <div class="timeline-signal-list mt-5">
          <article class="timeline-signal-item">
            <div class="flex items-start gap-3">
              <div class="stat-icon stat-icon--success">
                <CircleCheckBig class="h-5 w-5" />
              </div>
              <div>
                <div class="text-sm font-semibold text-[var(--journal-ink)]">成功信号</div>
                <div class="mt-2 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
                  {{ solveCount }}
                </div>
                <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                  成功解题偏少时，适合先回到推荐页选更贴近当前阶段的题目。
                </p>
              </div>
            </div>
          </article>

          <article class="timeline-signal-item">
            <div class="flex items-start gap-3">
              <div class="stat-icon stat-icon--warning">
                <Send class="h-5 w-5" />
              </div>
              <div>
                <div class="text-sm font-semibold text-[var(--journal-ink)]">提交密度</div>
                <div class="mt-2 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
                  {{ submitCount }}
                </div>
                <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                  提交多但成功少，通常说明方向跑偏，需要回看能力画像或题目切入点。
                </p>
              </div>
            </div>
          </article>

          <article class="timeline-signal-item">
            <div class="flex items-start gap-3">
              <div class="stat-icon stat-icon--primary">
                <Play class="h-5 w-5" />
              </div>
              <div>
                <div class="text-sm font-semibold text-[var(--journal-ink)]">实例节奏</div>
                <div class="mt-2 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
                  {{ instanceCount }}
                </div>
                <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                  实例操作多但提交少，通常代表分析阶段过长，适合更快进入验证。
                </p>
              </div>
            </div>
          </article>
        </div>
      </section>

      <section class="timeline-section">
        <div class="journal-eyebrow journal-eyebrow-soft">Timeline Log</div>
        <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">训练时间线</h3>
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
      </section>
    </div>
  </section>
</template>

<style scoped>
.journal-brief {
  border-color: var(--journal-shell-border);
  background: var(--journal-surface-subtle);
}

.timeline-board {
  border-top: 1px solid var(--journal-divider);
}

.timeline-board--embedded {
  margin-top: 1.25rem;
}

.timeline-section + .timeline-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--journal-divider);
}

.timeline-signal-list,
.timeline-group-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

.timeline-signal-item {
  padding: 1rem 1.1rem;
}

.timeline-signal-item + .timeline-signal-item {
  border-top: 1px solid var(--journal-divider);
}

.timeline-group {
  padding: 1rem 1.1rem;
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

@media (min-width: 1280px) {
  .timeline-signal-list {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .timeline-signal-item + .timeline-signal-item {
    border-top: 0;
    border-left: 1px solid var(--journal-divider);
  }
}


:global([data-theme='dark']) .timeline-signal-list,
:global([data-theme='dark']) .timeline-group-list {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}
</style>
