<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { CalendarRange, Clock3, Trophy } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getContests } from '@/api/contest'
import type { ContestListItem, ContestStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { usePagination } from '@/composables/usePagination'
import {
  getContestAccentColor,
  getContestActionLabel,
  getModeLabel,
  isStudentVisibleContestStatus,
  getStatusLabel,
} from '@/utils/contest'

const router = useRouter()
const { list, loading, error, refresh } = usePagination(getContests)

onMounted(() => {
  void refresh()
})

interface ContestSummaryMetric {
  key: string
  label: string
  value: number
  hint: string
}

const visibleContests = computed(() =>
  list.value.filter((contest) => isStudentVisibleContestStatus(contest.status))
)

const summaryMetrics = computed<ContestSummaryMetric[]>(() => {
  const runningCount = visibleContests.value.filter((contest) => contest.status === 'running').length
  const registeringCount = visibleContests.value.filter((contest) => contest.status === 'registering').length
  const endedCount = visibleContests.value.filter((contest) =>
    ['ended', 'cancelled', 'archived', 'frozen'].includes(contest.status)
  ).length

  return [
    {
      key: 'total',
      label: '竞赛总数',
      value: visibleContests.value.length,
      hint: '当前可查看的竞赛数量',
    },
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
  void router.push(`/contests/${id}`)
}

function openContest(contest: ContestListItem): void {
  goToDetail(contest.id)
}

function contestAccentStyle(status: ContestStatus): Record<string, string> {
  return { '--contest-row-accent': getContestAccentColor(status) }
}
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div class="contest-page">
      <header class="contest-topbar">
        <div class="contest-heading">
          <div class="workspace-overline">Contests</div>
          <h1 class="contest-title workspace-page-title">竞赛中心</h1>
          <p class="contest-subtitle">查看当前可参加和已结束的竞赛，直接进入竞赛工作区。</p>
        </div>
      </header>

      <section class="contest-summary">
        <div class="contest-summary-title">
          <Trophy class="h-4 w-4" />
          <span>当前竞赛概况</span>
        </div>
        <div class="contest-summary-grid metric-panel-grid">
          <div
            v-for="stat in summaryMetrics"
            :key="stat.key"
            class="contest-summary-item metric-panel-card"
          >
            <div class="contest-summary-label metric-panel-label">{{ stat.label }}</div>
            <div class="contest-summary-value metric-panel-value">{{ stat.value }}</div>
            <div class="contest-summary-helper metric-panel-helper">{{ stat.hint }}</div>
          </div>
        </div>
      </section>

      <div v-if="loading" class="contest-loading">
        <div class="contest-loading-spinner" />
      </div>

      <AppEmpty
        v-else-if="loadErrorMessage"
        class="contest-empty-state"
        icon="AlertTriangle"
        title="竞赛列表加载失败"
        :description="loadErrorMessage"
      >
        <template #action>
          <button type="button" class="ui-btn ui-btn--secondary" @click="refresh">重试</button>
        </template>
      </AppEmpty>

      <AppEmpty
        v-else-if="visibleContests.length === 0"
        class="contest-empty-state"
        icon="Flag"
        title="暂无竞赛"
        description="当前没有可展示的竞赛，稍后再来查看新的开赛计划。"
      />

      <section v-else class="contest-directory" aria-label="竞赛目录">
        <div class="contest-directory-top">
          <h2 class="contest-directory-title">竞赛列表</h2>
          <div class="contest-directory-meta">共 {{ visibleContests.length }} 场</div>
        </div>

        <div class="contest-directory-head">
          <span>竞赛</span>
          <span>时间</span>
          <span>状态</span>
          <span>节奏</span>
          <span>操作</span>
        </div>

        <button
          v-for="contest in visibleContests"
          :key="contest.id"
          type="button"
          class="contest-row"
          :style="contestAccentStyle(contest.status)"
          :aria-label="`${contest.title}，${getStatusLabel(contest.status)}，${getModeLabel(contest.mode)}`"
          @click="openContest(contest)"
        >
          <div class="contest-row-main">
            <div class="contest-row-status-strip">
              <span
                class="contest-chip"
                :style="{ '--contest-chip-color': 'var(--contest-row-accent)' }"
              >
                {{ getStatusLabel(contest.status) }}
              </span>
              <span class="contest-chip contest-chip-muted">{{ getModeLabel(contest.mode) }}</span>
            </div>
            <h3 class="contest-row-title" :title="contest.title">{{ contest.title }}</h3>
          </div>

          <div class="contest-row-time">
            <div class="contest-row-time-item">
              <CalendarRange class="h-3.5 w-3.5" />
              <span>{{ formatTime(contest.starts_at) }} - {{ formatTime(contest.ends_at) }}</span>
            </div>
          </div>

          <div class="contest-row-state">
            <span
              class="contest-state-chip"
              :style="{ '--contest-state-color': 'var(--contest-row-accent)' }"
            >
              {{ getStatusLabel(contest.status) }}
            </span>
          </div>

          <div class="contest-row-timeline">
            <div class="contest-row-time-item contest-row-time-item-strong">
              <Clock3 class="h-3.5 w-3.5" />
              <span>{{ getTimelineHint(contest) }}</span>
            </div>
          </div>

          <div class="contest-row-cta">
            <span>{{ getContestActionLabel(contest.status) }}</span>
          </div>
        </button>
      </section>
      </div>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
}

.contest-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.contest-subtitle {
  max-width: 680px;
}

.contest-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
}

.contest-loading-spinner {
  width: 32px;
  height: 32px;
  border: 4px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: contestSpin 900ms linear infinite;
}

:deep(.contest-empty-state) {
  margin-top: 24px;
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.contest-directory {
  margin-top: 24px;
}

.contest-directory-head {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(220px, 1fr) 120px 180px 120px;
  gap: 16px;
  padding: 0 0 12px;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.contest-row {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(220px, 1fr) 120px 180px 120px;
  gap: 16px;
  align-items: center;
  width: 100%;
  padding: 18px 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.contest-row:hover,
.contest-row:focus-visible {
  background: color-mix(in srgb, var(--contest-row-accent, var(--journal-accent)) 5%, transparent);
  box-shadow: inset 2px 0 0
    color-mix(in srgb, var(--contest-row-accent, var(--journal-accent)) 64%, transparent);
  outline: none;
}

.contest-row-main {
  min-width: 0;
}

.contest-row-status-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.contest-row-title {
  margin-top: 10px;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-18);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contest-chip,
.contest-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: 8px;
  font-size: var(--font-size-12);
  font-weight: 600;
}

.contest-chip {
  background: color-mix(in srgb, var(--contest-chip-color, var(--journal-accent)) 12%, transparent);
  color: var(--contest-chip-color, var(--journal-accent));
}

.contest-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.contest-state-chip {
  background: color-mix(
    in srgb,
    var(--contest-state-color, var(--journal-accent)) 12%,
    transparent
  );
  color: var(--contest-state-color, var(--journal-accent));
}

.contest-row-time,
.contest-row-timeline {
  font-size: var(--font-size-13);
  line-height: 1.5;
  color: var(--journal-muted);
}

.contest-row-time-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.contest-row-time-item-strong {
  color: var(--contest-row-accent, var(--journal-accent));
  font-weight: 600;
}

.contest-row-cta {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--contest-row-accent, var(--journal-accent));
}

@keyframes contestSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1180px) {
  .contest-directory-head {
    display: none;
  }

  .contest-row {
    grid-template-columns: 1fr;
  }
}
</style>
