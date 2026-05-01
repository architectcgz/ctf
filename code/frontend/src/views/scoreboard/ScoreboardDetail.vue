<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, BarChart2, RefreshCw, Shield } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import ScoreboardRealtimeBridge from '@/components/scoreboard/ScoreboardRealtimeBridge.vue'
import { getScoreboard } from '@/api/contest'
import type { ContestScoreboardData, ContestStatus } from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { getContestAccentColor, getStatusLabel } from '@/utils/contest'

const route = useRoute()
const toast = useToast()

const scoreboard = ref<ContestScoreboardData | null>(null)
const loading = ref(false)
const refreshing = ref(false)
const error = ref(false)
let requestToken = 0

const contestId = computed(() => String(route.params.contestId ?? ''))
const rows = computed(() => scoreboard.value?.scoreboard.list ?? [])
const contest = computed(() => scoreboard.value?.contest)
const supportsRealtime = computed(() => {
  const status = contest.value?.status
  return status === 'running' || status === 'frozen'
})
const accentStyle = computed<Record<string, string>>(() => ({
  '--scoreboard-accent': getContestAccentColor(contest.value?.status ?? 'ended'),
}))
const emptyTitle = computed(() => (error.value ? '排行榜加载失败' : '暂无排行榜数据'))
const emptyDescription = computed(() =>
  error.value ? '该竞赛排行榜暂时不可用，请稍后重新加载。' : '当前还没有队伍进入榜单。'
)
const topScore = computed(() => rows.value[0]?.score ?? 0)
const solvedCount = computed(() => rows.value.reduce((sum, row) => sum + row.solved_count, 0))

function formatDateTime(value?: string): string {
  if (!value) return '未记录'
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function formatContestWindow(payload?: ContestScoreboardData['contest']): string {
  if (!payload) return '未记录'
  return `${formatDateTime(payload.started_at)} ~ ${formatDateTime(payload.ends_at)}`
}

function getRowClass(rank: number): string {
  if (rank === 1) return 'sb-row sb-row--top1'
  if (rank === 2) return 'sb-row sb-row--top2'
  if (rank === 3) return 'sb-row sb-row--top3'
  return 'sb-row'
}

function getRankPillClass(rank: number): string[] {
  return [
    'sb-rank-pill',
    rank === 1 ? 'sb-rank-pill--top1' : '',
    rank === 2 ? 'sb-rank-pill--top2' : '',
    rank === 3 ? 'sb-rank-pill--top3' : '',
  ]
}

function getStatusCopy(status?: ContestStatus, frozen?: boolean): string {
  if (frozen || status === 'frozen') {
    return '封榜阶段展示冻结前排名，后续解封后同步最终成绩。'
  }
  if (status === 'running') {
    return '进行中竞赛支持实时更新，也可以手动刷新最新排名。'
  }
  return '历史竞赛展示最终成绩，用于复盘队伍表现。'
}

async function loadScoreboard(silent = false): Promise<void> {
  const currentContestId = contestId.value
  if (!currentContestId) {
    return
  }

  const token = ++requestToken
  const hadScoreboard = Boolean(scoreboard.value)
  if (silent) {
    refreshing.value = true
  } else {
    loading.value = true
  }
  error.value = false

  try {
    const payload = await getScoreboard(currentContestId, { page: 1, page_size: 100 })
    if (token !== requestToken) {
      return
    }
    scoreboard.value = payload
  } catch {
    if (token !== requestToken) {
      return
    }
    if (silent && hadScoreboard) {
      toast.error('排行榜刷新失败')
      return
    }
    scoreboard.value = null
    error.value = true
  } finally {
    if (token === requestToken) {
      loading.value = false
      refreshing.value = false
    }
  }
}

watch(
  contestId,
  () => {
    void loadScoreboard()
  },
  { immediate: true }
)
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
    :style="accentStyle"
  >
    <main class="content-pane">
      <div class="scoreboard-detail-page">
        <router-link
          class="scoreboard-back-link"
          :to="{ name: 'Scoreboard' }"
        >
          <ArrowLeft class="h-4 w-4" />
          返回排行列表
        </router-link>

        <header class="scoreboard-detail-hero">
          <div>
            <div class="workspace-overline scoreboard-panel-overline">
              Contest Scoreboard
            </div>
            <h1 class="scoreboard-detail-title workspace-page-title">
              {{ contest?.title || '竞赛排行榜' }}
            </h1>
            <p class="scoreboard-detail-subtitle">
              {{ getStatusCopy(contest?.status, scoreboard?.frozen) }}
            </p>
          </div>
          <button
            type="button"
            class="ui-btn ui-btn--secondary"
            :disabled="loading || refreshing"
            @click="loadScoreboard(true)"
          >
            <RefreshCw
              class="h-4 w-4"
              :class="{ 'animate-spin': refreshing }"
            />
            刷新
          </button>
        </header>

        <ScoreboardRealtimeBridge
          v-if="contest && supportsRealtime"
          :contest-id="contest.id"
          @updated="loadScoreboard(true)"
        />

        <section class="scoreboard-summary">
          <div class="scoreboard-summary-title">
            <BarChart2 class="h-4 w-4" />
            <span>排行概况</span>
          </div>
          <div class="scoreboard-summary-grid metric-panel-grid">
            <div class="scoreboard-summary-item metric-panel-card">
              <div class="scoreboard-summary-label metric-panel-label">
                榜单队伍
              </div>
              <div class="scoreboard-summary-value metric-panel-value">
                {{ rows.length }}
              </div>
              <div class="scoreboard-summary-helper metric-panel-helper">
                当前进入排行榜的队伍数量
              </div>
            </div>
            <div class="scoreboard-summary-item metric-panel-card">
              <div class="scoreboard-summary-label metric-panel-label">
                最高分
              </div>
              <div class="scoreboard-summary-value metric-panel-value">
                {{ topScore }}
              </div>
              <div class="scoreboard-summary-helper metric-panel-helper">
                当前榜首队伍分数
              </div>
            </div>
            <div class="scoreboard-summary-item metric-panel-card">
              <div class="scoreboard-summary-label metric-panel-label">
                总解题
              </div>
              <div class="scoreboard-summary-value metric-panel-value">
                {{ solvedCount }}
              </div>
              <div class="scoreboard-summary-helper metric-panel-helper">
                榜单队伍累计解题数
              </div>
            </div>
            <div class="scoreboard-summary-item metric-panel-card">
              <div class="scoreboard-summary-label metric-panel-label">
                榜单状态
              </div>
              <div class="scoreboard-summary-value metric-panel-value">
                {{ scoreboard?.frozen ? '封榜' : getStatusLabel(contest?.status ?? 'ended') }}
              </div>
              <div class="scoreboard-summary-helper metric-panel-helper">
                {{ formatContestWindow(contest) }}
              </div>
            </div>
          </div>
        </section>

        <div
          v-if="loading && !scoreboard"
          class="scoreboard-loading"
        >
          <div class="scoreboard-loading-spinner" />
        </div>

        <AppEmpty
          v-else-if="!rows.length"
          class="scoreboard-empty-state"
          icon="Trophy"
          :title="emptyTitle"
          :description="emptyDescription"
        >
          <template #action>
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="loadScoreboard(true)"
            >
              重新加载
            </button>
          </template>
        </AppEmpty>

        <section
          v-else
          class="scoreboard-directory"
          aria-label="排行详情"
        >
          <div class="scoreboard-directory-top">
            <h2 class="scoreboard-directory-title">
              排行详情
            </h2>
            <div class="scoreboard-directory-meta">
              展示前 {{ rows.length }} 支队伍
              <span
                v-if="scoreboard?.frozen"
                class="scoreboard-frozen-inline"
              >
                <Shield class="h-3 w-3" /> 已冻结
              </span>
            </div>
          </div>

          <div class="scoreboard-table-shell overflow-x-auto">
            <table class="sb-table">
              <thead>
                <tr>
                  <th>排名</th>
                  <th>队伍</th>
                  <th>得分</th>
                  <th>解题数</th>
                  <th>最近得分</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in rows"
                  :key="item.team_id"
                  data-testid="scoreboard-detail-row"
                  :class="getRowClass(item.rank)"
                >
                  <td class="sb-cell--rank">
                    <span :class="getRankPillClass(item.rank)">{{ item.rank }}</span>
                  </td>
                  <td>{{ item.team_name }}</td>
                  <td class="sb-cell--mono">
                    {{ item.score }}
                  </td>
                  <td>{{ item.solved_count }}</td>
                  <td class="sb-cell--muted">
                    {{ formatDateTime(item.last_submission_at) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
}

.scoreboard-detail-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.scoreboard-back-link {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: var(--radius-md);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--journal-muted);
}

.scoreboard-back-link:hover,
.scoreboard-back-link:focus-visible {
  color: var(--journal-accent);
  border-color: color-mix(in srgb, var(--journal-accent) 36%, var(--journal-border));
}

.scoreboard-detail-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-top: var(--space-5);
  padding-bottom: var(--space-5);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.scoreboard-panel-overline {
  margin-bottom: var(--space-2);
}

.scoreboard-detail-title {
  max-width: 760px;
}

.scoreboard-detail-subtitle {
  max-width: 720px;
  margin-top: var(--space-3);
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.scoreboard-summary {
  margin-top: var(--space-6);
}

.scoreboard-summary-title {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-3);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--journal-ink);
}

