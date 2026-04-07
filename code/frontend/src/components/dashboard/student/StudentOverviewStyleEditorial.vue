<script setup lang="ts">
import { computed } from 'vue'
import { Activity, BellRing, MapPinned, Trophy } from 'lucide-vue-next'

import RadarChart from '@/components/charts/RadarChart.vue'
import type { SkillDimensionScore } from '@/api/contracts'

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
const normalizedSkillDimensions = computed<SkillDimensionScore[]>(() =>
  props.skillDimensions
    .map((item, index) => {
      const normalizedName = item.name?.trim()
      const numericValue = Number(item.value)
      if (!normalizedName || !Number.isFinite(numericValue)) {
        return null
      }
      return {
        ...item,
        key: item.key || `${normalizedName}-${index}`,
        name: normalizedName,
        value: Math.min(100, Math.max(0, numericValue)),
      }
    })
    .filter((item): item is SkillDimensionScore => item !== null)
)
const hasRadarData = computed(() => normalizedSkillDimensions.value.length > 0)
const radarIndicators = computed(() =>
  normalizedSkillDimensions.value.map((item) => ({ name: item.name, max: 100 }))
)
const radarValues = computed(() => normalizedSkillDimensions.value.map((item) => item.value))
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
    class="journal-soft-surface space-y-6 flex min-h-full flex-1 flex-col"
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
          <div class="journal-panel-head">
            <div>
              <div class="journal-eyebrow">Skill Matrix</div>
              <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">能力雷达</h3>
            </div>
            <MapPinned class="h-5 w-5 text-[var(--journal-accent-strong)]" />
          </div>
          <div v-if="hasRadarData" class="journal-radar-body mt-4">
            <div class="journal-radar-chart">
              <div class="skill-dimension-chart__frame">
                <div class="skill-dimension-chart__inner">
                  <RadarChart
                    :indicators="radarIndicators"
                    :values="radarValues"
                    name="能力值"
                    height-class="h-[18rem] md:h-[21rem] xl:h-[23rem]"
                    :label-font-size="15"
                    :axis-name-gap="10"
                    radius="70%"
                    center-y="50%"
                  />
                </div>
              </div>
            </div>
            <div class="journal-radar-dimensions mt-4">
              <article
                v-for="item in normalizedSkillDimensions"
                :key="item.key"
                class="journal-radar-dimension"
              >
                <div class="journal-radar-dimension__label">{{ item.name }}</div>
                <div class="journal-radar-dimension__value tech-font">{{ item.value }}</div>
              </article>
            </div>
          </div>
          <div
            v-else
            class="mt-6 rounded-[18px] border border-dashed border-[var(--journal-shell-border)] px-4 py-10 text-center text-sm text-[var(--journal-muted)]"
          >
            当前能力数据不足，完成更多题目后将生成雷达图。
          </div>
        </article>

        <article class="journal-panel journal-rank-card px-6 py-6">
          <div class="journal-panel-head">
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
              <div class="mt-3 text-[30px] font-semibold tracking-tight text-[var(--journal-ink)]">
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
          <div class="journal-panel-head">
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
                  <component :is="item.icon" class="h-4 w-4 text-[var(--journal-accent-strong)]" />
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
.journal-soft-surface {
  --journal-soft-eyebrow-size: 11px;
  --journal-soft-eyebrow-spacing: 0.12em;
  --journal-soft-eyebrow-color: var(--journal-accent-strong);
  --journal-soft-button-height: 36px;
  --journal-soft-button-padding: 0.5rem 1rem;
  --journal-soft-button-hover-transform: translateY(-1px);
}

.journal-board {
  margin-top: 1.5rem;
  position: relative;
  padding-top: 1.25rem;
  --journal-board-divider: color-mix(in srgb, var(--journal-ink) 22%, transparent);
}

.journal-board::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--journal-board-divider);
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

.journal-bento {
  display: grid;
  gap: 0;
  --journal-bento-divider: color-mix(in srgb, var(--journal-ink) 22%, transparent);
}

