<script setup lang="ts">
import { computed } from 'vue'
import { ChevronRight, Cpu, ScanSearch, TerminalSquare } from 'lucide-vue-next'

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
const consoleStats = computed(() => [
  { key: 'score', label: 'score_total', value: props.progress.total_score ?? 0 },
  { key: 'solved', label: 'solved_count', value: props.progress.total_solved ?? 0 },
  { key: 'rank', label: 'rank_now', value: `#${props.progress.rank ?? '-'}` },
  { key: 'coverage', label: 'coverage_rate', value: `${props.completionRate}%` },
])
</script>

<template>
  <div class="signal-shell space-y-6">
    <section class="signal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="signal-kicker">session://dashboard/{{ displayName }}</div>
        <div class="flex flex-wrap items-center gap-2 text-xs text-[var(--signal-text-dim)]">
          <span class="signal-chip">role: student</span>
          <span class="signal-chip">class: {{ className || 'free-mode' }}</span>
          <span class="signal-chip">variant: terminal-grid</span>
        </div>
      </div>

      <div class="mt-6 grid gap-6 xl:grid-cols-[1.02fr_0.98fr]">
        <div>
          <h2 class="text-3xl font-semibold tracking-tight text-[var(--signal-text)] md:text-[2.4rem]">
            SIGNAL CONSOLE / {{ displayName }}
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--signal-text-dim)]">
            这套方案更像终端与蓝图混合的控制面板，强调技术感、扫描感和操作台气质，更偏 CTF / SOC / 攻防演练中心。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <ElButton type="primary" @click="emit('openChallenges')">继续训练</ElButton>
            <ElButton plain @click="emit('openSkillProfile')">查看能力画像</ElButton>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <article
            v-for="item in consoleStats"
            :key="item.key"
            class="signal-stat rounded-[20px] border px-4 py-4"
          >
            <div class="text-[11px] uppercase tracking-[0.24em] text-[var(--signal-text-dim)]">{{ item.label }}</div>
            <div class="mt-3 text-[30px] font-semibold tracking-tight text-[var(--signal-text)]">{{ item.value }}</div>
          </article>
        </div>
      </div>
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
      <div class="grid gap-4">
        <article class="signal-panel rounded-[26px] border px-6 py-6">
          <div class="flex items-center gap-3">
            <TerminalSquare class="h-5 w-5 text-[var(--signal-accent)]" />
            <div>
              <div class="signal-kicker">queue://recommended</div>
              <h3 class="mt-2 text-xl font-semibold text-[var(--signal-text)]">推荐任务流</h3>
            </div>
          </div>

          <div v-if="quickRecommendations.length === 0" class="mt-5 rounded-[18px] border border-dashed border-white/12 px-4 py-10 text-center text-sm text-[var(--signal-text-dim)]">
            当前没有推荐题目，直接去挑战列表挑一道新题即可。
          </div>

          <div v-else class="mt-5 space-y-3">
            <button
              v-for="item in quickRecommendations"
              :key="item.challenge_id"
              type="button"
              class="signal-rec-item flex w-full items-start gap-4 rounded-[18px] border px-4 py-4 text-left transition"
              @click="emit('openChallenge', item.challenge_id)"
            >
              <div class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-[var(--signal-accent)]/30 bg-[var(--signal-accent)]/10 text-[var(--signal-accent)]">
                <ChevronRight class="h-4 w-4" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <div class="text-sm font-semibold text-[var(--signal-text)]">{{ item.title }}</div>
                    <div class="mt-1 text-xs uppercase tracking-[0.2em] text-[var(--signal-text-dim)]">{{ item.category }}</div>
                  </div>
                  <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="difficultyClass(item.difficulty)">
                    {{ difficultyLabel(item.difficulty) }}
                  </span>
                </div>
                <div class="mt-3 text-sm leading-6 text-[var(--signal-text-dim)]">{{ item.reason }}</div>
              </div>
            </button>
          </div>
        </article>

        <article class="signal-panel rounded-[26px] border px-6 py-6">
          <div class="flex items-center gap-3">
            <ScanSearch class="h-5 w-5 text-[var(--signal-accent)]" />
            <div>
              <div class="signal-kicker">log://recent-activity</div>
              <h3 class="mt-2 text-xl font-semibold text-[var(--signal-text)]">近期日志</h3>
            </div>
          </div>

          <div v-if="recentTimeline.length === 0" class="mt-5 rounded-[18px] border border-dashed border-white/12 px-4 py-10 text-center text-sm text-[var(--signal-text-dim)]">
            当前还没有训练动态。
          </div>

          <div v-else class="mt-5 space-y-3">
            <article
              v-for="event in recentTimeline"
              :key="event.id"
              class="signal-log rounded-[18px] border px-4 py-4"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="text-sm font-medium text-[var(--signal-text)]">{{ event.title }}</div>
                <div class="text-xs text-[var(--signal-text-dim)]">{{ formatDate(event.created_at) }}</div>
              </div>
              <div class="mt-2 text-sm leading-6 text-[var(--signal-text-dim)]">{{ timelineSummary(event) }}</div>
            </article>
          </div>
        </article>
      </div>

      <div class="grid gap-4">
        <article class="signal-panel rounded-[26px] border px-6 py-6">
          <div class="flex items-center gap-3">
            <Cpu class="h-5 w-5 text-[var(--signal-accent)]" />
            <div>
              <div class="signal-kicker">node://weak-points</div>
              <h3 class="mt-2 text-xl font-semibold text-[var(--signal-text)]">薄弱节点</h3>
            </div>
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <span
              v-for="item in weakDimensions.slice(0, 4)"
              :key="item"
              class="rounded-full border border-[var(--signal-accent)]/26 bg-[var(--signal-accent)]/10 px-3 py-1 text-xs font-medium text-[var(--signal-accent-soft)]"
            >
              {{ item }}
            </span>
            <span
              v-if="weakDimensions.length === 0"
              class="rounded-full border border-emerald-400/22 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-200"
            >
              balanced
            </span>
          </div>
        </article>

        <article class="signal-panel rounded-[26px] border px-6 py-6">
          <div class="signal-kicker">board://status-cards</div>
          <h3 class="mt-2 text-xl font-semibold text-[var(--signal-text)]">状态摘要</h3>

          <div class="mt-5 space-y-3">
            <article
              v-for="item in highlightItems"
              :key="item.label"
              class="signal-summary rounded-[18px] border px-4 py-4"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="text-sm font-medium text-[var(--signal-text)]">{{ item.label }}</div>
                <div class="text-lg font-semibold text-[var(--signal-accent-soft)]">{{ item.value }}</div>
              </div>
              <div class="mt-2 text-sm leading-6 text-[var(--signal-text-dim)]">{{ item.description }}</div>
            </article>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.signal-shell {
  --signal-accent: #22d3ee;
  --signal-accent-soft: #a5f3fc;
  --signal-text: #e2f7ff;
  --signal-text-dim: rgba(148, 163, 184, 0.82);
  font-family: "JetBrains Mono", "IBM Plex Mono", "SFMono-Regular", "Noto Sans Mono CJK SC", monospace;
}

.signal-hero,
.signal-panel,
.signal-stat,
.signal-rec-item,
.signal-log,
.signal-summary {
  border-color: rgba(255, 255, 255, 0.08);
}

.signal-hero {
  background:
    linear-gradient(rgba(34, 211, 238, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(34, 211, 238, 0.06) 1px, transparent 1px),
    linear-gradient(135deg, rgba(2, 6, 23, 0.94), rgba(8, 47, 73, 0.74));
  background-size: 18px 18px, 18px 18px, auto;
}

.signal-panel,
.signal-stat,
.signal-rec-item,
.signal-log,
.signal-summary {
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.84), rgba(2, 6, 23, 0.94));
}

.signal-kicker {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.24em;
  text-transform: uppercase;
  color: var(--signal-text-dim);
}

.signal-chip {
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  padding: 0.32rem 0.7rem;
  background: rgba(15, 23, 42, 0.5);
}

.signal-rec-item:hover {
  transform: translateY(-1px);
  border-color: rgba(34, 211, 238, 0.24);
}
</style>
