<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { CalendarRange, Clock3, Trophy } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getContests } from '@/api/contest'
import type { ContestListItem, ContestStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
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

function contestItemStyle(status: ContestStatus): Record<string, string> {
  return {
    '--contest-accent': getContestAccentColor(status),
  }
}
</script>

<template>
  <div class="contest-center space-y-6">
    <PageHeader
      eyebrow="Contest Center"
      title="竞赛中心"
      description="统一查看所有竞赛窗口，快速识别正在进行、可报名和已结束的场次。"
    />

    <section class="contest-overview">
      <div class="contest-overview__lead">
        <div class="contest-overview__kicker">
          Operations Snapshot
        </div>
        <h2 class="contest-overview__title">
          本周竞赛态势面板
        </h2>
        <p class="contest-overview__desc">
          重点查看进行中与报名中竞赛，避免错过开赛窗口。进入详情后可直接完成报名、组队和题目攻防。
        </p>
      </div>
      <div class="contest-overview__metrics">
        <article
          v-for="item in summaryMetrics"
          :key="item.key"
          class="contest-kpi"
        >
          <div class="contest-kpi__label">
            {{ item.label }}
          </div>
          <div class="contest-kpi__value">
            {{ item.value }}
          </div>
          <div class="contest-kpi__hint">
            {{ item.hint }}
          </div>
        </article>
      </div>
    </section>

    <div
      v-if="loadErrorMessage"
      class="contest-error"
      role="alert"
      aria-live="polite"
    >
      <div class="contest-error__title">
        竞赛列表加载失败
      </div>
      <div class="contest-error__text">
        {{ loadErrorMessage }}
      </div>
      <ElButton
        class="mt-3"
        type="danger"
        @click="refresh"
      >
        重试
      </ElButton>
    </div>

    <section class="contest-list">
      <div
        v-if="loading"
        class="contest-list__skeleton"
      >
        <div
          v-for="index in 4"
          :key="index"
          class="contest-skeleton-row"
        />
      </div>

      <AppEmpty
        v-else-if="list.length === 0"
        icon="Flag"
        title="暂无竞赛"
        description="当前没有可展示的竞赛，稍后再来查看新的开赛计划。"
      />

      <div
        v-else
        class="contest-list__rows"
      >
        <article
          v-for="contest in list"
          :key="contest.id"
          class="contest-item"
          :style="contestItemStyle(contest.status)"
          tabindex="0"
          @click="openContest(contest)"
          @keydown="onKeyboardOpen($event, contest)"
        >
          <div class="contest-item__main">
            <div class="contest-item__head">
              <span class="contest-item__status">{{ getStatusLabel(contest.status) }}</span>
              <span class="contest-item__mode">{{ getModeLabel(contest.mode) }}</span>
            </div>

            <h3 class="contest-item__title">
              {{ contest.title }}
            </h3>
            <p class="contest-item__description">
              该竞赛支持组队协作与实时排行，进入详情页可查看完整规则与题目面板。
            </p>

            <div class="contest-item__meta">
              <div class="contest-item__meta-row">
                <CalendarRange class="h-4 w-4" />
                <span>{{ formatTime(contest.starts_at) }} ~ {{ formatTime(contest.ends_at) }}</span>
              </div>
              <div class="contest-item__meta-row contest-item__meta-row--strong">
                <Clock3 class="h-4 w-4" />
                <span>{{ getTimelineHint(contest) }}</span>
              </div>
            </div>
          </div>

          <div class="contest-item__action-wrap">
            <Trophy class="contest-item__icon h-5 w-5" />
            <ElButton
              type="primary"
              @click.stop="openContest(contest)"
            >
              {{ getContestActionLabel(contest.status) }}
            </ElButton>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.contest-center {
  --contest-accent: var(--color-primary);
}

.contest-overview {
  padding-bottom: 1rem;
  border-bottom: 1px solid color-mix(in srgb, var(--contest-accent) 24%, var(--color-border-default));
}

.contest-overview__lead {
  max-width: 60ch;
}

.contest-overview__kicker {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--color-text-secondary) 82%, #9be7f5);
}

.contest-overview__title {
  margin-top: 0.85rem;
  font-size: clamp(1.2rem, 2.7vw, 1.72rem);
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--color-text-primary);
}

