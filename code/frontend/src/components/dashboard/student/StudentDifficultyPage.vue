<script setup lang="ts">
import { computed } from 'vue'
import { Flame, Layers2, ShieldCheck } from 'lucide-vue-next'

import AppCard from '@/components/common/AppCard.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { difficultyLabel } from '@/utils/challenge'

import { progressRate } from './utils'

interface DifficultyStat {
  difficulty: string
  total: number
  solved: number
}

const props = defineProps<{
  difficultyStats: DifficultyStat[]
}>()

const difficultyOrder = ['beginner', 'easy', 'medium', 'hard', 'hell']
const barMap: Record<string, string> = {
  beginner: 'bg-emerald-400',
  easy: 'bg-sky-400',
  medium: 'bg-amber-400',
  hard: 'bg-orange-400',
  hell: 'bg-rose-400',
}

const orderedStats = computed(() =>
  difficultyOrder
    .map((difficulty) => props.difficultyStats.find((item) => item.difficulty === difficulty))
    .filter((item): item is DifficultyStat => Boolean(item))
    .map((item) => ({
      ...item,
      rate: progressRate(item.total, item.solved),
    })),
)

const nextFocus = computed(() =>
  [...orderedStats.value]
    .filter((item) => item.total > 0)
    .sort((left, right) => left.rate - right.rate)[0] || null,
)

function accentForDifficulty(difficulty: string): 'success' | 'primary' | 'warning' | 'danger' | 'violet' {
  if (difficulty === 'beginner') return 'success'
  if (difficulty === 'easy') return 'primary'
  if (difficulty === 'medium') return 'warning'
  if (difficulty === 'hard') return 'danger'
  return 'violet'
}
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 lg:grid-cols-2 xl:grid-cols-5">
      <AppCard
        v-for="item in orderedStats"
        :key="item.difficulty"
        variant="metric"
        :accent="accentForDifficulty(item.difficulty)"
        :eyebrow="difficultyLabel(item.difficulty)"
        :title="`${item.rate}%`"
        :subtitle="`${item.solved} / ${item.total}`"
      >
        <div class="h-2.5 rounded-full bg-black/20">
          <div class="h-2.5 rounded-full" :class="barMap[item.difficulty]" :style="{ width: `${item.rate}%` }" />
        </div>
      </AppCard>
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
      <SectionCard title="难度层级视图" subtitle="观察不同难度的覆盖情况，判断训练是否长期停留在舒适区。">
        <div v-if="orderedStats.length === 0" class="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-text-secondary">
          暂无难度统计数据。
        </div>

        <div v-else class="space-y-4">
          <AppCard
            v-for="item in orderedStats"
            :key="item.difficulty"
            variant="action"
            :accent="accentForDifficulty(item.difficulty)"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <div class="text-sm font-medium text-text-primary">{{ difficultyLabel(item.difficulty) }}</div>
                <div class="mt-2 text-sm text-text-secondary">当前完成 {{ item.solved }} 题，共 {{ item.total }} 题</div>
              </div>
              <div class="text-right">
                <div class="text-2xl font-semibold text-text-primary">{{ item.rate }}%</div>
                <div class="mt-1 text-xs uppercase tracking-[0.14em] text-text-muted">覆盖比例</div>
              </div>
            </div>
            <div class="mt-4 h-3 rounded-full bg-[var(--color-bg-base)]">
              <div class="h-3 rounded-full" :class="barMap[item.difficulty]" :style="{ width: `${item.rate}%` }" />
            </div>
          </AppCard>
        </div>
      </SectionCard>

      <div class="grid gap-4">
        <SectionCard title="下一阶段建议" subtitle="根据当前难度覆盖情况给出训练方向。">
          <AppCard variant="action" accent="warning">
            <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
              <Flame class="h-4 w-4 text-amber-300" />
              建议优先处理
            </div>
            <div class="mt-3 text-2xl font-semibold text-text-primary">{{ nextFocus ? difficultyLabel(nextFocus.difficulty) : '暂无数据' }}</div>
            <div class="mt-2 text-sm leading-6 text-text-secondary">
              {{
                nextFocus
                  ? `当前仅完成 ${nextFocus.rate}% ，建议补齐这一难度层级，避免训练结构长期停留在低或中低难度。`
                  : '先完成几道题，系统才能给出更清晰的难度结构判断。'
              }}
            </div>
          </AppCard>

          <AppCard variant="action" accent="primary">
            <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
              <Layers2 class="h-4 w-4 text-sky-300" />
              难度结构
            </div>
            <div class="mt-2 text-sm leading-6 text-text-secondary">
              如果你只做简单题，总分和解题数会涨，但整体结构不会健康。这里更适合看难度分层是否均衡上探。
            </div>
          </AppCard>

          <AppCard variant="action" accent="success">
            <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
              <ShieldCheck class="h-4 w-4 text-emerald-300" />
              目标状态
            </div>
            <div class="mt-2 text-sm leading-6 text-text-secondary">
              目标不是所有难度都平均，而是从当前能力台阶稳定上探，逐步形成“简单稳定、中等推进、困难试探”的结构。
            </div>
          </AppCard>
        </SectionCard>
      </div>
    </section>
  </div>
</template>
