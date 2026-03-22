<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Crosshair, ShieldAlert, Sparkles } from 'lucide-vue-next'

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
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 xl:grid-cols-[0.96fr_1.04fr]">
      <AppCard
        variant="hero"
        accent="warning"
        eyebrow="Targeted Training"
        title="补短板计划"
        subtitle="根据当前薄弱维度给出优先训练顺序，建议先完成靠前题目，再回看能力画像确认是否抬升。"
      >
        <div class="grid gap-4 md:grid-cols-[1.08fr_0.92fr]">
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
              <ShieldAlert class="h-4 w-4 text-amber-500" />
              优先修复的能力维度
            </div>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="item in weakDimensions.slice(0, 4)"
                :key="item"
                class="rounded-full border border-amber-500/18 bg-amber-500/8 px-3 py-1 text-xs font-medium text-amber-600"
              >
                {{ item }}
              </span>
              <span
                v-if="weakDimensions.length === 0"
                class="rounded-full border border-emerald-500/18 bg-emerald-500/8 px-3 py-1 text-xs font-medium text-emerald-600"
              >
                暂无明显短板
              </span>
            </div>
            <div class="text-sm text-text-secondary">当前首要关注：{{ headline }}</div>
          </div>

          <AppCard variant="action" accent="warning">
            <div class="text-[11px] font-semibold uppercase tracking-[0.18em] text-text-muted">
              当前队列
            </div>
            <div class="mt-3 text-3xl font-semibold tracking-tight text-text-primary">
              {{ recommendations.length }}
            </div>
            <div class="mt-2 text-sm leading-6 text-text-secondary">
              可直接推进的建议题目，建议按顺序完成前排任务。
            </div>
          </AppCard>
        </div>

        <template #footer>
          <div class="flex flex-wrap gap-3">
            <ElButton type="primary" @click="emit('openChallenges')">打开挑战列表</ElButton>
            <ElButton plain @click="emit('openSkillProfile')">回看能力画像</ElButton>
          </div>
        </template>
      </AppCard>

      <SectionCard
        title="推荐队列"
        subtitle="按当前训练阶段排序的优先挑战，建议按顺序推进，先消化队列前排任务。"
      >
        <div
          v-if="recommendations.length === 0"
          class="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-text-secondary"
        >
          当前没有推荐题目，可以先去挑战列表探索新的方向。
        </div>

        <div v-else class="grid gap-4">
          <AppCard
            v-for="(item, index) in recommendations.slice(0, 3)"
            :key="item.challenge_id"
            variant="metric"
            accent="warning"
            :eyebrow="`Queue ${index + 1}`"
            :title="item.title"
            :subtitle="item.reason"
          >
            <template #header>
              <div
                class="flex h-11 w-11 items-center justify-center rounded-2xl border border-amber-500/18 bg-amber-500/10 text-amber-600"
              >
                <Crosshair class="h-4 w-4" />
              </div>
            </template>
            <div
              class="flex flex-wrap items-center gap-2 text-xs uppercase tracking-[0.16em] text-text-muted"
            >
              <span>{{ item.category }}</span>
              <span class="h-1 w-1 rounded-full bg-border" />
              <span
                class="rounded-full px-2.5 py-1 text-xs font-medium normal-case tracking-normal"
                :class="difficultyClass(item.difficulty)"
              >
                {{ difficultyLabel(item.difficulty) }}
              </span>
            </div>
          </AppCard>
        </div>
      </SectionCard>
    </section>

    <SectionCard title="推荐列表" subtitle="完整推荐项保留推荐原因和进入动作，适合直接逐条推进。">
      <div
        v-if="recommendations.length === 0"
        class="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-text-secondary"
      >
        当前没有推荐题目，可以先去挑战列表探索新的方向。
      </div>

      <div v-else class="space-y-4">
        <AppCard
          v-for="(item, index) in recommendations"
          :key="item.challenge_id"
          as="button"
          variant="action"
          accent="warning"
          interactive
          class="cursor-pointer text-left"
          @click="emit('openChallenge', item.challenge_id)"
        >
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-amber-500/18 bg-amber-500/10 text-base font-semibold text-amber-600"
              >
                {{ index + 1 }}
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <p class="text-lg font-semibold text-text-primary">{{ item.title }}</p>
                  <span
                    class="rounded-full px-2.5 py-1 text-xs font-medium"
                    :class="difficultyClass(item.difficulty)"
                  >
                    {{ difficultyLabel(item.difficulty) }}
                  </span>
                </div>
                <div
                  class="mt-2 flex items-center gap-2 text-xs uppercase tracking-[0.16em] text-text-muted"
                >
                  <Crosshair class="h-3.5 w-3.5" />
                  <span>{{ item.category }}</span>
                  <span class="h-1 w-1 rounded-full bg-border" />
                  <span>Queue {{ index + 1 }}</span>
                </div>
              </div>
            </div>
            <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-amber-500" />
          </div>
          <div class="mt-4 rounded-2xl border border-border-subtle bg-base/40 px-4 py-4">
            <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
              <Sparkles class="h-4 w-4 text-amber-500" />
              推荐理由
            </div>
            <p class="mt-2 text-sm leading-6 text-text-secondary">{{ item.reason }}</p>
          </div>
        </AppCard>
      </div>
    </SectionCard>
  </div>
</template>
