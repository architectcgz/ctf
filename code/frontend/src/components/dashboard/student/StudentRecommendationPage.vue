<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Crosshair, ShieldAlert, Sparkles, TrendingUp } from 'lucide-vue-next'

import type { RecommendationItem } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'

const props = defineProps<{
  weakDimensions: string[]
  recommendations: RecommendationItem[]
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
  openChallenges: []
  openSkillProfile: []
}>()

const headline = computed(() => props.weakDimensions[0] || '当前训练结构较均衡')
const topRecs = computed(() => props.recommendations.slice(0, 3))
</script>

<template>
  <div class="space-y-6">
    <!-- 顶部概览 -->
    <section class="grid gap-4 lg:grid-cols-[1fr_1fr_auto]">
      <!-- 薄弱维度 -->
      <AppCard variant="hero" accent="warning" eyebrow="Priority Focus" title="补短板计划"
        subtitle="根据当前薄弱维度给出优先训练顺序，建议先完成靠前题目，再回看能力画像确认是否抬升。">
        <div class="mt-2 flex flex-wrap gap-2">
          <template v-if="weakDimensions.length > 0">
            <span
              v-for="dim in weakDimensions.slice(0, 4)"
              :key="dim"
              class="inline-flex items-center gap-1.5 rounded-md border border-[var(--color-warning)]/20 bg-[var(--color-warning)]/8 px-2.5 py-1 text-xs font-medium text-[var(--color-warning)]"
            >
              <ShieldAlert class="h-3 w-3" />
              {{ dim }}
            </span>
          </template>
          <span
            v-else
            class="inline-flex items-center rounded-md border border-[var(--color-success)]/20 bg-[var(--color-success)]/8 px-2.5 py-1 text-xs font-medium text-[var(--color-success)]"
          >
            暂无明显短板
          </span>
        </div>
        <div class="mt-3 text-sm text-text-secondary">
          当前首要关注：<span class="font-medium text-text-primary">{{ headline }}</span>
        </div>
      </AppCard>

      <!-- 队列数量 -->
      <AppCard variant="metric" accent="warning" eyebrow="Queued" :title="String(recommendations.length)">
        <template #header>
          <div class="rec-icon-box rec-icon-box--warning">
            <TrendingUp class="h-5 w-5" />
          </div>
        </template>
        <div class="text-sm leading-6 text-text-secondary">可直接推进的建议题目，建议按顺序完成前排任务。</div>
      </AppCard>

      <!-- 快捷操作 -->
      <div class="flex flex-col justify-between gap-3 py-0.5">
        <button
          class="rec-action-btn rec-action-btn--primary"
          @click="emit('openChallenges')"
        >
          打开挑战列表
          <ArrowRight class="h-3.5 w-3.5" />
        </button>
        <button
          class="rec-action-btn rec-action-btn--ghost"
          @click="emit('openSkillProfile')"
        >
          查看能力画像
        </button>
      </div>
    </section>

    <!-- 推荐队列（Top 3 卡片） -->
    <SectionCard
      title="推荐队列"
      subtitle="按当前训练阶段排序的优先挑战，建议按顺序推进。"
    >
      <template #header>
        <span class="rounded-full border border-[var(--color-warning)]/20 bg-[var(--color-warning)]/8 px-2.5 py-0.5 text-xs font-semibold text-[var(--color-warning)]">
          TOP {{ topRecs.length }}
        </span>
      </template>

      <div v-if="recommendations.length === 0" class="rounded-2xl border border-dashed border-border-subtle px-4 py-12 text-center text-sm text-text-secondary">
        当前没有推荐题目，可以先去挑战列表探索新的方向。
      </div>

      <div v-else class="grid gap-3 md:grid-cols-3">
        <AppCard
          v-for="(item, index) in topRecs"
          :key="item.challenge_id"
          variant="metric"
          accent="warning"
          :eyebrow="`Queue ${index + 1}`"
          :title="item.title"
        >
          <template #header>
            <div class="rec-icon-box rec-icon-box--warning">
              <Crosshair class="h-4 w-4" />
            </div>
          </template>
          <div class="flex flex-wrap items-center gap-2 text-xs text-text-muted">
            <span class="uppercase tracking-wide">{{ item.category }}</span>
            <span class="h-1 w-1 rounded-full bg-border-subtle" />
            <span
              class="rounded-full px-2 py-0.5 text-xs font-medium normal-case tracking-normal"
              :class="difficultyClass(item.difficulty)"
            >
              {{ difficultyLabel(item.difficulty) }}
            </span>
          </div>
        </AppCard>
      </div>
    </SectionCard>

    <!-- 完整推荐列表 -->
    <SectionCard
      title="推荐列表"
      subtitle="完整推荐项，保留推荐原因，适合逐条推进。"
    >
      <div v-if="recommendations.length === 0" class="rounded-2xl border border-dashed border-border-subtle px-4 py-12 text-center text-sm text-text-secondary">
        当前没有推荐题目，可以先去挑战列表探索新的方向。
      </div>

      <div v-else class="space-y-3">
        <button
          v-for="(item, index) in recommendations"
          :key="item.challenge_id"
          class="rec-list-item group w-full cursor-pointer text-left"
          @click="emit('openChallenge', item.challenge_id)"
        >
          <div class="flex items-start gap-4">
            <!-- 序号 -->
            <div
              class="rec-list-index"
              :class="index === 0 ? 'rec-list-index--top' : 'rec-list-index--rest'"
            >
              {{ index + 1 }}
            </div>

            <div class="min-w-0 flex-1">
              <!-- 标题行 -->
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-sm font-semibold text-text-primary">{{ item.title }}</span>
                <span
                  class="rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="difficultyClass(item.difficulty)"
                >
                  {{ difficultyLabel(item.difficulty) }}
                </span>
              </div>

              <!-- 元信息 -->
              <div class="mt-1 flex items-center gap-2 text-xs text-text-muted">
                <Crosshair class="h-3 w-3" />
                <span class="uppercase tracking-wide">{{ item.category }}</span>
                <span class="h-1 w-1 rounded-full bg-border-subtle" />
                <span>Queue {{ index + 1 }}</span>
              </div>

              <!-- 推荐理由 -->
              <div class="mt-3 flex items-start gap-2 rounded-xl border border-border-subtle bg-bg-base/50 px-3 py-2.5">
                <Sparkles class="mt-0.5 h-3.5 w-3.5 shrink-0 text-[var(--color-warning)]/70" />
                <p class="text-xs leading-5 text-text-secondary">{{ item.reason }}</p>
              </div>
            </div>

            <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-text-muted transition-transform duration-150 group-hover:translate-x-0.5 group-hover:text-[var(--color-warning)]" />
          </div>
        </button>
      </div>
    </SectionCard>
  </div>
