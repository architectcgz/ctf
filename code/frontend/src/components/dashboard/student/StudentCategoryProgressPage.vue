<script setup lang="ts">
import { computed } from 'vue'

import { rankCategoryActionItems } from './utils'

interface CategoryStat {
  category: string
  total: number
  solved: number
}

interface RankedCategoryStat extends CategoryStat {
  rate: number
  remaining: number
}

const props = withDefaults(
  defineProps<{
    categoryStats: CategoryStat[]
    completionRate: number
    embedded?: boolean
  }>(),
  {
    embedded: false,
  }
)

const emit = defineEmits<{
  openChallenges: []
  openCategoryChallenges: [category: string]
  openSkillProfile: []
}>()

const rankedCategories = computed<RankedCategoryStat[]>(() =>
  rankCategoryActionItems(props.categoryStats)
)
const primaryCategory = computed(() => rankedCategories.value[0] || null)
const hasCategoryStats = computed(() => rankedCategories.value.length > 0)
const headlineTitle = computed(() =>
  primaryCategory.value
    ? `优先补这个分类：${primaryCategory.value.category}`
    : '先开始积累分类覆盖面'
)
const summaryCards = computed(() => [
  {
    key: 'focus',
    label: '当前待补题量',
    value: primaryCategory.value ? `${primaryCategory.value.remaining} 道` : '待生成',
    helper: primaryCategory.value
      ? `${primaryCategory.value.category} 还有 ${primaryCategory.value.remaining} 道题待补，先从这里补回训练短板。`
      : '先完成几道题，这里会自动形成下一步最值得先补的分类。',
  },
  {
    key: 'coverage',
    label: '整体覆盖率',
    value: `${props.completionRate}%`,
    helper:
      rankedCategories.value.length > 0
        ? '按当前分类题量加权后的总体进度，方便判断覆盖面是否在稳定扩大。'
        : '当前还没有足够的分类数据来衡量整体覆盖率。',
  },
  {
    key: 'ranking',
    label: '当前排序依据',
    value: rankedCategories.value.length > 0 ? `${rankedCategories.value.length} 类` : '待生成',
    helper:
      rankedCategories.value.length > 0
        ? '先按完成率找短板，再用题量打破并列，避免样本太小的分类抢到最前面。'
        : '分类进度会在完成训练后自动更新。',
  },
])

function categoryActionCopy(item: RankedCategoryStat, index: number): string {
  if (item.remaining <= 0) {
    return '这一类已经补齐，可以放到后面用于维持训练手感。'
  }

  if (index === 0) {
    return `还有 ${item.remaining} 道题待补，当前完成率 ${item.rate}%，这组最适合先拉回到稳定区间。`
  }

  return `还差 ${item.remaining} / ${item.total}，做完前面的分类后，可以继续沿这条线补齐覆盖面。`
}

function openPrimaryCategory(): void {
  if (primaryCategory.value) {
    emit('openCategoryChallenges', primaryCategory.value.category)
    return
  }
  emit('openChallenges')
}
</script>

