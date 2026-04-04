<script setup lang="ts">
import { computed } from 'vue'
import {
  Activity,
  ArrowRight,
  BellRing,
  MapPinned,
  Sparkles,
  Trophy,
} from 'lucide-vue-next'

import RadarChart from '@/components/charts/RadarChart.vue'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { formatDate } from '@/utils/format'

import type { StudentOverviewProps } from './overviewProps'
import { timelineSummary } from './utils'

const props = defineProps<StudentOverviewProps>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()

const quickRecommendations = computed(() => props.recommendations.slice(0, 3))
const recentTimeline = computed(() => props.timeline.slice(0, 4))
const storyMetrics = computed(() => [
  { label: '总得分', value: props.progress.total_score ?? 0, tone: 'default' },
  { label: '已解题数', value: props.progress.total_solved ?? 0, tone: 'success' },
  { label: '当前排名', value: `#${props.progress.rank ?? '-'}`, tone: 'accent' },
  { label: '完成率', value: `${props.completionRate}%`, tone: 'accent' },
])
const radarIndicators = computed(() =>
  props.skillDimensions.map((item) => ({ name: item.name, max: 100 }))
)
const radarValues = computed(() => props.skillDimensions.map((item) => item.value))
const rankSummary = computed(() => props.progress.rank ?? '-')
const operationsSummary = computed(() => [
  {
    label: '环境状态',
    value: quickRecommendations.value.length > 0 ? '可训练' : '空闲',
    description:
      quickRecommendations.value.length > 0 ? '存在可立即进入的推荐题目' : '当前没有推荐训练任务',
    status: quickRecommendations.value.length > 0 ? 'ready' : 'idle',
    icon: Activity,
  },
  {
    label: '能力分布',
    value: props.skillDimensions.length > 0 ? `${props.skillDimensions.length} 维` : '未生成',
    description:
      props.skillDimensions.length > 0 ? '基于当前训练数据实时更新' : '完成更多题目后将自动生成',
    status: props.skillDimensions.length > 0 ? 'ready' : 'idle',
    icon: MapPinned,
  },
  {
    label: '训练提示',
    value: props.weakDimensions[0] || '保持节奏',
    description:
      props.weakDimensions.length > 0
        ? `优先补强 ${props.weakDimensions.join(' / ')}`
        : '当前结构比较均衡，继续推进即可',
    status: props.weakDimensions.length > 0 ? 'warning' : 'ready',
    icon: BellRing,
  },
])

const categoryThemeMap: Record<string, string> = {
  web: 'tag-web',
  pwn: 'tag-pwn',
  reverse: 'tag-reverse',
  crypto: 'tag-crypto',
  misc: 'tag-misc',
  forensics: 'tag-forensics',
}

function categoryClass(category: string): string {
  return categoryThemeMap[category] || 'tag-default'
}

function recommendationStatus(itemId: string): string {
  return Number(itemId) % 2 === 0 ? 'ready' : 'idle'
}

function timelineStatus(eventType: string): string {
  if (eventType === 'solve') return 'solved'
  if (eventType === 'instance' || eventType.includes('instance')) return 'ready'
  return 'idle'
}
</script>

