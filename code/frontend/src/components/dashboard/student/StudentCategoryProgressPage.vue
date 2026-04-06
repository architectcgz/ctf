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
    class="space-y-6 flex min-h-full flex-1 flex-col"
    :class="
      embedded
        ? 'journal-shell-embedded'
        : 'journal-shell journal-hero rounded-[30px] border px-6 py-6 md:px-8'
    "
  >
    <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Coverage Overview</div>
          <h1 class="journal-page-title mt-3 text-[var(--journal-ink)]">
            分类覆盖概况
          </h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
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

    <div class="category-board mt-6 px-1 pt-5 md:px-2 md:pt-6" :class="{ 'category-board--embedded': embedded }">
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
            <div
              v-for="item in rankedCategories"
              :key="item.category"
              class="category-item"
            >
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
.journal-shell-embedded,
.journal-shell {
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--color-border-default) 70%, transparent);
  --journal-control-border: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --journal-divider: color-mix(in srgb, var(--color-border-default) 64%, transparent);
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  --category-cols: minmax(130px, 1fr) minmax(88px, 120px) minmax(110px, 150px);
  font-family:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
}

.journal-shell-embedded {
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.journal-hero {
  border-color: var(--journal-shell-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-page-title {
  font-size: clamp(32px, 4vw, 46px);
  line-height: 1.02;
  letter-spacing: -0.04em;
  font-weight: 600;
}

.journal-brief {
  border-color: var(--journal-shell-border);
  background: var(--journal-surface-subtle);
}

.journal-note {
  border-radius: 16px;
  border: 1px solid var(--journal-soft-border);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 96%, transparent), color-mix(in srgb, var(--journal-surface-subtle) 94%, transparent));
  padding: 0.875rem 1rem;
}

.journal-note-label {
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-eyebrow-soft {
  color: var(--journal-muted);
  border-color: var(--journal-soft-border);
  background: color-mix(in srgb, var(--journal-track) 82%, transparent);
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.category-board {
  border-top: 1px solid var(--journal-divider);
}

.category-board--embedded {
  margin-top: 1.25rem;
}

.category-section + .category-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--journal-divider);
}

.category-highlight,
.category-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: 1rem 1.1rem;
}

.category-head {
  display: grid;
  grid-template-columns: var(--category-cols);
  gap: 0.5rem;
  padding-bottom: 0.75rem;
  font-size: 0.69rem;
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
  margin-top: 1rem;
  padding-top: 1rem;
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
  gap: 0.5rem;
  align-items: center;
}

.category-row__name {
  font-size: 0.86rem;
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
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.category-row__count {
  font-size: 0.78rem;
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

.journal-btn-primary,
.journal-btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  min-height: 34px;
  border-radius: 10px;
  padding: 0.45rem 1rem;
  font-size: 0.82rem;
  font-weight: 600;
  transition: all 0.15s;
}

.journal-btn-primary {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-btn-primary:hover {
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.journal-btn-outline {
  border: 1px solid var(--journal-control-border);
  background: var(--journal-surface);
  color: var(--journal-muted);
}

.journal-btn-outline:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 52%, var(--journal-control-border));
  color: var(--journal-accent-strong);
}

.journal-btn-primary:focus-visible,
.journal-btn-outline:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--journal-accent) 58%, white);
  outline-offset: 2px;
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: color-mix(in srgb, var(--color-text-primary) 88%, var(--color-text-secondary));
  --journal-muted: var(--color-text-secondary);
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--color-border-default) 70%, transparent);
  --journal-control-border: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --journal-divider: color-mix(in srgb, var(--color-border-default) 64%, transparent);
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 16%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
}

:global([data-theme='dark']) .journal-note,
:global([data-theme='dark']) .category-highlight,
:global([data-theme='dark']) .category-list,
:global([data-theme='dark']) .journal-btn-outline {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

@media (max-width: 767px) {
  .journal-shell {
    --category-cols: minmax(100px, 1fr) minmax(68px, 84px) minmax(92px, 116px);
  }

  .journal-btn-primary,
  .journal-btn-outline {
    min-height: 36px;
  }
}
</style>