<template>
  <section
    class="journal-soft-surface space-y-6 flex min-h-full flex-1 flex-col"
    :class="
      embedded
        ? 'journal-shell-embedded'
        : 'workspace-shell journal-shell journal-shell-user journal-hero'
    "
  >
    <div :class="embedded ? undefined : 'content-pane'">
      <div class="category-header">
      <div class="workspace-overline">Action Ranking</div>
      <h1 class="journal-page-title workspace-page-title text-[var(--journal-ink)]">
        {{ headlineTitle }}
      </h1>
      <p class="workspace-page-copy max-w-2xl">
        {{
          hasCategoryStats
            ? '按分类找短板，先补当前最需要回填的那一类。'
            : '先完成几道题，这里会自动排出下一步最该补的分类。'
        }}
      </p>

      <div class="mt-5 flex flex-wrap gap-3" role="group" aria-label="分类进度快捷操作">
        <button class="journal-btn-primary" @click="openPrimaryCategory">
          {{ primaryCategory ? `先去 ${primaryCategory.category}` : '去训练' }}
        </button>
        <button class="journal-btn-outline" @click="emit('openChallenges')">浏览全部题目</button>
        <button class="journal-btn-outline" @click="emit('openSkillProfile')">能力画像</button>
      </div>

      <div
        class="category-summary-strip mt-5 progress-strip metric-panel-grid metric-panel-default-surface"
      >
        <article
          v-for="card in summaryCards"
          :key="card.key"
          class="category-summary-card progress-card metric-panel-card"
        >
          <div class="journal-note-label progress-card-label metric-panel-label">
            {{ card.label }}
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ card.value }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            {{ card.helper }}
          </div>
        </article>
      </div>
      </div>

      <div
        class="category-board mt-6 px-1 pt-5 md:px-2 md:pt-6"
        :class="{ 'category-board--embedded': embedded }"
      >
        <section class="category-section">
        <div v-if="rankedCategories.length > 0" class="category-toolbar">
          <p class="category-toolbar__copy">从排序最前的分类开始，完成一类再继续往后推。</p>
        </div>

        <div
          v-if="rankedCategories.length === 0"
          class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-shell-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
        >
          当前还没有分类统计数据，先完成几道题再回来查看。
        </div>

        <div v-else class="category-action-list mt-5">
          <article
            v-for="(item, index) in rankedCategories"
            :key="item.category"
            class="category-action-item"
            :data-test="`category-action-${item.category}`"
          >
            <div class="category-action-item__body">
              <div class="category-action-rank">
                {{ index + 1 }}
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="category-action-item__name">{{ item.category }}</span>
                  <span class="category-action-item__rate">{{ item.rate }}%</span>
                  <span class="category-action-item__count"
                    >{{ item.solved }}/{{ item.total }}</span
                  >
                </div>
                <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                  {{ categoryActionCopy(item, index) }}
                </p>
              </div>
              <button
                type="button"
                class="journal-btn-primary category-action-item__cta"
                @click="emit('openCategoryChallenges', item.category)"
              >
                去做这个分类
              </button>
            </div>

            <div class="category-track mt-4 h-2 rounded-full">
              <div
                class="category-track-fill h-2 rounded-full"
                :style="{ width: `${item.rate}%` }"
              />
            </div>
          </article>
        </div>
        </section>
      </div>
    </div>
  </section>
</template>

<style scoped>
.journal-soft-surface {
  --journal-soft-button-height: 34px;
  --journal-soft-button-padding: var(--space-2) var(--space-4);
  --journal-soft-button-size: 0.82rem;
  --journal-soft-button-primary-background: color-mix(
    in srgb,
    var(--journal-accent) 8%,
    transparent
  );
  --journal-soft-button-primary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 14%,
    transparent
  );
  --journal-soft-button-primary-color: var(--journal-accent);
}

.category-header {
  display: grid;
  gap: var(--space-3);
}

.category-summary-strip {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.category-summary-strip.metric-panel-default-surface {
  --metric-panel-border: var(--journal-soft-border);
  --metric-panel-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 12%, transparent),
      transparent 42%
    ),
    linear-gradient(
      165deg,
      color-mix(in srgb, var(--journal-surface-subtle) 92%, transparent),
      color-mix(in srgb, var(--journal-surface) 96%, transparent)
    );
  --metric-panel-shadow: 0 10px 20px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent);
}

.category-board {
  border-top: 1px solid var(--journal-divider);
}

.category-board--embedded {
  margin-top: var(--space-5);
}

.category-toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.category-toolbar__copy {
  margin: 0;
  font-size: var(--font-size-0-82);
  line-height: 1.7;
  color: var(--journal-muted);
}

.category-action-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-4) var(--space-4-5);
}

.category-action-item + .category-action-item {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--journal-divider);
}

.category-action-item__body {
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
}

.category-action-rank {
  display: flex;
  min-width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-accent-strong);
  font-size: var(--font-size-0-82);
  font-weight: 700;
}

.category-action-item__name {
  font-size: var(--font-size-0-86);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-ink);
}

.category-action-item__rate,
.category-action-item__count {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.2rem 0.55rem;
  font-size: var(--font-size-0-74);
  font-weight: 600;
}

.category-action-item__rate {
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-accent-strong);
}

.category-action-item__count {
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, transparent);
  color: var(--journal-muted);
}

.category-action-item__cta {
  flex-shrink: 0;
}

.category-track {
  background: var(--journal-track);
}

.category-track-fill {
  background: color-mix(in srgb, var(--journal-accent) 68%, var(--color-primary-hover));
}

:global([data-theme='dark']) .category-action-list {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

:global([data-theme='dark']) .journal-soft-surface {
  --journal-soft-button-outline-background: color-mix(
    in srgb,
    var(--journal-surface) 94%,
    transparent
  );
}

@media (max-width: 900px) {
  .category-summary-strip {
    --metric-panel-columns: 1fr;
  }

  .category-action-item__body {
    flex-direction: column;
  }

  .category-action-item__cta {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 767px) {
  .journal-soft-surface {
    --journal-soft-button-height: 36px;
  }
}
</style>