<template>
  <section class="journal-shell space-y-6 journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div>
        <div class="journal-eyebrow">Training Journal</div>
        <h2
          class="mt-3 max-w-3xl text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
        >
          {{ displayName }} 的极简训练面板
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          这里汇总了训练进度、推荐题目和最近动态。
        </p>

        <div class="mt-6 flex flex-wrap gap-3">
          <ElButton type="primary" @click="emit('openChallenges')">继续训练</ElButton>
          <ElButton plain @click="emit('openSkillProfile')">查看能力画像</ElButton>
        </div>
      </div>
      <div class="journal-board">
        <section class="journal-bento">
          <article class="journal-panel journal-radar-card px-6 py-6">
            <div class="flex items-center justify-between gap-4">
              <div>
                <div class="journal-eyebrow">Skill Matrix</div>
                <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">能力雷达</h3>
              </div>
              <MapPinned class="h-5 w-5 text-[var(--journal-accent-strong)]" />
            </div>
            <div v-if="skillDimensions.length > 0" class="mt-4">
              <RadarChart :indicators="radarIndicators" :values="radarValues" name="能力值" />
            </div>
            <div
              v-else
              class="mt-6 rounded-[18px] border border-dashed border-[var(--journal-soft-border)] bg-[var(--journal-surface-subtle)]/52 px-4 py-10 text-center text-sm text-[var(--journal-muted)]"
            >
              当前能力数据不足，完成更多题目后将生成雷达图。
            </div>
          </article>

          <article class="journal-panel journal-rank-card px-6 py-6">
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="journal-eyebrow">Leaderboard</div>
                <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">竞技表现</h3>
              </div>
              <Trophy class="h-5 w-5 text-[var(--journal-accent-strong)]" />
            </div>
            <div class="mt-6 grid gap-3 md:grid-cols-2">
              <article
                v-for="item in storyMetrics"
                :key="item.label"
                class="journal-metric px-4 py-4"
                :class="item.tone === 'accent' ? 'journal-metric-accent' : ''"
              >
                <div
                  class="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--journal-muted)]"
                >
                  {{ item.label }}
                </div>
                <div
                  class="mt-3 text-[30px] font-semibold tracking-tight text-[var(--journal-ink)]"
                >
                  {{ item.value }}
                </div>
              </article>
            </div>
            <div class="journal-rank-summary mt-5 px-4 py-4">
              <div class="flex items-center gap-2 text-sm text-[var(--journal-muted)]">
                <span class="status-dot status-dot-solved" />
                当前排名
              </div>
              <div class="mt-2 tech-font text-2xl font-semibold text-[var(--journal-ink)]">
                #{{ rankSummary }}
              </div>
            </div>
          </article>

          <article class="journal-panel journal-ops-card px-6 py-6">
            <div class="flex items-center justify-between gap-4">
              <div>
                <div class="journal-eyebrow">Operations</div>
                <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">公告与状态</h3>
              </div>
              <BellRing class="h-5 w-5 text-[var(--journal-accent-strong)]" />
            </div>
            <div class="mt-5 space-y-3">
              <article
                v-for="item in operationsSummary"
                :key="item.label"
                class="journal-inline-item px-4 py-4"
              >
                <div class="flex items-center justify-between gap-3">
                  <div class="flex items-center gap-3">
                    <component
                      :is="item.icon"
                      class="h-4 w-4 text-[var(--journal-accent-strong)]"
                    />
                    <div class="text-sm font-medium text-[var(--journal-ink)]">
                      {{ item.label }}
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="status-dot" :class="`status-dot-${item.status}`" />
                    <span class="tech-font text-sm font-medium text-[var(--journal-ink)]">{{
                      item.value
                    }}</span>
                  </div>
                </div>
                <div class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                  {{ item.description }}
                </div>
              </article>
            </div>
          </article>

          <article class="journal-panel journal-recommend-card px-6 py-6">
            <div class="flex items-center justify-between gap-4">
              <div>
                <div class="journal-eyebrow">Recommended Track</div>
                <h3 class="mt-2 text-2xl font-semibold text-[var(--journal-ink)]">推荐训练队列</h3>
              </div>
              <Sparkles class="hidden h-10 w-10 text-[var(--journal-accent)] md:block" />
            </div>

            <div
              v-if="quickRecommendations.length === 0"
              class="mt-6 rounded-[22px] border border-dashed border-[var(--journal-soft-border)] bg-[var(--journal-surface-subtle)]/52 px-4 py-10 text-center text-sm text-[var(--journal-muted)]"
            >
              当前没有推荐题目，直接去挑战列表挑一道新题即可。
            </div>

            <div v-else class="mt-6 grid gap-3">
              <button
                v-for="item in quickRecommendations"
                :key="item.challenge_id"
                type="button"
                class="journal-rec-item flex w-full items-start gap-4 px-4 py-4 text-left transition"
                @click="emit('openChallenge', item.challenge_id)"
              >
                <div
                  class="journal-id-badge mt-0.5 flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl text-sm font-semibold tech-font"
                >
                  #{{ item.challenge_id }}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div class="space-y-2">
                      <div class="flex flex-wrap items-center gap-2">
                        <span
                          class="status-dot"
                          :class="`status-dot-${recommendationStatus(item.challenge_id)}`"
                        />
                        <div class="text-base font-semibold text-[var(--journal-ink)]">
                          {{ item.title }}
                        </div>
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="category-chip" :class="categoryClass(item.category)">{{
                          item.category.toUpperCase()
                        }}</span>
                        <span
                          class="rounded-full px-2.5 py-1 text-xs font-medium"
                          :class="difficultyClass(item.difficulty)"
                        >
                          <span class="difficulty-dot" :class="`difficulty-${item.difficulty}`" />
                          {{ difficultyLabel(item.difficulty) }}
                        </span>
                      </div>
                    </div>
                    <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-[var(--journal-accent-strong)]" />
                  </div>
                  <p class="mt-3 text-sm leading-6 text-[var(--journal-muted)]">
                    {{ item.reason }}
                  </p>
                </div>
              </button>
            </div>
          </article>

          <article class="journal-panel journal-timeline-card px-6 py-6">
            <div class="journal-eyebrow">Recent Notes</div>
            <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">训练记录</h3>

            <div
              v-if="recentTimeline.length === 0"
              class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-soft-border)] bg-[var(--journal-surface-subtle)]/52 px-4 py-10 text-center text-sm text-[var(--journal-muted)]"
            >
              当前还没有训练动态。
            </div>

            <div v-else class="mt-5 space-y-3">
              <article
                v-for="event in recentTimeline"
                :key="event.id"
                class="journal-log px-4 py-4"
              >
                <div class="flex items-center justify-between gap-3">
                  <div class="flex min-w-0 items-center gap-3">
                    <span
                      class="status-dot shrink-0"
                      :class="`status-dot-${timelineStatus(event.type)}`"
                    />
                    <div class="truncate text-sm font-medium text-[var(--journal-ink)]">
                      {{ event.title }}
                    </div>
                  </div>
                  <div class="tech-font text-xs text-[var(--journal-muted)]">
                    {{ formatDate(event.created_at) }}
                  </div>
                </div>
                <div class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                  {{ timelineSummary(event) }}
                </div>
              </article>
            </div>
          </article>
        </section>
      </div>
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--journal-border) 68%, transparent);
  --journal-divider: color-mix(in srgb, var(--journal-border) 56%, transparent);
  --journal-control-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-board {
  margin-top: 1.5rem;
  border-top: 1px dashed var(--journal-divider);
  padding-top: 1.25rem;
}

