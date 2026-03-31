<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { CalendarRange, Clock3, Flame, Trophy } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getContests } from '@/api/contest'
import type { ContestListItem, ContestStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { usePagination } from '@/composables/usePagination'
import {
  getContestAccentColor,
  getContestActionLabel,
  getModeLabel,
  getStatusLabel,
} from '@/utils/contest'

const router = useRouter()
const { list, loading, error, refresh } = usePagination(getContests)

onMounted(() => {
  refresh()
})

interface ContestSummaryMetric {
  key: string
  label: string
  value: number
  hint: string
}

const summaryMetrics = computed<ContestSummaryMetric[]>(() => {
  const runningCount = list.value.filter((contest) => contest.status === 'running').length
  const registeringCount = list.value.filter((contest) => contest.status === 'registering').length
  const endedCount = list.value.filter((contest) =>
    ['ended', 'cancelled', 'archived', 'frozen'].includes(contest.status)
  ).length

  return [
    { key: 'total', label: '竞赛总数', value: list.value.length, hint: '当前可查看的竞赛数量' },
    { key: 'running', label: '进行中', value: runningCount, hint: '已经开赛且仍可参与' },
    { key: 'registering', label: '报名中', value: registeringCount, hint: '近期可以报名的竞赛' },
    { key: 'ended', label: '已结束', value: endedCount, hint: '可用于复盘或排行回看' },
  ]
})

const loadErrorMessage = computed(() => {
  if (!error.value) return ''
  if (error.value instanceof Error && error.value.message.trim().length > 0) {
    return error.value.message
  }
  return '竞赛列表加载失败，请稍后重试。'
})

const nowMs = computed(() => Date.now())

function formatTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatDuration(ms: number): string {
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour

  if (ms <= minute) return '少于 1 分钟'
  if (ms >= day) {
    const days = Math.floor(ms / day)
    const hours = Math.floor((ms % day) / hour)
    return `${days} 天 ${hours} 小时`
  }
  if (ms >= hour) {
    const hours = Math.floor(ms / hour)
    const minutes = Math.floor((ms % hour) / minute)
    return `${hours} 小时 ${minutes} 分钟`
  }
  return `${Math.floor(ms / minute)} 分钟`
}

function getTimelineHint(contest: ContestListItem): string {
  const startMs = new Date(contest.starts_at).getTime()
  const endMs = new Date(contest.ends_at).getTime()
  if (Number.isNaN(startMs) || Number.isNaN(endMs)) return '时间待定'

  if (nowMs.value < startMs) {
    return `距开始 ${formatDuration(startMs - nowMs.value)}`
  }
  if (nowMs.value <= endMs && ['running', 'registering', 'published'].includes(contest.status)) {
    return `距结束 ${formatDuration(endMs - nowMs.value)}`
  }
  return '竞赛已结束'
}

function goToDetail(id: string): void {
  router.push(`/contests/${id}`)
}

function openContest(contest: ContestListItem): void {
  goToDetail(contest.id)
}

function onKeyboardOpen(event: KeyboardEvent, contest: ContestListItem): void {
  if (event.key !== 'Enter' && event.key !== ' ') return
  event.preventDefault()
  openContest(contest)
}

function contestAccentStyle(status: ContestStatus): Record<string, string> {
  return { '--contest-accent': getContestAccentColor(status) }
}
</script>

<template>
  <section class="journal-shell space-y-6 journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="journal-eyebrow">Contest Center</div>
      <h2
        class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
      >
        竞赛中心
      </h2>
      <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
        在这里查看当前可参加和已结束的竞赛。
      </p>

      <div class="mt-5 grid gap-3 sm:grid-cols-2">
        <article
          v-for="stat in summaryMetrics"
          :key="stat.key"
          class="journal-metric rounded-[20px] border px-4 py-4"
        >
          <div
            class="text-[11px] font-semibold uppercase tracking-[0.2em] text-[var(--journal-muted)]"
          >
            {{ stat.label }}
          </div>
          <div class="mt-2 text-lg font-semibold text-[var(--journal-ink)] tech-font">
            {{ stat.value }}
          </div>
          <div class="mt-2 text-xs leading-5 text-[var(--journal-muted)]">
            {{ stat.hint }}
          </div>
        </article>
      </div>

      <div class="contest-board mt-6 px-1 pt-5 md:px-2 md:pt-6">
        <div v-if="loading" class="space-y-3 py-1">
          <div
            v-for="i in 4"
            :key="i"
            class="h-24 rounded-[18px] animate-pulse"
            style="background: rgba(226, 232, 240, 0.5)"
          />
        </div>

        <AppEmpty
          v-else-if="loadErrorMessage"
          class="contest-empty-state"
          icon="AlertTriangle"
          title="竞赛列表加载失败"
          :description="loadErrorMessage"
        >
          <template #action>
            <button type="button" class="contest-btn mt-3" @click="refresh">重试</button>
          </template>
        </AppEmpty>

        <AppEmpty
          v-else-if="list.length === 0"
          class="contest-empty-state"
          icon="Flag"
          title="暂无竞赛"
          description="当前没有可展示的竞赛，稍后再来查看新的开赛计划。"
        />

        <div v-else class="contest-list mt-5">
          <article
            v-for="contest in list"
            :key="contest.id"
            class="contest-item journal-log px-5 py-5 cursor-pointer"
            :style="{
              ...contestAccentStyle(contest.status),
              borderLeftWidth: '3px',
              borderLeftColor: 'var(--contest-accent)',
            }"
            tabindex="0"
            @click="openContest(contest)"
            @keydown="onKeyboardOpen($event, contest)"
          >
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap gap-2 items-center">
                  <span class="contest-badge">{{ getStatusLabel(contest.status) }}</span>
                  <span class="contest-mode">{{ getModeLabel(contest.mode) }}</span>
                </div>
                <h3 class="mt-2 font-semibold text-lg text-[var(--journal-ink)] leading-snug">
                  {{ contest.title }}
                </h3>
                <div class="mt-3 flex flex-wrap gap-4">
                  <div class="flex items-center gap-1.5 text-xs text-[var(--journal-muted)]">
                    <CalendarRange class="h-3.5 w-3.5" />
                    <span
                      >{{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}</span
                    >
                  </div>
                  <div
                    class="flex items-center gap-1.5 text-xs font-medium"
                    :style="{ color: 'var(--contest-accent)' }"
                  >
                    <Clock3 class="h-3.5 w-3.5" />
                    <span>{{ getTimelineHint(contest) }}</span>
                  </div>
                </div>
              </div>

              <div class="flex items-center gap-3 shrink-0">
                <Trophy class="h-5 w-5 opacity-60" :style="{ color: 'var(--contest-accent)' }" />
                <button
                  type="button"
                  class="contest-action-btn"
                  :style="{ '--btn-color': 'var(--contest-accent)' }"
                  @click.stop="openContest(contest)"
                >
                  {{ getContestActionLabel(contest.status) }}
                </button>
              </div>
            </div>
          </article>
        </div>
      </div>
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-metric {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.journal-metric:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 35%, transparent);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.08);
}

