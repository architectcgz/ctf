<script setup lang="ts">
import { computed } from 'vue'
import {
  Activity,
  BellRing,
  MapPinned,
  Trophy,
} from 'lucide-vue-next'

import RadarChart from '@/components/charts/RadarChart.vue'

import type { StudentOverviewProps } from './overviewProps'

const props = withDefaults(
  defineProps<
    StudentOverviewProps & {
      embedded?: boolean
    }
  >(),
  {
    embedded: false,
  }
)

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()

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
    value: props.recommendations.length > 0 ? '可训练' : '空闲',
    description:
      props.recommendations.length > 0 ? '存在可立即进入的推荐题目' : '当前没有推荐训练任务',
    status: props.recommendations.length > 0 ? 'ready' : 'idle',
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
</script>

<template>
  <section
    class="space-y-6 flex min-h-full flex-1 flex-col"
    :class="
      embedded
        ? 'journal-shell-embedded'
        : 'journal-shell journal-hero rounded-[30px] border px-6 py-6 md:px-8'
    "
  >
    <div>
        <div class="journal-eyebrow">Training Journal</div>
        <h1 class="journal-page-title mt-3 max-w-3xl text-[var(--journal-ink)]">
          {{ displayName }} 的训练总览
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          这里汇总了训练进度、能力分布和近期状态。
        </p>

        <div class="journal-actions mt-6">
          <button type="button" class="journal-btn-primary" @click="emit('openChallenges')">
            继续训练
          </button>
          <button type="button" class="journal-btn-outline" @click="emit('openSkillProfile')">
            查看能力画像
          </button>
        </div>
    </div>
    <div class="journal-board" :class="{ 'journal-board--embedded': embedded }">
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
              class="mt-6 rounded-[18px] border border-dashed border-[var(--journal-shell-border)] px-4 py-10 text-center text-sm text-[var(--journal-muted)]"
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
      </section>
    </div>
  </section>
</template>

<style scoped>
.journal-shell-embedded,
.journal-shell {
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--color-border-default) 70%, transparent);
  --journal-control-border: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --journal-divider: color-mix(in srgb, var(--color-border-default) 64%, transparent);
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  font-family:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
}

.journal-shell-embedded {
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.journal-hero {
  border-color: var(--journal-shell-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base)));
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-board {
  margin-top: 1.5rem;
  border-top: 1px solid var(--journal-divider);
  padding-top: 1.25rem;
}

.journal-board--embedded {
  margin-top: 1.25rem;
}

.journal-panel {
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.journal-metric,
.journal-inline-item,
.journal-rank-summary {
  border: 1px solid var(--journal-shell-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  box-shadow: none;
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: 0.2rem 0.75rem;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.journal-bento {
  display: grid;
  gap: 1.25rem;
}

.journal-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.journal-btn-primary,
.journal-btn-outline {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 36px;
  border-radius: 10px;
  padding: 0.5rem 1rem;
  font-size: 0.82rem;
  font-weight: 600;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}

.journal-btn-primary:hover,
.journal-btn-outline:hover {
  transform: translateY(-1px);
}

.journal-btn-primary {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

.journal-btn-primary:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 66%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 16%, transparent);
}

.journal-btn-outline {
  border: 1px solid var(--journal-control-border);
  background: var(--journal-surface);
  color: var(--journal-muted);
}

.journal-btn-outline:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 52%, var(--journal-control-border));
  color: var(--journal-accent-strong);
}

.journal-btn-primary:focus-visible,
.journal-btn-outline:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--journal-accent) 58%, white);
  outline-offset: 2px;
}

@media (min-width: 1280px) {
  .journal-bento {
    grid-template-columns: 1.1fr 0.92fr 0.88fr;
    grid-template-areas: 'radar rank ops';
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
    border-right: 1px solid var(--journal-divider);
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
    border-right: 1px solid var(--journal-divider);
  }
  .journal-ops-card {
    grid-area: ops;
  }
}

.journal-metric-accent {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface)),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, transparent)
  );
}

.journal-inline-item + .journal-inline-item {
  margin-top: 0.75rem;
}

.tech-font {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
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

:global([data-theme='dark']) .journal-shell {
  --journal-ink: color-mix(in srgb, var(--color-text-primary) 88%, var(--color-text-secondary));
  --journal-muted: var(--color-text-secondary);
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--color-border-default) 70%, transparent);
  --journal-control-border: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --journal-divider: color-mix(in srgb, var(--color-border-default) 64%, transparent);
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 16%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
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

@media (max-width: 767px) {
  .journal-btn-primary,
  .journal-btn-outline {
    min-height: 38px;
    padding-inline: 0.95rem;
  }
}
</style>
