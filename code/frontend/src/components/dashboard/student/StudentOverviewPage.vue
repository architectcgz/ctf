<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Compass, Radar, Sparkles } from 'lucide-vue-next'

import type { MyProgressData, RecommendationItem, TimelineEvent } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { formatDate } from '@/utils/format'

import type { DashboardHighlightItem } from './types'
import { timelineSummary } from './utils'

const props = defineProps<{
  displayName: string
  className?: string
  progress: MyProgressData
  completionRate: number
  highlightItems: DashboardHighlightItem[]
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  weakDimensions: string[]
}>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()

const quickRecommendations = computed(() => props.recommendations.slice(0, 3))
const recentTimeline = computed(() => props.timeline.slice(0, 3))
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 xl:grid-cols-[1.12fr_0.88fr]">
      <div class="student-overview-legacy-hero pb-6">
        <div class="student-overview-legacy-eyebrow flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em]">
          <span>Student Workspace</span>
          <span class="student-overview-legacy-eyebrow-separator px-2 py-1">
            {{ className || '自由训练' }}
          </span>
        </div>
        <h2 class="student-overview-legacy-title mt-3 text-3xl font-semibold tracking-tight">
          为 {{ displayName }} 定制的训练概览
        </h2>
        <p class="student-overview-legacy-copy mt-3 text-sm leading-7">
          先看当前排名、完成率和待加强维度，再决定今天优先推进哪一类训练。
        </p>

        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="student-overview-legacy-stat px-4 py-4">
            <div class="student-overview-legacy-stat-label text-[11px] uppercase tracking-[0.18em]">当前排名</div>
            <div class="student-overview-legacy-stat-value mt-2 text-2xl font-semibold">#{{ progress.rank ?? '-' }}</div>
            <div class="student-overview-legacy-stat-copy mt-2 text-sm">综合全站训练表现计算</div>
          </div>
          <div class="student-overview-legacy-stat px-4 py-4">
            <div class="student-overview-legacy-stat-label text-[11px] uppercase tracking-[0.18em]">完成率</div>
            <div class="student-overview-legacy-stat-value mt-2 text-2xl font-semibold">{{ completionRate }}%</div>
            <div class="student-overview-legacy-stat-copy mt-2 text-sm">按当前题量估算的覆盖程度</div>
          </div>
          <div class="student-overview-legacy-stat px-4 py-4">
            <div class="student-overview-legacy-stat-label text-[11px] uppercase tracking-[0.18em]">待加强维度</div>
            <div class="student-overview-legacy-stat-value mt-2 text-2xl font-semibold">
              {{ weakDimensions[0] || '暂无明显短板' }}
            </div>
            <div class="student-overview-legacy-stat-copy mt-2 text-sm">建议从左侧“训练建议”进入细看</div>
          </div>
        </div>

        <div class="mt-6 flex flex-wrap gap-3">
          <ElButton type="primary" @click="emit('openChallenges')">继续训练</ElButton>
          <ElButton plain @click="emit('openSkillProfile')">查看能力画像</ElButton>
        </div>
      </div>

      <div class="grid gap-3">
        <AppCard
          v-for="item in highlightItems"
          :key="item.label"
          variant="metric"
          accent="primary"
          :eyebrow="item.label"
          :title="String(item.value)"
        >
          <template #header>
            <div class="flex h-12 w-12 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary">
              <component :is="item.icon" class="h-5 w-5" />
            </div>
          </template>
          <div class="text-sm leading-6 text-text-secondary">{{ item.description }}</div>
        </AppCard>
      </div>
    </section>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <MetricCard label="总得分" :value="progress.total_score ?? 0" hint="当前训练累计积分" accent="primary" />
      <MetricCard label="已解题数" :value="progress.total_solved ?? 0" hint="已成功提交并判定正确的题目数" accent="success" />
      <MetricCard label="当前排名" :value="`#${progress.rank ?? '-'}`" hint="综合全站训练结果计算" accent="warning" />
      <MetricCard label="完成率" :value="`${completionRate}%`" hint="按分类总题量估算的覆盖比例" accent="primary" />
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.02fr_0.98fr]">
      <SectionCard title="优先训练队列" subtitle="主页保留最值得立刻动手的题目，完整列表请从左侧“训练建议”进入。">
        <div v-if="quickRecommendations.length === 0" class="border border-dashed border-border px-4 py-10 text-center text-sm text-text-secondary">
          当前没有推荐题目，直接去题目列表挑一道新题即可。
        </div>

        <div v-else class="space-y-3">
          <AppCard
            v-for="(item, index) in quickRecommendations"
            :key="item.challenge_id"
            as="button"
            variant="action"
            accent="primary"
            interactive
            class="cursor-pointer"
            @click="emit('openChallenge', item.challenge_id)"
          >
            <div class="flex w-full items-start gap-4">
              <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-lg font-semibold text-primary">
                0{{ index + 1 }}
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <div class="text-base font-semibold text-text-primary">{{ item.title }}</div>
                    <div class="mt-1 flex flex-wrap items-center gap-2 text-xs uppercase tracking-[0.16em] text-text-muted">
                      <span>{{ item.category }}</span>
                      <span class="h-1 w-1 rounded-full bg-border" />
                      <span>推荐入口</span>
                    </div>
                  </div>
                  <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="difficultyClass(item.difficulty)">
                    {{ difficultyLabel(item.difficulty) }}
                  </span>
                </div>
                <p class="mt-3 text-sm leading-6 text-text-secondary">{{ item.reason }}</p>
              </div>
              <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-primary" />
            </div>
          </AppCard>
        </div>
      </SectionCard>

      <div class="grid gap-4">
        <SectionCard title="训练雷达" subtitle="当前最需要关注的三类信息。">
          <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
            <AppCard variant="action" accent="warning">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Compass class="h-4 w-4 text-primary" />
                待加强维度
              </div>
              <div class="mt-3 flex flex-wrap gap-2">
                <span
                  v-for="item in weakDimensions.slice(0, 3)"
                  :key="item"
                  class="rounded-full border border-[var(--color-warning)]/20 bg-[var(--color-warning)]/10 px-3 py-1 text-xs font-medium text-[var(--color-warning)]"
                >
                  {{ item }}
                </span>
                <span
                  v-if="weakDimensions.length === 0"
                  class="rounded-full border border-[var(--color-success)]/20 bg-[var(--color-success)]/10 px-3 py-1 text-xs font-medium text-[var(--color-success)]"
                >
                  结构均衡
                </span>
              </div>
            </AppCard>
            <AppCard variant="action" accent="primary">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Radar class="h-4 w-4 text-[var(--color-primary-hover)]" />
                动态概览
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ timeline.length }}</div>
              <div class="mt-2 text-sm text-text-secondary">最近训练动作的浓缩视图</div>
            </AppCard>
            <AppCard variant="action" accent="violet">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Sparkles class="h-4 w-4 text-[var(--color-cat-reverse)]" />
                训练建议
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ recommendations.length }}</div>
              <div class="mt-2 text-sm text-text-secondary">已经拆成独立页面，可从左侧继续细看</div>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard title="近期速览" subtitle="只保留最近三条关键动作。">
          <div v-if="recentTimeline.length === 0" class="border border-dashed border-border px-4 py-10 text-center text-sm text-text-secondary">
            当前还没有训练动态。
          </div>

          <div v-else class="space-y-3">
            <AppCard
              v-for="event in recentTimeline"
              :key="event.id"
              variant="action"
              accent="neutral"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="text-sm font-medium text-text-primary">{{ event.title }}</div>
                <div class="text-xs text-text-muted">{{ formatDate(event.created_at) }}</div>
              </div>
              <div class="mt-2 text-sm leading-6 text-text-secondary">{{ timelineSummary(event) }}</div>
            </AppCard>
          </div>
        </SectionCard>
      </div>
    </section>
  </div>
</template>

<style scoped>
.student-overview-legacy-hero {
  border-bottom: 1px solid color-mix(in srgb, var(--color-primary) 25%, transparent);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface)),
    color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base))
  );
}

.student-overview-legacy-eyebrow {
  color: color-mix(in srgb, var(--color-text-primary) 75%, transparent);
}

.student-overview-legacy-eyebrow-separator,
.student-overview-legacy-stat {
  border-left: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
}

.student-overview-legacy-title,
.student-overview-legacy-stat-value {
  color: var(--color-text-primary);
}

.student-overview-legacy-copy {
  color: color-mix(in srgb, var(--color-text-primary) 80%, var(--color-text-secondary));
}

.student-overview-legacy-stat-label {
  color: color-mix(in srgb, var(--color-text-primary) 60%, transparent);
}

.student-overview-legacy-stat-copy {
  color: color-mix(in srgb, var(--color-text-primary) 70%, var(--color-text-secondary));
}
</style>
