<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpRight, Gauge, MoveRight, Radar } from 'lucide-vue-next'

import SectionCard from '@/components/common/SectionCard.vue'

import { progressRate } from './utils'

interface CategoryStat {
  category: string
  total: number
  solved: number
}

const props = defineProps<{
  categoryStats: CategoryStat[]
  completionRate: number
}>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
}>()

const rankedCategories = computed(() =>
  [...props.categoryStats]
    .map((item) => ({
      ...item,
      rate: progressRate(item.total, item.solved),
    }))
    .sort((left, right) => right.rate - left.rate),
)

const strongestCategory = computed(() => rankedCategories.value[0] || null)
const weakestCategory = computed(() => rankedCategories.value.at(-1) || null)
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 xl:grid-cols-[0.7fr_1.3fr]">
      <div class="grid gap-4">
        <article class="rounded-[28px] border border-border bg-[linear-gradient(180deg,rgba(8,145,178,0.16),rgba(15,23,42,0.86))] p-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
            <Radar class="h-4 w-4 text-primary" />
            覆盖概况
          </div>
          <div class="mt-4 text-4xl font-semibold tracking-tight text-text-primary">{{ completionRate }}%</div>
          <div class="mt-2 text-sm leading-6 text-text-secondary">当前总训练覆盖率。这个页面专门聚焦不同题型的完成深度，而不是继续复用主页摘要。</div>
        </article>

        <article class="rounded-[28px] border border-border bg-surface/88 p-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
            <ArrowUpRight class="h-4 w-4 text-emerald-300" />
            最强方向
          </div>
          <div class="mt-4 text-2xl font-semibold text-text-primary">{{ strongestCategory?.category || '-' }}</div>
          <div class="mt-2 text-sm text-text-secondary">{{ strongestCategory ? `${strongestCategory.solved}/${strongestCategory.total}，完成 ${strongestCategory.rate}%` : '暂无数据' }}</div>
        </article>

        <article class="rounded-[28px] border border-border bg-surface/88 p-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
            <Gauge class="h-4 w-4 text-amber-300" />
            最弱方向
          </div>
          <div class="mt-4 text-2xl font-semibold text-text-primary">{{ weakestCategory?.category || '-' }}</div>
          <div class="mt-2 text-sm text-text-secondary">{{ weakestCategory ? `${weakestCategory.solved}/${weakestCategory.total}，完成 ${weakestCategory.rate}%` : '暂无数据' }}</div>
        </article>
      </div>

      <SectionCard title="分类进度板" subtitle="把不同题型拆成独立进度轨道，方便快速识别训练结构是否均衡。">
        <div v-if="rankedCategories.length === 0" class="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-text-secondary">
          当前还没有分类统计数据，先完成几道题再回来查看。
        </div>

        <div v-else class="space-y-4">
          <article
            v-for="item in rankedCategories"
            :key="item.category"
            class="rounded-[24px] border border-border bg-base/70 px-5 py-5"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <div class="text-xs font-semibold uppercase tracking-[0.18em] text-text-muted">{{ item.category }}</div>
                <div class="mt-2 text-xl font-semibold text-text-primary">{{ item.rate }}%</div>
              </div>
              <div class="text-right text-sm text-text-secondary">
                <div>{{ item.solved }} / {{ item.total }}</div>
                <div class="mt-1 text-xs uppercase tracking-[0.14em] text-text-muted">完成题数</div>
              </div>
            </div>
            <div class="mt-4 h-3 rounded-full bg-[var(--color-bg-base)]">
              <div
                class="h-3 rounded-full bg-[linear-gradient(90deg,rgba(34,211,238,0.95),rgba(56,189,248,0.72))]"
                :style="{ width: `${item.rate}%` }"
              />
            </div>
          </article>
        </div>

        <template #header>
          <div class="flex flex-wrap gap-2">
            <ElButton plain @click="emit('openSkillProfile')">能力画像</ElButton>
            <ElButton type="primary" @click="emit('openChallenges')">继续训练</ElButton>
          </div>
        </template>
      </SectionCard>
    </section>

    <section class="grid gap-4 md:grid-cols-2">
      <SectionCard title="建议动作" subtitle="先补最弱方向，再拉平整体训练结构。">
        <div class="space-y-3">
          <div class="rounded-[22px] border border-border bg-base/70 px-4 py-4 text-sm leading-6 text-text-secondary">
            如果最近一段时间只刷熟悉题型，整体完成率会上升，但结构会失衡。建议优先补 weakest category，再回到强项巩固。
          </div>
          <button
            type="button"
            class="flex w-full items-center justify-between rounded-[22px] border border-border bg-base/70 px-4 py-4 text-left transition hover:border-primary/50"
            @click="emit('openChallenges')"
          >
            <span class="text-sm font-medium text-text-primary">打开挑战列表，按短板方向继续练习</span>
            <MoveRight class="h-4 w-4 text-primary" />
          </button>
        </div>
      </SectionCard>

      <SectionCard title="结构判断" subtitle="看的是训练面，而不只是总分。">
        <div class="grid gap-3 md:grid-cols-2">
          <div class="rounded-[22px] border border-emerald-500/20 bg-emerald-500/8 px-4 py-4">
            <div class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-200">当前强项</div>
            <div class="mt-2 text-lg font-semibold text-text-primary">{{ strongestCategory?.category || '暂无数据' }}</div>
          </div>
          <div class="rounded-[22px] border border-amber-500/20 bg-amber-500/8 px-4 py-4">
            <div class="text-xs font-semibold uppercase tracking-[0.16em] text-amber-200">当前短板</div>
            <div class="mt-2 text-lg font-semibold text-text-primary">{{ weakestCategory?.category || '暂无数据' }}</div>
          </div>
        </div>
      </SectionCard>
    </section>
  </div>
</template>
