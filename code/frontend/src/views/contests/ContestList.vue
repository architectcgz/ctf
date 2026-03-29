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
  <div class="journal-shell space-y-6">
    <!-- Hero -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div>
          <div class="journal-eyebrow">Contest Center</div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            竞赛中心
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            统一查看所有竞赛窗口，快速识别正在进行、可报名和已结束的场次。
          </p>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <Flame class="h-5 w-5 text-[var(--journal-accent)]" />
            竞赛态势
          </div>
          <div class="mt-4 grid grid-cols-2 gap-3">
            <div v-for="stat in summaryMetrics" :key="stat.key" class="journal-note">
              <div class="journal-note-label">{{ stat.label }}</div>
              <div class="journal-note-value">{{ stat.value }}</div>
            </div>
          </div>
        </article>
      </div>
    </section>

    <!-- 加载错误 -->
    <div v-if="loadErrorMessage" class="journal-panel rounded-[24px] border px-5 py-5" role="alert">
      <div class="text-sm font-semibold text-[var(--journal-ink)]">竞赛列表加载失败</div>
      <div class="mt-1 text-sm text-[var(--journal-muted)]">{{ loadErrorMessage }}</div>
      <button type="button" class="contest-btn mt-3" @click="refresh">重试</button>
    </div>

    <!-- 加载中骨架 -->
    <div v-if="loading" class="space-y-3">
      <div
        v-for="i in 4"
        :key="i"
        class="h-24 rounded-[20px] animate-pulse"
        style="background: rgba(226,232,240,0.5)"
      />
    </div>

    <!-- 空状态 -->
    <AppEmpty
      v-else-if="list.length === 0 && !loadErrorMessage"
      icon="Flag"
      title="暂无竞赛"
      description="当前没有可展示的竞赛，稍后再来查看新的开赛计划。"
    />

    <!-- 竞赛列表 -->
    <section v-else class="space-y-3">
      <article
        v-for="contest in list"
        :key="contest.id"
        class="contest-item journal-log rounded-[22px] border px-5 py-5 cursor-pointer"
        :style="contestAccentStyle(contest.status)"
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
            <div class="mt-2 flex flex-wrap gap-4">
              <div class="flex items-center gap-1.5 text-xs text-[var(--journal-muted)]">
                <CalendarRange class="h-3.5 w-3.5" />
                <span>{{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}</span>
              </div>
              <div class="flex items-center gap-1.5 text-xs font-medium" :style="{ color: 'var(--contest-accent)' }">
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
    </section>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: #ffffff;
  --journal-surface-subtle: rgba(248, 250, 252, 0.9);
  --journal-accent: var(--color-primary);
}

.journal-hero {
  background:
    radial-gradient(circle at top right, rgba(99, 102, 241, 0.08), transparent 18rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.95), rgba(241, 245, 249, 0.9));
  border-color: var(--journal-border);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
}

.journal-panel {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
}

.journal-log {
  background: var(--journal-surface);
  border-color: var(--journal-border);
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.22);
  background: rgba(99, 102, 241, 0.07);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.65rem 0.85rem;
}

.journal-note-label {
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.4rem;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-item {
  transition: border-color 180ms ease, box-shadow 180ms ease;
}

.contest-item:hover,
.contest-item:focus-visible {
  border-color: var(--contest-accent, var(--journal-accent)) !important;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--contest-accent, var(--journal-accent)) 12%, transparent);
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
  transition: background 150ms ease, border-color 150ms ease;
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