.scoreboard-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-16) 0;
}

.scoreboard-loading-spinner {
  width: 2rem;
  height: 2rem;
  border: 0.25rem solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: scoreboardSpin 900ms linear infinite;
}

:deep(.scoreboard-empty-state) {
  margin-top: var(--space-6);
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.scoreboard-directory {
  margin-top: var(--space-6);
}

.scoreboard-frozen-inline {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  margin-left: var(--space-2);
  color: var(--scoreboard-accent, var(--journal-accent));
}

.scoreboard-table-shell {
  overflow-x: auto;
}

.sb-rank-pill {
  display: inline-flex;
  align-items: center;
  min-height: 1.625rem;
  padding: 0 var(--space-2);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-12);
  font-weight: 600;
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.sb-table {
  width: 100%;
  border-collapse: collapse;
}

.sb-table th {
  padding: 0 0 var(--space-3);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  text-align: left;
  color: var(--journal-muted);
}

.sb-row td {
  padding: var(--space-4) 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  font-size: var(--font-size-14);
  color: var(--journal-ink);
}

.sb-row--top1 td,
.sb-rank-pill--top1 {
  color: color-mix(in srgb, var(--color-warning) 84%, var(--journal-ink));
}

.sb-row--top2 td,
.sb-rank-pill--top2 {
  color: color-mix(in srgb, var(--color-text-secondary) 80%, var(--journal-ink));
}

.sb-row--top3 td,
.sb-rank-pill--top3 {
  color: color-mix(in srgb, var(--color-danger) 42%, var(--color-warning));
}

.sb-cell--rank,
.sb-cell--mono {
  font-family: var(--font-family-mono);
}

.sb-cell--muted {
  color: var(--journal-muted);
}

@keyframes scoreboardSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}
</style>
