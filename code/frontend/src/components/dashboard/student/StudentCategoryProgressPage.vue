<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpRight, Gauge, MoveRight } from 'lucide-vue-next'

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
  <div class="journal-shell space-y-6">
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Coverage Overview</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            分类覆盖概况
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            看各分类的完成情况和当前强弱项。
          </p>
          <div class="mt-6 flex flex-wrap gap-3">
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

      <div class="category-board mt-6 px-1 pt-5 md:px-2 md:pt-6">
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
            class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
          >
            当前还没有分类统计数据，先完成几道题再回来查看。
          </div>

          <div v-else class="category-list mt-5">
            <div
              v-for="item in rankedCategories"
              :key="item.category"
              class="category-item"
            >
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div
                  class="text-sm font-semibold uppercase tracking-[0.18em] text-[var(--journal-ink)]"
                >
                  {{ item.category }}
                </div>
                <div class="text-right">
                  <span class="text-sm font-semibold text-[var(--journal-ink)]">{{ item.rate }}%</span>
                  <span class="ml-2 text-xs text-[var(--journal-muted)]"
                    >{{ item.solved }}/{{ item.total }}</span
                  >
                </div>
              </div>
              <div class="category-track mt-3 h-2 rounded-full">
                <div
                  class="h-2 rounded-full bg-[linear-gradient(90deg,rgba(34,211,238,0.95),rgba(56,189,248,0.72))]"
                  :style="{ width: `${item.rate}%` }"
                />
              </div>
            </div>
          </div>
        </section>
      </div>
    </section>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-ink: #0f172a;
  --journal-muted: #475569;
  --journal-border: rgba(226, 232, 240, 0.72);
  --journal-surface: #ffffff;
  --journal-surface-subtle: #f8fafc;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.98), rgba(241, 245, 249, 0.95));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.05);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
}

.journal-note {
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.12);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.78), rgba(248, 250, 252, 0.92));
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
  border: 1px solid rgba(99, 102, 241, 0.22);
  background: rgba(99, 102, 241, 0.07);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-eyebrow-soft {
  color: var(--journal-muted);
  border-color: rgba(148, 163, 184, 0.28);
  background: rgba(148, 163, 184, 0.08);
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.category-board {
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.category-section + .category-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.category-highlight,
.category-list {
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(255, 255, 255, 0.42);
  padding: 1rem 1.1rem;
}

.category-item + .category-item {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.category-track {
  background: rgba(226, 232, 240, 0.65);
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
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  color: var(--journal-muted);
}

.journal-btn-outline:hover {
  border-color: #6366f1;
  color: var(--journal-accent-strong);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f8fafc;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.82);
  --journal-surface-subtle: rgba(30, 41, 59, 0.72);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(99, 102, 241, 0.18), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .journal-note,
:global([data-theme='dark']) .category-highlight,
:global([data-theme='dark']) .category-list,
:global([data-theme='dark']) .journal-btn-outline {
  background: rgba(15, 23, 42, 0.42);
}
</style>