.journal-bento > .journal-panel + .journal-panel {
  position: relative;
  padding-top: 1.25rem;
}

.journal-bento > .journal-panel + .journal-panel::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--journal-bento-divider);
}

.journal-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.journal-panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.journal-radar-body {
  display: grid;
  gap: 1rem;
}

.journal-radar-chart {
  width: min(100%, 420px);
  margin: 0 auto;
}

.skill-dimension-chart__frame {
  position: relative;
  margin: 0 auto;
  width: min(100%, 420px);
  aspect-ratio: 1.04;
  overflow: visible;
}

.skill-dimension-chart__frame::before,
.skill-dimension-chart__frame::after {
  content: '';
  position: absolute;
  pointer-events: none;
}

.skill-dimension-chart__frame::before {
  inset: 0;
  clip-path: polygon(25% 6%, 75% 6%, 100% 50%, 75% 94%, 25% 94%, 0 50%);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 94%, var(--color-bg-base)),
      color-mix(
        in srgb,
        var(--journal-surface-subtle, var(--color-bg-elevated)) 96%,
        var(--color-bg-base)
      )
    ),
    linear-gradient(135deg, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent);
  border: 1px solid var(--journal-shell-border);
}

.skill-dimension-chart__frame::after {
  inset: 18px;
  clip-path: polygon(25% 6%, 75% 6%, 100% 50%, 75% 94%, 25% 94%, 0 50%);
  border: 1px solid
    color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 78%, transparent);
  background:
    radial-gradient(
      circle at 50% 45%,
      color-mix(in srgb, var(--journal-accent) 12%, transparent),
      transparent 60%
    ),
    color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 76%, var(--color-bg-base));
}

.skill-dimension-chart__inner {
  position: absolute;
  inset: 18px;
  z-index: 1;
}

.journal-radar-dimensions {
  display: grid;
  gap: 0;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border-radius: 16px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

.journal-radar-dimension {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.9rem 0.95rem;
}

.journal-radar-dimension:nth-child(n + 3) {
  border-top: 1px solid var(--journal-divider);
}

.journal-radar-dimension:nth-child(2n) {
  border-left: 1px solid var(--journal-divider);
}

.journal-radar-dimension__label {
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.journal-radar-dimension__value {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--journal-ink);
}

@media (min-width: 1280px) {
  .journal-bento {
    grid-template-columns: 1.1fr 0.92fr 0.88fr;
    grid-template-areas: 'radar rank ops';
  }

  .journal-radar-card {
    grid-area: radar;
    position: relative;
    padding-right: 1.25rem;
  }

  .journal-rank-card {
    grid-area: rank;
    position: relative;
    padding-right: 1.25rem;
    padding-left: 1.25rem;
  }

  .journal-ops-card {
    grid-area: ops;
    padding-left: 1.25rem;
  }

  .journal-bento > .journal-panel + .journal-panel {
    padding-top: 0;
  }

  .journal-bento > .journal-panel + .journal-panel::before {
    top: 0.5rem;
    right: auto;
    bottom: 0.5rem;
    width: 1px;
    height: auto;
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
  .journal-soft-surface {
    --journal-soft-button-height: 38px;
    --journal-soft-button-padding: 0.5rem 0.95rem;
  }

  .journal-radar-dimensions {
    grid-template-columns: minmax(0, 1fr);
  }

  .journal-radar-dimension:nth-child(2n) {
    border-left: 0;
  }

  .journal-radar-dimension + .journal-radar-dimension {
    border-top: 1px solid var(--journal-divider);
  }
}

:global([data-theme='dark']) .skill-dimension-chart__frame::before {
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    ),
    linear-gradient(135deg, color-mix(in srgb, var(--journal-accent) 18%, transparent), transparent);
}

:global([data-theme='dark']) .skill-dimension-chart__frame::after {
  background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  border-color: rgba(148, 163, 184, 0.2);
}
</style>
