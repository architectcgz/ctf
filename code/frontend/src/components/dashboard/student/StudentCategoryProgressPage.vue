<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpRight, Gauge, MoveRight } from 'lucide-vue-next'

import { progressRate } from './utils'

interface CategoryStat {
  category: string
  total: number
  solved: number
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
  <section
    class="journal-soft-surface space-y-6 flex min-h-full flex-1 flex-col"
    :class="
      embedded
        ? 'journal-shell-embedded'
        : 'journal-shell journal-hero rounded-[30px] border px-6 py-6 md:px-8'
    "
  >
    <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
      <div>
        <div class="journal-eyebrow">Coverage Overview</div>
        <h1 class="journal-page-title workspace-tab-heading__title text-[var(--journal-ink)]">
          分类覆盖概况
        </h1>
        <p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          看各分类的完成情况和当前强弱项。
        </p>
        <div class="mt-6 flex flex-wrap gap-3" role="group" aria-label="分类进度快捷操作">
          <button class="journal-btn-primary" @click="emit('openChallenges')">去训练</button>
          <button class="journal-btn-outline" @click="emit('openSkillProfile')">能力画像</button>
        </div>
      </div>

      <article class="journal-brief rounded-[24px] border px-5 py-5">
        <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
          <Gauge class="h-5 w-5 text-[var(--journal-accent)]" />
          结构快照
        </div>
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
          <div class="journal-note">
            <div class="journal-note-label">整体覆盖率</div>
            <div class="journal-note-value">{{ completionRate }}%</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">分类数量</div>
            <div class="journal-note-value">{{ rankedCategories.length }} 类</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">当前强项</div>
            <div class="journal-note-value">{{ strongestCategory?.category || '-' }}</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">当前短板</div>
            <div class="journal-note-value">{{ weakestCategory?.category || '-' }}</div>
          </div>
        </div>
      </article>
    </div>

    <div
      class="category-board mt-6 px-1 pt-5 md:px-2 md:pt-6"
      :class="{ 'category-board--embedded': embedded }"
    >
      <section class="category-section">
        <div class="grid gap-6 xl:grid-cols-2">
          <article class="category-highlight">
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="journal-eyebrow journal-eyebrow-soft">Strongest Direction</div>
                <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">
                  {{ strongestCategory?.category || '-' }}
                </h3>
              </div>
              <div class="direction-icon direction-icon--success">
                <ArrowUpRight class="h-4 w-4" />
              </div>
            </div>
            <div v-if="strongestCategory" class="mt-4">
              <div class="flex items-end justify-between text-sm">
                <span class="text-[var(--journal-muted)]">完成进度</span>
                <span class="font-semibold text-[var(--journal-ink)]"
                  >{{ strongestCategory.solved }} / {{ strongestCategory.total }}</span
                >
              </div>
              <div class="category-track mt-2 h-2.5 rounded-full">
                <div
                  class="h-2.5 rounded-full bg-emerald-500"
                  :style="{ width: `${strongestCategory.rate}%` }"
                />
              </div>
              <div class="mt-2 text-right text-xs font-semibold text-emerald-600">
                {{ strongestCategory.rate }}%
              </div>
            </div>
          </article>

          <article class="category-highlight">
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="journal-eyebrow journal-eyebrow-soft">Weakest Direction</div>
                <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">
                  {{ weakestCategory?.category || '-' }}
                </h3>
              </div>
              <div class="direction-icon direction-icon--warning">
                <Gauge class="h-4 w-4" />
              </div>
            </div>
            <div v-if="weakestCategory" class="mt-4">
              <div class="flex items-end justify-between text-sm">
                <span class="text-[var(--journal-muted)]">完成进度</span>
                <span class="font-semibold text-[var(--journal-ink)]"
                  >{{ weakestCategory.solved }} / {{ weakestCategory.total }}</span
                >
              </div>
              <div class="category-track mt-2 h-2.5 rounded-full">
                <div
                  class="h-2.5 rounded-full bg-amber-400"
                  :style="{ width: `${weakestCategory.rate}%` }"
                />
              </div>
              <div class="mt-2 text-right text-xs font-semibold text-amber-600">
                {{ weakestCategory.rate }}%
              </div>
            </div>
          </article>
        </div>
      </section>

