<script setup lang="ts">
import { computed } from 'vue'
import { CalendarClock, CircleCheckBig, Play, Send } from 'lucide-vue-next'

import type { TimelineEvent } from '@/api/contracts'
import { formatDate, formatTime } from '@/utils/format'

import { timelineSummary, timelineTypeLabel, timelineTypeTone } from './utils'

const props = defineProps<{
  timeline: TimelineEvent[]
}>()

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
  <div class="journal-shell space-y-6">
    <!-- Hero -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Training Timeline</div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            训练节奏总览
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            把成功解题、提交动作和实例操作放在同一视图里，更容易看清最近一段训练是否顺畅。
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
    </section>

    <!-- 节奏信号 -->
    <section class="grid gap-4 md:grid-cols-3">
      <article class="journal-panel rounded-[24px] border px-6 py-5">
        <div class="flex items-center gap-3">
          <div class="stat-icon stat-icon--success">
            <CircleCheckBig class="h-5 w-5" />
          </div>
          <div class="journal-eyebrow">成功信号</div>
        </div>
        <div class="mt-4 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">{{ solveCount }}</div>
        <div class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
          次成功解题。若偏低，建议回到推荐页选更适合当前阶段的题目。
        </div>
      </article>

      <article class="journal-panel rounded-[24px] border px-6 py-5">
        <div class="flex items-center gap-3">
          <div class="stat-icon stat-icon--warning">
            <Send class="h-5 w-5" />
          </div>
          <div class="journal-eyebrow">提交密度</div>
        </div>
        <div class="mt-4 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">{{ submitCount }}</div>
        <div class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
          次提交。若提交多但成功少，说明方向可能跑偏，需要回看能力画像。
        </div>
      </article>

      <article class="journal-panel rounded-[24px] border px-6 py-5">
        <div class="flex items-center gap-3">
          <div class="stat-icon stat-icon--primary">
            <Play class="h-5 w-5" />
          </div>
          <div class="journal-eyebrow">实例节奏</div>
        </div>
        <div class="mt-4 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">{{ instanceCount }}</div>
        <div class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
          次实例操作。操作多但提交少，通常代表分析阶段过长。
        </div>
      </article>
    </section>

    <!-- 时间线 -->
    <section class="journal-panel rounded-[24px] border px-6 py-6">
      <div class="journal-eyebrow">Timeline Log</div>
      <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">训练时间线</h3>

      <div
        v-if="groupedTimeline.length === 0"
        class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
      >
        当前还没有训练动态。
      </div>

      <div v-else class="mt-5 space-y-8">
        <section v-for="group in groupedTimeline" :key="group.date">
          <div class="mb-3 text-xs font-semibold uppercase tracking-[0.22em] text-[var(--journal-muted)]">
            {{ group.date }}
          </div>
          <div class="space-y-3">
            <article
              v-for="event in group.events"
              :key="event.id"
              class="journal-log rounded-[18px] border px-5 py-4"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="flex items-start gap-3">
                  <span class="status-dot mt-1.5 shrink-0" :class="`status-dot-${event.type === 'solve' ? 'solved' : event.type.includes('instance') ? 'ready' : 'idle'}`" />
                  <div>
                    <div class="text-sm font-semibold text-[var(--journal-ink)]">{{ event.title }}</div>
                    <div class="mt-1 text-sm leading-6 text-[var(--journal-muted)]">{{ timelineSummary(event) }}</div>
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <span class="rounded-full border px-2.5 py-1 text-xs font-medium" :class="timelineTypeTone(event)">
                    {{ timelineTypeLabel(event) }}
                  </span>
                  <span class="tech-font text-xs text-[var(--journal-muted)]">{{ formatTime(event.created_at) }}</span>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </section>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-ink: #0f172a;
  --journal-muted: #475569;
  --journal-border: rgba(226, 232, 240, 0.72);
  --journal-surface: #ffffff;
  --journal-surface-subtle: #f8fafc;
  font-family: "Inter", "Noto Sans SC", system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(191, 219, 254, 0.75), transparent 15rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-panel,
.journal-log {
  border-color: var(--journal-border);
  background: var(--journal-surface);
}

.journal-panel {
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-log {
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.03);
  transition: all 0.2s ease-in-out;
}

.journal-log:hover {
  border-color: #6366f1;
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.06);
}

.journal-note {
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: var(--journal-surface-subtle);
  padding: 0.95rem 1rem;
}

.journal-note-label,
.journal-eyebrow {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.26em;
  text-transform: uppercase;
  color: #64748b;
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.stat-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid rgba(226, 232, 240, 0.72);
  background: #f8fafc;
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
  color: #4f46e5;
  border-color: rgba(79, 70, 229, 0.2);
  background: rgba(79, 70, 229, 0.08);
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
  font-family: "JetBrains Mono", "Fira Code", monospace;
}

@keyframes dot-pulse {
  0% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.38); }
  70% { box-shadow: 0 0 0 8px rgba(16, 185, 129, 0); }
  100% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}
</style>
