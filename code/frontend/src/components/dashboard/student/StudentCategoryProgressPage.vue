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
    <!-- Hero -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Coverage Overview</div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            分类覆盖概况
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            从整体覆盖率判断训练结构，再对照分类进度决定下一步优先补强的方向。
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
    </section>

    <!-- 强弱方向 -->
    <section class="grid gap-4 md:grid-cols-2">
      <article class="journal-panel rounded-[24px] border px-6 py-6">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="journal-eyebrow">Strongest Direction</div>
            <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">{{ strongestCategory?.category || '-' }}</h3>
          </div>
          <div class="direction-icon direction-icon--success">
            <ArrowUpRight class="h-4 w-4" />
          </div>
        </div>
        <div v-if="strongestCategory" class="mt-4">
          <div class="flex items-end justify-between text-sm">
            <span class="text-[var(--journal-muted)]">完成进度</span>
            <span class="font-semibold text-[var(--journal-ink)]">{{ strongestCategory.solved }} / {{ strongestCategory.total }}</span>
          </div>
          <div class="category-track mt-2 h-2.5 rounded-full">
            <div class="h-2.5 rounded-full bg-emerald-500" :style="{ width: `${strongestCategory.rate}%` }" />
          </div>
          <div class="mt-2 text-right text-xs font-semibold text-emerald-600">{{ strongestCategory.rate }}%</div>
        </div>
      </article>

      <article class="journal-panel rounded-[24px] border px-6 py-6">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="journal-eyebrow">Weakest Direction</div>
            <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">{{ weakestCategory?.category || '-' }}</h3>
          </div>
          <div class="direction-icon direction-icon--warning">
            <Gauge class="h-4 w-4" />
          </div>
        </div>
        <div v-if="weakestCategory" class="mt-4">
          <div class="flex items-end justify-between text-sm">
            <span class="text-[var(--journal-muted)]">完成进度</span>
            <span class="font-semibold text-[var(--journal-ink)]">{{ weakestCategory.solved }} / {{ weakestCategory.total }}</span>
          </div>
          <div class="category-track mt-2 h-2.5 rounded-full">
            <div class="h-2.5 rounded-full bg-amber-400" :style="{ width: `${weakestCategory.rate}%` }" />
          </div>
          <div class="mt-2 text-right text-xs font-semibold text-amber-600">{{ weakestCategory.rate }}%</div>
        </div>
      </article>
    </section>

    <!-- 分类进度板 -->
    <section class="journal-panel rounded-[24px] border px-6 py-6">
      <div class="flex items-start justify-between gap-4">
        <div>
          <div class="journal-eyebrow">Category Board</div>
          <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">分类进度板</h3>
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

      <div v-else class="mt-5 space-y-4">
        <div
          v-for="item in rankedCategories"
          :key="item.category"
          class="journal-log rounded-[18px] border px-5 py-4"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm font-semibold uppercase tracking-[0.18em] text-[var(--journal-ink)]">
              {{ item.category }}
            </div>
            <div class="text-right">
              <span class="text-sm font-semibold text-[var(--journal-ink)]">{{ item.rate }}%</span>
              <span class="ml-2 text-xs text-[var(--journal-muted)]">{{ item.solved }}/{{ item.total }}</span>
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
  font-family: "Inter", "Noto Sans SC", system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(191, 219, 254, 0.75), transparent 15rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-panel,
.journal-log {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-log {
  transition: all 0.2s ease-in-out;
}

.journal-log:hover {
  border-color: #6366f1;
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.06);
}

.journal-note {
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: var(--journal-surface-subtle);
  padding: 0.95rem 1rem;
}

.journal-note-label,
.journal-eyebrow {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.26em;
  text-transform: uppercase;
  color: #64748b;
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.direction-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid rgba(226, 232, 240, 0.72);
  background: #f8fafc;
}

.direction-icon--success {
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.2);
  background: rgba(16, 185, 129, 0.08);
}

.direction-icon--warning {
  color: #f59e0b;
  border-color: rgba(245, 158, 11, 0.2);
  background: rgba(245, 158, 11, 0.08);
}

.category-track {
  background: rgba(226, 232, 240, 0.6);
}

.journal-btn-primary {
  border-radius: 10px;
  background: var(--journal-accent);
  padding: 0.45rem 1.1rem;
  font-size: 0.82rem;
  font-weight: 600;
  color: #fff;
  transition: background 0.15s;
}

.journal-btn-primary:hover {
  background: var(--journal-accent-strong);
}

.journal-btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border: 1px solid var(--journal-border);
  border-radius: 10px;
  background: var(--journal-surface);
  padding: 0.4rem 1rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--journal-muted);
  transition: all 0.15s;
}

.journal-btn-outline:hover {
  border-color: #6366f1;
  color: var(--journal-accent-strong);
}
</style>
