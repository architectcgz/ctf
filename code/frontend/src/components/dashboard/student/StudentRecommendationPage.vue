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
      <div class="rounded-[30px] border border-amber-500/20 bg-[linear-gradient(145deg,rgba(120,53,15,0.48),rgba(15,23,42,0.92))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-100/75">Targeted Training</div>
        <h2 class="mt-3 text-3xl font-semibold tracking-tight text-white">补短板计划</h2>
        <p class="mt-3 text-sm leading-7 text-amber-50/80">
          根据当前薄弱维度给出优先训练顺序，建议先完成靠前题目，再回看能力画像确认是否抬升。
        </p>
        <div class="mt-6 rounded-[24px] border border-white/10 bg-white/6 px-5 py-5">
          <div class="flex items-center gap-2 text-sm font-medium text-white">
            <ShieldAlert class="h-4 w-4 text-amber-200" />
            优先修复的能力维度
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <span
              v-for="item in weakDimensions.slice(0, 4)"
              :key="item"
              class="rounded-full border border-amber-300/15 bg-amber-300/12 px-3 py-1 text-xs font-medium text-amber-100"
            >
              {{ item }}
            </span>
            <span
              v-if="weakDimensions.length === 0"
              class="rounded-full border border-emerald-300/15 bg-emerald-300/12 px-3 py-1 text-xs font-medium text-emerald-100"
            >
              暂无明显短板
            </span>
          </div>
          <div class="mt-5 text-sm text-amber-50/75">当前首要关注：{{ headline }}</div>
        </div>
        <div class="mt-6 flex flex-wrap gap-3">
          <ElButton type="primary" @click="emit('openChallenges')">打开挑战列表</ElButton>
          <ElButton plain @click="emit('openSkillProfile')">回看能力画像</ElButton>
        </div>
      </div>

      <div class="rounded-[30px] border border-amber-500/20 bg-[linear-gradient(145deg,rgba(59,24,3,0.78),rgba(15,23,42,0.94))] p-5 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b border-white/10 pb-4">
          <div>
            <div class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-100/75">Recommendation Queue</div>
            <h3 class="mt-2 text-2xl font-semibold tracking-tight text-white">推荐队列</h3>
            <p class="mt-2 text-sm leading-6 text-amber-50/78">
              按当前训练阶段排序的优先挑战，建议按顺序推进，先消化队列前排任务。
            </p>
          </div>
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-3 text-right">
            <div class="text-[11px] font-semibold uppercase tracking-[0.18em] text-amber-100/60">当前队列</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ recommendations.length }}</div>
            <div class="mt-1 text-xs text-amber-50/70">可直接推进的建议题目</div>
          </div>
        </div>

        <div v-if="recommendations.length === 0" class="mt-5 rounded-[24px] border border-dashed border-white/10 bg-white/5 px-4 py-12 text-center text-sm text-amber-50/72">
          当前没有推荐题目，可以先去挑战列表探索新的方向。
        </div>

        <div v-else class="mt-5 space-y-4">
          <AppCard
            v-for="(item, index) in recommendations"
            :key="item.challenge_id"
            as="button"
            variant="action"
            accent="warning"
            interactive
            class="cursor-pointer border-white/10 bg-white/5 text-left hover:border-amber-300/30"
            @click="emit('openChallenge', item.challenge_id)"
          >
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="flex items-start gap-4">
                <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-amber-300/15 bg-amber-300/12 text-base font-semibold text-amber-100">
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
                    <span class="h-1 w-1 rounded-full bg-white/20" />
                    <span>Queue {{ index + 1 }}</span>
                  </div>
                </div>
              </div>
              <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-amber-200" />
            </div>
            <div class="mt-4 rounded-2xl border border-white/5 bg-black/10 px-4 py-4">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Sparkles class="h-4 w-4 text-amber-200" />
                推荐理由
              </div>
              <p class="mt-2 text-sm leading-6 text-text-secondary">{{ item.reason }}</p>
            </div>
          </AppCard>
        </div>
      </div>
    </section>
  </div>
</template>