      <section class="category-section">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="journal-eyebrow journal-eyebrow-soft">Category Board</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">分类进度板</h3>
          </div>
          <button class="journal-btn-outline" @click="emit('openChallenges')">
            <MoveRight class="h-3.5 w-3.5" />
            去训练
          </button>
        </div>

        <div
          v-if="rankedCategories.length === 0"
          class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-shell-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
        >
          当前还没有分类统计数据，先完成几道题再回来查看。
        </div>

        <div v-else class="category-list mt-5">
          <div class="category-head" aria-hidden="true">
            <span>分类</span>
            <span>完成率</span>
            <span>已解 / 总量</span>
          </div>
          <div v-for="item in rankedCategories" :key="item.category" class="category-item">
            <div class="category-row">
              <span class="category-row__name">{{ item.category }}</span>
              <span class="category-row__rate">{{ item.rate }}%</span>
              <span class="category-row__count">{{ item.solved }}/{{ item.total }}</span>
            </div>
            <div class="category-track mt-3 h-2 rounded-full">
              <div
                class="category-track-fill h-2 rounded-full"
                :style="{ width: `${item.rate}%` }"
              />
            </div>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.journal-soft-surface {
  --category-cols: minmax(130px, 1fr) minmax(88px, 120px) minmax(110px, 150px);
  --journal-soft-button-height: 34px;
  --journal-soft-button-padding: var(--space-2) var(--space-4);
  --journal-soft-button-size: 0.82rem;
  --journal-soft-button-primary-background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  --journal-soft-button-primary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 14%,
    transparent
  );
  --journal-soft-button-primary-color: var(--journal-accent);
}

.journal-brief {
  border-color: var(--journal-shell-border);
  background: var(--journal-surface-subtle);
}

.category-board {
  border-top: 1px solid var(--journal-divider);
}

.category-board--embedded {
  margin-top: var(--space-5);
}

.category-section + .category-section {
  margin-top: var(--space-6);
  padding-top: var(--space-6);
  border-top: 1px solid var(--journal-divider);
}

.category-highlight,
.category-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-4) var(--space-4-5);
}

.category-head {
  display: grid;
  grid-template-columns: var(--category-cols);
  gap: var(--space-2);
  padding-bottom: var(--space-3);
  font-size: var(--font-size-0-69);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.category-head > :nth-child(2),
.category-head > :nth-child(3) {
  text-align: right;
}

.category-item + .category-item {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--journal-divider);
}

.category-track {
  background: var(--journal-track);
}

.category-track-fill {
  background: color-mix(in srgb, var(--journal-accent) 68%, #0ea5e9);
}

.category-row {
  display: grid;
  grid-template-columns: var(--category-cols);
  gap: var(--space-2);
  align-items: center;
}

.category-row__name {
  font-size: var(--font-size-0-86);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-ink);
}

.category-row__rate,
.category-row__count {
  text-align: right;
}

.category-row__rate {
  font-size: var(--font-size-0-90);
  font-weight: 700;
  color: var(--journal-ink);
}

.category-row__count {
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.direction-icon {
  display: flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.9rem;
}

.direction-icon--success {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}

.direction-icon--warning {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
}

:global([data-theme='dark']) .category-highlight,
:global([data-theme='dark']) .category-list {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

:global([data-theme='dark']) .journal-soft-surface {
  --journal-soft-button-outline-background: color-mix(
    in srgb,
    var(--journal-surface) 94%,
    transparent
  );
}

@media (max-width: 767px) {
  .journal-soft-surface {
    --category-cols: minmax(100px, 1fr) minmax(68px, 84px) minmax(92px, 116px);
    --journal-soft-button-height: 36px;
  }
}
</style>