.journal-panel {
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.journal-metric,
.journal-rec-item,
.journal-log,
.journal-inline-item,
.journal-rank-summary {
  border: 1px solid var(--journal-shell-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  box-shadow: none;
}

.journal-id-badge {
  border: 1px solid var(--journal-soft-border);
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base));
  color: var(--journal-muted);
}

.journal-eyebrow {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.26em;
  text-transform: uppercase;
  color: #64748b;
}

.journal-bento {
  display: grid;
  gap: 1.25rem;
}

@media (min-width: 1280px) {
  .journal-bento {
    grid-template-columns: 1.1fr 0.92fr 0.88fr;
    grid-template-areas:
      'radar rank ops'
      'recommend recommend timeline';
  }

  .journal-radar-card {
    grid-area: radar;
    position: relative;
    padding-right: 1.5rem;
  }

  .journal-radar-card::after {
    content: '';
    position: absolute;
    top: 0.5rem;
    right: -0.625rem;
    bottom: 0.5rem;
    border-right: 1px dashed var(--journal-divider);
  }
  .journal-rank-card {
    grid-area: rank;
    position: relative;
    padding-right: 1.5rem;
  }

  .journal-rank-card::after {
    content: '';
    position: absolute;
    top: 0.5rem;
    right: -0.625rem;
    bottom: 0.5rem;
    border-right: 1px dashed var(--journal-divider);
  }
  .journal-ops-card {
    grid-area: ops;
  }
  .journal-recommend-card {
    grid-area: recommend;
    position: relative;
    padding-right: 1.5rem;
    padding-top: 1.75rem;
  }

  .journal-recommend-card::before {
    content: '';
    position: absolute;
    top: -0.625rem;
    left: 0;
    right: 0.625rem;
    border-top: 1px dashed var(--journal-divider);
  }

  .journal-recommend-card::after {
    content: '';
    position: absolute;
    top: 0.5rem;
    right: -0.625rem;
    bottom: 0.5rem;
    border-right: 1px dashed var(--journal-divider);
  }
  .journal-timeline-card {
    grid-area: timeline;
    position: relative;
    padding-top: 1.75rem;
  }

  .journal-timeline-card::before {
    content: '';
    position: absolute;
    top: -0.625rem;
    left: 0;
    right: 0;
    border-top: 1px dashed var(--journal-divider);
  }
}

