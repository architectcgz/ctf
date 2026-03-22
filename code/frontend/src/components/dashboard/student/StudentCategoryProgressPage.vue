<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpRight, Gauge, MoveRight, Radar } from 'lucide-vue-next'

import AppCard from '@/components/common/AppCard.vue'
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
    .sort((left, right) => right.rate - left.rate)
)

const strongestCategory = computed(() => rankedCategories.value[0] || null)
const weakestCategory = computed(() => rankedCategories.value.at(-1) || null)
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 xl:grid-cols-[0.7fr_1.3fr]">
      <div class="grid gap-4">
        <AppCard
          class="category-overview-card"
          variant="hero"
          accent="primary"
          eyebrow="Coverage Overview"
          title="覆盖概况"
          subtitle="从整体覆盖率判断训练结构，再对照分类进度决定下一步优先补强的方向。"
        >
          <div class="grid gap-4 md:grid-cols-[auto_1fr] md:items-end">
            <div class="text-5xl font-semibold tracking-tight text-text-primary">
              {{ completionRate }}%
            </div>
            <AppCard class="category-overview-note" variant="action" accent="primary">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Radar class="h-4 w-4 text-primary" />
                覆盖率代表已完成题目在全部分类题量中的占比
              </div>
              <div class="mt-2 text-sm leading-6 text-text-secondary">
                先看覆盖率，再看强弱方向，最后回到分类进度板逐项补齐。
              </div>
            </AppCard>
          </div>
        </AppCard>

        <AppCard
          variant="metric"
          accent="success"
          eyebrow="Strongest Direction"
          :title="strongestCategory?.category || '-'"
          :subtitle="
            strongestCategory
              ? `${strongestCategory.solved}/${strongestCategory.total}，完成 ${strongestCategory.rate}%`
              : '暂无数据'
          "
        >
          <template #header>
            <div class="category-direction-icon category-direction-icon--success">
              <ArrowUpRight class="h-4 w-4" />
            </div>
          </template>
        </AppCard>

        <AppCard
          variant="metric"
          accent="warning"
          eyebrow="Weakest Direction"
          :title="weakestCategory?.category || '-'"
          :subtitle="
            weakestCategory
              ? `${weakestCategory.solved}/${weakestCategory.total}，完成 ${weakestCategory.rate}%`
              : '暂无数据'
          "
        >
          <template #header>
            <div class="category-direction-icon category-direction-icon--warning">
              <Gauge class="h-4 w-4" />
            </div>
          </template>
        </AppCard>
      </div>

      <SectionCard
        title="分类进度板"
        subtitle="把不同题型拆成独立进度轨道，方便快速识别训练结构是否均衡。"
      >
        <div
          v-if="rankedCategories.length === 0"
          class="rounded-2xl border border-dashed border-border px-4 py-12 text-center text-sm text-text-secondary"
        >
          当前还没有分类统计数据，先完成几道题再回来查看。
        </div>

        <div v-else class="space-y-4">
          <AppCard
            v-for="item in rankedCategories"
            :key="item.category"
            variant="action"
            accent="neutral"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <div class="text-xs font-semibold uppercase tracking-[0.18em] text-text-muted">
                  {{ item.category }}
                </div>
                <div class="mt-2 text-xl font-semibold text-text-primary">{{ item.rate }}%</div>
              </div>
              <div class="text-right text-sm text-text-secondary">
                <div>{{ item.solved }} / {{ item.total }}</div>
                <div class="mt-1 text-xs uppercase tracking-[0.14em] text-text-muted">完成题数</div>
              </div>
            </div>
            <div class="category-progress-track mt-4 h-3 rounded-full">
              <div
                class="h-3 rounded-full bg-[linear-gradient(90deg,rgba(34,211,238,0.95),rgba(56,189,248,0.72))]"
                :style="{ width: `${item.rate}%` }"
              />
            </div>
          </AppCard>
        </div>
      </SectionCard>
    </section>

    <section class="grid gap-4 md:grid-cols-2">
      <SectionCard title="建议动作" subtitle="先补最弱方向，再拉平整体训练结构。">
        <div class="space-y-3">
          <AppCard variant="action" accent="neutral">
            如果最近一段时间只刷熟悉题型，整体完成率会上升，但结构会失衡。建议优先补 weakest
            category，再回到强项巩固。
          </AppCard>
          <AppCard
            as="button"
            variant="action"
            accent="primary"
            interactive
            class="flex w-full items-center justify-between text-left"
            @click="emit('openChallenges')"
          >
            <span class="text-sm font-medium text-text-primary"
              >打开挑战列表，按短板方向继续练习</span
            >
            <MoveRight class="h-4 w-4 text-primary" />
          </AppCard>
        </div>
      </SectionCard>

      <SectionCard title="结构判断" subtitle="看的是训练面，而不只是总分。">
        <div class="grid gap-3 md:grid-cols-2">
          <AppCard
            variant="metric"
            accent="success"
            eyebrow="当前强项"
            :title="strongestCategory?.category || '暂无数据'"
          />
          <AppCard
            variant="metric"
            accent="warning"
            eyebrow="当前短板"
            :title="weakestCategory?.category || '暂无数据'"
          />
        </div>
      </SectionCard>
    </section>
  </div>
</template>

<style scoped>
.category-overview-card {
  border-color: color-mix(
    in srgb,
    var(--color-primary) 14%,
    var(--color-border-default)
  ) !important;
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--color-primary) 6%, var(--color-bg-surface)),
    color-mix(in srgb, var(--color-bg-surface) 98%, var(--color-bg-base))
  ) !important;
}

.category-overview-note {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-primary) 7%, var(--color-bg-surface)),
    color-mix(in srgb, var(--color-bg-surface) 99%, var(--color-bg-base))
  ) !important;
}

.category-direction-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
}

.category-direction-icon--success {
  color: var(--color-success);
  border-color: color-mix(in srgb, var(--color-success) 18%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-success) 10%, var(--color-bg-surface));
}

.category-direction-icon--warning {
  color: var(--color-warning);
  border-color: color-mix(in srgb, var(--color-warning) 18%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-warning) 10%, var(--color-bg-surface));
}

.category-progress-track {
  background: color-mix(in srgb, var(--color-border-subtle) 60%, transparent);
}

:global([data-theme='light']) .category-overview-card {
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--color-primary) 3%, white),
    color-mix(in srgb, #f8fafc 98%, white)
  ) !important;
}

:global([data-theme='light']) .category-overview-note {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-primary) 4%, white),
    color-mix(in srgb, #f8fafc 99%, white)
  ) !important;
}
</style>
