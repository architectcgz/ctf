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
      <AppCard
        variant="hero"
        accent="primary"
        eyebrow="Student Workspace"
        :title="`为 ${displayName} 定制的训练概览`"
        subtitle="主页只保留最关键的训练摘要。具体的分类进度、训练建议、近期动态和难度分布，已经拆成左侧导航下的独立页面。"
      >
        <template #header>
          <span
            class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
            style="border-color: color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default)); background-color: var(--color-primary-soft); color: var(--color-primary);"
          >
            {{ className || '自由训练' }}
          </span>
        </template>

        <div class="grid gap-3 md:grid-cols-3">
          <AppCard variant="metric" accent="primary" eyebrow="当前排名" :title="`#${progress.rank ?? '-'}`" subtitle="综合全站训练表现计算" />
          <AppCard variant="metric" accent="primary" eyebrow="完成率" :title="`${completionRate}%`" subtitle="按当前题量估算的覆盖程度" />
          <AppCard
            variant="metric"
            accent="warning"
            eyebrow="待加强维度"
            :title="weakDimensions[0] || '暂无明显短板'"
            subtitle="建议从左侧“训练建议”进入细看"
          />
        </div>

        <template #footer>
          <div class="flex flex-wrap gap-3">
            <ElButton type="primary" @click="emit('openChallenges')">继续训练</ElButton>
            <ElButton plain @click="emit('openSkillProfile')">查看能力画像</ElButton>
          </div>
        </template>
      </AppCard>

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
        <div v-if="quickRecommendations.length === 0" class="rounded-2xl border border-dashed border-border px-4 py-10 text-center text-sm text-text-secondary">
          当前没有推荐题目，直接去挑战列表挑一道新题即可。
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
                  class="rounded-full border border-amber-500/20 bg-amber-500/10 px-3 py-1 text-xs font-medium text-amber-200"
                >
                  {{ item }}
                </span>
                <span
                  v-if="weakDimensions.length === 0"
                  class="rounded-full border border-emerald-500/20 bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-200"
                >
                  结构均衡
                </span>
              </div>
            </AppCard>
            <AppCard variant="action" accent="primary">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Radar class="h-4 w-4 text-sky-300" />
                动态概览
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ timeline.length }}</div>
              <div class="mt-2 text-sm text-text-secondary">最近实例与提交动作的浓缩视图</div>
            </AppCard>
            <AppCard variant="action" accent="violet">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Sparkles class="h-4 w-4 text-fuchsia-300" />
                训练建议
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ recommendations.length }}</div>
              <div class="mt-2 text-sm text-text-secondary">已经拆成独立页面，可从左侧继续细看</div>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard title="近期速览" subtitle="只保留最近三条关键动作。">
          <div v-if="recentTimeline.length === 0" class="rounded-2xl border border-dashed border-border px-4 py-10 text-center text-sm text-text-secondary">
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