.journal-log {
  background: var(--journal-surface);
  border-color: var(--journal-border);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.625rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--journal-muted);
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-ready {
  background: #10b981;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
  animation: pulse-dot 2s infinite;
}

.tech-font {
  font-family: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
}

@keyframes pulse-dot {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.contest-board {
  border-top: 1px dashed rgba(148, 163, 184, 0.72);
}

.contest-list {
  position: relative;
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.62);
  overflow: hidden;
}

.contest-item {
  position: relative;
  border-bottom: 1px dashed rgba(148, 163, 184, 0.56);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.84), rgba(248, 250, 252, 0.76));
  transition:
    border-color 180ms ease,
    background 180ms ease;
}

.contest-item:last-child {
  border-bottom: 0;
}

.contest-item:hover,
.contest-item:focus-visible {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.88));
  outline: none;
}

.contest-badge {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--contest-accent) 30%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 10%, transparent);
  padding: 0.18rem 0.6rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: color-mix(in srgb, var(--contest-accent) 80%, var(--journal-ink));
}

.contest-mode {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid var(--journal-border);
  background: rgba(226, 232, 240, 0.4);
  padding: 0.18rem 0.6rem;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.contest-action-btn {
  border-radius: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--btn-color, var(--journal-accent)) 36%, transparent);
  background: color-mix(in srgb, var(--btn-color, var(--journal-accent)) 10%, transparent);
  padding: 0.4rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--btn-color, var(--journal-accent)) 90%, var(--journal-ink));
  cursor: pointer;
  transition:
    background 150ms ease,
    border-color 150ms ease;
}

.contest-action-btn:hover {
  background: color-mix(in srgb, var(--btn-color, var(--journal-accent)) 18%, transparent);
  border-color: color-mix(in srgb, var(--btn-color, var(--journal-accent)) 60%, transparent);
}

.contest-btn {
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.4rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--journal-ink);
  cursor: pointer;
  transition: border-color 150ms ease;
}

.contest-btn:hover {
  border-color: var(--journal-accent);
}

:deep(.contest-empty-state) {
  border-top-style: dashed;
  border-bottom-style: dashed;
  border-top-color: rgba(148, 163, 184, 0.58);
  border-bottom-color: rgba(148, 163, 184, 0.58);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f1f5f9;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.85);
  --journal-surface-subtle: rgba(30, 41, 59, 0.6);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(99, 102, 241, 0.14), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}
</style>
