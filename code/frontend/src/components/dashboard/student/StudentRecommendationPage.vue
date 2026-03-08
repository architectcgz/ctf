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
        subtitle="这页不再复用主页模板，而是只围绕推荐队列展开。建议先完成序号靠前的题目，再回到能力画像确认维度是否抬升。"
      >
        <AppCard variant="action" accent="warning">
          <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
            <ShieldAlert class="h-4 w-4" style="color: var(--color-warning);" />
            优先修复的能力维度
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <span
              v-for="item in weakDimensions.slice(0, 4)"
              :key="item"
              class="rounded-full border px-3 py-1 text-xs font-medium"
              style="border-color: rgba(210,153,34,0.2); background-color: rgba(210,153,34,0.1); color: color-mix(in srgb, white 86%, var(--color-warning));"
            >
              {{ item }}
            </span>
            <span
              v-if="weakDimensions.length === 0"
              class="rounded-full border px-3 py-1 text-xs font-medium"
              style="border-color: rgba(63,185,80,0.2); background-color: rgba(63,185,80,0.1); color: color-mix(in srgb, white 86%, var(--color-success));"
            >
              暂无明显短板
            </span>
          </div>
          <div class="mt-5 text-sm text-text-secondary">当前首要关注：{{ headline }}</div>
        </AppCard>

        <template #footer>
          <div class="flex flex-wrap gap-3">
            <ElButton type="primary" @click="emit('openChallenges')">打开挑战列表</ElButton>
            <ElButton plain @click="emit('openSkillProfile')">回看能力画像</ElButton>
          </div>
        </template>
      </AppCard>

      <SectionCard title="推荐队列" subtitle="按当前训练阶段排序的优先挑战，建议按顺序推进。">
        <div v-if="recommendations.length === 0" class="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-text-secondary">
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
                <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-primary/12 text-base font-semibold text-primary">
                  {{ index + 1 }}
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <p class="text-lg font-semibold text-text-primary">{{ item.title }}</p>
                    <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="difficultyClass(item.difficulty)">
                      {{ difficultyLabel(item.difficulty) }}
                    </span>
                  </div>
                  <div class="mt-2 flex items-center gap-2 text-xs uppercase tracking-[0.16em] text-text-muted">
                    <Crosshair class="h-3.5 w-3.5" />
                    <span>{{ item.category }}</span>
                  </div>
                </div>
              </div>
              <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-primary" />
            </div>
            <div class="mt-4 rounded-2xl border border-white/5 bg-white/4 px-4 py-4">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Sparkles class="h-4 w-4 text-amber-200" />
                推荐理由
              </div>
              <p class="mt-2 text-sm leading-6 text-text-secondary">{{ item.reason }}</p>
            </div>
          </AppCard>
        </div>
      </SectionCard>
    </section>
  </div>
</template>