.journal-metric-accent {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, #6366f1 14%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base))
  );
}

.journal-inline-item + .journal-inline-item {
  margin-top: 0.75rem;
}

.journal-rec-item:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.journal-log {
  transition: all 0.2s ease-in-out;
}

.journal-log:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.category-chip {
  display: inline-flex;
  align-items: center;
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.28rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.04em;
}

.tag-web {
  border-color: color-mix(in srgb, #60a5fa 18%, transparent);
  background: color-mix(in srgb, #60a5fa 11%, transparent);
  color: color-mix(in srgb, #60a5fa 78%, var(--journal-ink));
}
.tag-pwn {
  border-color: color-mix(in srgb, #8b5cf6 18%, transparent);
  background: color-mix(in srgb, #8b5cf6 11%, transparent);
  color: color-mix(in srgb, #8b5cf6 78%, var(--journal-ink));
}
.tag-reverse {
  border-color: color-mix(in srgb, #f87171 18%, transparent);
  background: color-mix(in srgb, #f87171 11%, transparent);
  color: color-mix(in srgb, #f87171 76%, var(--journal-ink));
}
.tag-crypto {
  border-color: color-mix(in srgb, #34d399 18%, transparent);
  background: color-mix(in srgb, #34d399 11%, transparent);
  color: color-mix(in srgb, #34d399 76%, var(--journal-ink));
}
.tag-misc {
  border-color: color-mix(in srgb, #fbbf24 18%, transparent);
  background: color-mix(in srgb, #fbbf24 11%, transparent);
  color: color-mix(in srgb, #fbbf24 76%, var(--journal-ink));
}
.tag-forensics {
  border-color: color-mix(in srgb, #38bdf8 18%, transparent);
  background: color-mix(in srgb, #38bdf8 11%, transparent);
  color: color-mix(in srgb, #38bdf8 78%, var(--journal-ink));
}
.tag-default {
  border-color: var(--journal-soft-border);
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base));
  color: var(--journal-muted);
}

:deep(.el-button.is-plain) {
  border-color: var(--journal-control-border);
  background: var(--journal-surface);
  color: var(--journal-ink);
}

:deep(.el-button.is-plain:hover) {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 6%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.tech-font {
  font-family: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 999px;
}

.status-dot-ready {
  background: #10b981;
  box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.4);
  animation: dot-pulse 1.8s infinite;
}

.status-dot-idle {
  background: #94a3b8;
}

.status-dot-warning {
  background: #f59e0b;
}

.status-dot-solved {
  background: #22c55e;
}

.difficulty-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  margin-right: 6px;
  border-radius: 999px;
}

.difficulty-beginner,
.difficulty-easy {
  background-color: #10b981;
}
.difficulty-medium {
  background-color: #f59e0b;
}
.difficulty-hard {
  background-color: #f97316;
}
.difficulty-insane {
  background-color: #ef4444;
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
</style>