</template>

<style scoped>
.rec-icon-box {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
}

.rec-icon-box--warning {
  color: var(--color-warning);
  border-color: color-mix(in srgb, var(--color-warning) 18%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg-surface));
}

.rec-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  border-radius: 0.75rem;
  padding: 0.5rem 1rem;
  font-size: 0.8125rem;
  font-weight: 500;
  transition: background 150ms, color 150ms, border-color 150ms;
  cursor: pointer;
  white-space: nowrap;
}

.rec-action-btn--primary {
  background: var(--color-warning);
  color: #fff;
  border: 1px solid transparent;
}

.rec-action-btn--primary:hover {
  background: color-mix(in srgb, var(--color-warning) 85%, white);
}

.rec-action-btn--ghost {
  background: transparent;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-default);
}

.rec-action-btn--ghost:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}

.rec-list-item {
  border-radius: 1rem;
  border: 1px solid var(--color-border-subtle);
  background: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  padding: 1rem 1.25rem;
  transition: border-color 150ms, background 150ms;
}

.rec-list-item:hover {
  border-color: color-mix(in srgb, var(--color-warning) 30%, var(--color-border-subtle));
  background: color-mix(in srgb, var(--color-warning) 4%, var(--color-bg-surface));
}

.rec-list-index {
  display: flex;
  height: 2rem;
  width: 2rem;
  shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 0.625rem;
  font-size: 0.875rem;
  font-weight: 600;
  flex-shrink: 0;
}

.rec-list-index--top {
  background: var(--color-warning);
  color: #fff;
}

.rec-list-index--rest {
  border: 1px solid color-mix(in srgb, var(--color-warning) 20%, var(--color-border-subtle));
  background: color-mix(in srgb, var(--color-warning) 8%, var(--color-bg-surface));
  color: var(--color-warning);
}
</style>