.contest-overview__desc {
  margin-top: 0.56rem;
  font-size: 0.9rem;
  line-height: 1.75;
  color: var(--color-text-secondary);
}

.contest-overview__metrics {
  margin-top: 1rem;
  display: grid;
  gap: 0.68rem;
}

.contest-kpi {
  border-left: 2px solid color-mix(in srgb, var(--contest-accent) 42%, var(--color-border-default));
  padding: 0.1rem 0 0.1rem 0.72rem;
}

.contest-kpi__label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--color-text-secondary) 80%, #b8ecf7);
}

.contest-kpi__value {
  margin-top: 0.4rem;
  font-size: 1.55rem;
  font-weight: 700;
  line-height: 1.15;
  color: var(--color-text-primary);
}

.contest-kpi__hint {
  margin-top: 0.35rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--color-text-secondary);
}

.contest-error {
  border-left: 2px solid color-mix(in srgb, var(--color-danger) 62%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-danger) 6%, transparent);
  padding: 0.72rem 0.8rem;
}

.contest-error__title {
  font-size: 0.95rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--color-danger) 85%, var(--color-text-primary));
}

.contest-error__text {
  margin-top: 0.25rem;
  font-size: 0.85rem;
  line-height: 1.6;
  color: color-mix(in srgb, var(--color-danger) 65%, var(--color-text-primary));
}

.contest-list__skeleton {
  display: grid;
  gap: 0.7rem;
  margin-top: 0.35rem;
}

.contest-skeleton-row {
  height: 5.25rem;
  border-bottom: 1px solid var(--color-border-default);
  background: linear-gradient(
    90deg,
    transparent,
    color-mix(in srgb, var(--color-bg-surface) 88%, var(--contest-accent)),
    transparent
  );
  background-size: 220% 100%;
  animation: contestSkeletonMove 1.35s linear infinite;
}

.contest-list__rows {
  display: grid;
  margin-top: 0.1rem;
  border-top: 1px solid var(--color-border-default);
}

.contest-item {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  justify-content: space-between;
  gap: 0.9rem;
  border-bottom: 1px solid var(--color-border-default);
  border-left: 2px solid color-mix(in srgb, var(--contest-accent) 46%, transparent);
  padding: 0.95rem 0.15rem 0.95rem 0.82rem;
  cursor: pointer;
  transition: background-color 180ms ease, border-left-color 180ms ease;
}

.contest-item:hover,
.contest-item:focus-visible {
  background: color-mix(in srgb, var(--contest-accent) 5%, transparent);
  border-left-color: color-mix(in srgb, var(--contest-accent) 80%, transparent);
  outline: none;
}

.contest-item__main {
  min-width: 0;
  flex: 1 1 36rem;
}

.contest-item__head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.contest-item__status,
.contest-item__mode {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.2rem 0.56rem;
  font-size: 0.73rem;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.contest-item__status {
  border: 1px solid color-mix(in srgb, var(--contest-accent) 32%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 10%, transparent);
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--color-text-primary));
}

.contest-item__mode {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 90%, transparent);
  color: var(--color-text-secondary);
}

.contest-item__title {
  margin-top: 0.62rem;
  font-size: 1.06rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-item__description {
  margin-top: 0.45rem;
  font-size: 0.84rem;
  line-height: 1.65;
  color: var(--color-text-secondary);
}

.contest-item__meta {
  margin-top: 0.72rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 0.9rem;
}

.contest-item__meta-row {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.78rem;
  color: var(--color-text-secondary);
}

.contest-item__meta-row--strong {
  color: color-mix(in srgb, var(--contest-accent) 86%, var(--color-text-primary));
}

.contest-item__action-wrap {
  display: flex;
  min-width: 9.2rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 0.72rem;
}

.contest-item__icon {
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--color-text-primary));
  opacity: 0.72;
}

@media (min-width: 960px) {
  .contest-overview__metrics {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 639px) {
  .contest-item__action-wrap {
    width: 100%;
    min-width: 0;
    justify-content: space-between;
  }
}

:global([data-theme='light']) .contest-overview {
  border-bottom-color: color-mix(in srgb, var(--contest-accent) 20%, var(--color-border-default));
}

@keyframes contestSkeletonMove {
  from {
    background-position-x: 0%;
  }
  to {
    background-position-x: 240%;
  }
}
</style>
