<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Compass, Radar, Sparkles } from 'lucide-vue-next'

import type { MyProgressData, RecommendationItem, TimelineEvent } from '@/api/contracts'
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
      <div class="overflow-hidden rounded-[30px] border border-cyan-400/20 bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.2),transparent_45%),linear-gradient(135deg,rgba(15,23,42,0.92),rgba(8,47,73,0.84))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.24em] text-cyan-100/80">
          <span>Student Workspace</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">{{ className || '自由训练' }}</span>
        </div>
        <div class="mt-5 max-w-2xl">
          <h2 class="text-3xl font-semibold tracking-tight text-white">为 {{ displayName }} 定制的训练概览</h2>
          <p class="mt-3 text-sm leading-7 text-cyan-50/78">
            主页只保留最关键的训练摘要。具体的分类进度、训练建议、近期动态和难度分布，已经拆成左侧导航下的独立页面。
          </p>
        </div>
        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="rounded-2xl border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">当前排名</div>
            <div class="mt-2 text-2xl font-semibold text-white">#{{ progress.rank ?? '-' }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">综合全站训练表现计算</div>
          </div>
          <div class="rounded-2xl border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">完成率</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ completionRate }}%</div>
            <div class="mt-2 text-sm text-cyan-50/70">按当前题量估算的覆盖程度</div>
          </div>
          <div class="rounded-2xl border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">待加强维度</div>
            <div class="mt-2 text-lg font-semibold text-white">{{ weakDimensions[0] || '暂无明显短板' }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">建议从左侧“训练建议”进入细看</div>
          </div>
        </div>
        <div class="mt-6 flex flex-wrap gap-3">
          <ElButton type="primary" @click="emit('openChallenges')">继续训练</ElButton>
          <ElButton plain @click="emit('openSkillProfile')">查看能力画像</ElButton>
        </div>
      </div>

      <div class="grid gap-3">
        <article
          v-for="item in highlightItems"
          :key="item.label"
          class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]"
        >
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">{{ item.label }}</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ item.value }}</div>
            </div>
            <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <component :is="item.icon" class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">{{ item.description }}</div>
        </article>
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
          <button
            v-for="(item, index) in quickRecommendations"
            :key="item.challenge_id"
            type="button"
            class="flex w-full items-start gap-4 rounded-[24px] border border-border bg-[linear-gradient(180deg,rgba(15,23,42,0.8),rgba(8,15,32,0.74))] px-5 py-5 text-left transition hover:-translate-y-0.5 hover:border-primary/60"
            @click="emit('openChallenge', item.challenge_id)"
          >
            <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-primary/12 text-lg font-semibold text-primary">
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
          </button>
        </div>
      </SectionCard>

      <div class="grid gap-4">
        <SectionCard title="训练雷达" subtitle="当前最需要关注的三类信息。">
          <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
            <div class="rounded-[22px] border border-border bg-base/70 px-4 py-4">
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
            </div>
            <div class="rounded-[22px] border border-border bg-base/70 px-4 py-4">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Radar class="h-4 w-4 text-sky-300" />
                动态概览
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ timeline.length }}</div>
              <div class="mt-2 text-sm text-text-secondary">最近实例与提交动作的浓缩视图</div>
            </div>
            <div class="rounded-[22px] border border-border bg-base/70 px-4 py-4">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Sparkles class="h-4 w-4 text-fuchsia-300" />
                训练建议
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ recommendations.length }}</div>
              <div class="mt-2 text-sm text-text-secondary">已经拆成独立页面，可从左侧继续细看</div>
            </div>
          </div>
        </SectionCard>

        <SectionCard title="近期速览" subtitle="只保留最近三条关键动作。">
          <div v-if="recentTimeline.length === 0" class="rounded-2xl border border-dashed border-border px-4 py-10 text-center text-sm text-text-secondary">
            当前还没有训练动态。
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="event in recentTimeline"
              :key="event.id"
              class="rounded-[22px] border border-border bg-base/70 px-4 py-4"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="text-sm font-medium text-text-primary">{{ event.title }}</div>
                <div class="text-xs text-text-muted">{{ formatDate(event.created_at) }}</div>
              </div>
              <div class="mt-2 text-sm leading-6 text-text-secondary">{{ timelineSummary(event) }}</div>
            </div>
          </div>
        </SectionCard>
      </div>
    </section>
  </div>
</template>
