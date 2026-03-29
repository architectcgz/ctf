<script setup lang="ts">
import { computed } from 'vue'
import { Flame, Layers2, ShieldCheck } from 'lucide-vue-next'

import { difficultyLabel } from '@/utils/challenge'

import { progressRate } from './utils'

interface DifficultyStat {
  difficulty: string
  total: number
  solved: number
}

const props = defineProps<{
  difficultyStats: DifficultyStat[]
}>()

const difficultyOrder = ['beginner', 'easy', 'medium', 'hard', 'insane']
const barColorMap: Record<string, string> = {
  beginner: '#10b981',
  easy: '#22d3ee',
  medium: '#f59e0b',
  hard: '#f97316',
  insane: '#ef4444',
}

const orderedStats = computed(() =>
  difficultyOrder
    .map((difficulty) => props.difficultyStats.find((item) => item.difficulty === difficulty))
    .filter((item): item is DifficultyStat => Boolean(item))
    .map((item) => ({
      ...item,
      rate: progressRate(item.total, item.solved),
    }))
)

const nextFocus = computed(
  () =>
    [...orderedStats.value]
      .filter((item) => item.total > 0)
      .sort((left, right) => left.rate - right.rate)[0] || null
)
</script>

<template>
  <div class="journal-shell space-y-6">
    <!-- Hero -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Difficulty Ladder</div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            难度层级总览
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            观察不同难度的覆盖情况，判断训练是否长期停留在舒适区，并给出下一阶段更适合推进的台阶。
          </p>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <Flame class="h-5 w-5 text-[var(--journal-accent)]" />
            下一步重点
          </div>
          <div class="mt-5">
            <div v-if="nextFocus">
              <div class="journal-note">
                <div class="journal-note-label">建议主攻</div>
                <div class="journal-note-value">{{ difficultyLabel(nextFocus.difficulty) }}</div>
              </div>
              <div class="mt-3 text-sm leading-6 text-[var(--journal-muted)]">
                当前覆盖率最低：{{ nextFocus.rate }}%（{{ nextFocus.solved }}/{{ nextFocus.total }}），优先突破此难度层。
              </div>
            </div>
            <div v-else class="text-sm text-[var(--journal-muted)]">
              暂无数据，完成更多题目后将给出建议。
            </div>
          </div>
        </article>
      </div>
    </section>

    <!-- 难度卡片 bento -->
    <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
      <article
        v-for="item in orderedStats"
        :key="item.difficulty"
        class="journal-metric rounded-[24px] border px-5 py-5"
      >
        <div class="journal-eyebrow">{{ difficultyLabel(item.difficulty) }}</div>
        <div class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">{{ item.rate }}%</div>
        <div class="mt-1 text-sm text-[var(--journal-muted)]">{{ item.solved }} / {{ item.total }}</div>
        <div class="mt-4 h-2 rounded-full" style="background: rgba(226,232,240,0.5)">
          <div
            class="h-2 rounded-full transition-all"
            :style="{ width: `${item.rate}%`, background: barColorMap[item.difficulty] }"
          />
        </div>
      </article>
    </section>

    <!-- 详细视图 + 解读 -->
    <section class="grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
      <!-- 进度条列表 -->
      <article class="journal-panel rounded-[24px] border px-6 py-6">
        <div class="journal-eyebrow">Difficulty Layer View</div>
        <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">难度层级视图</h3>

        <div
          v-if="orderedStats.length === 0"
          class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
        >
          暂无难度统计数据。
        </div>

        <div v-else class="mt-5 space-y-4">
          <div
            v-for="item in orderedStats"
            :key="item.difficulty"
            class="journal-log rounded-[18px] border px-5 py-4"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="text-sm font-semibold text-[var(--journal-ink)]">{{ difficultyLabel(item.difficulty) }}</div>
              <div class="text-sm font-semibold text-[var(--journal-ink)]">{{ item.rate }}%</div>
            </div>
            <div class="mt-1 text-xs text-[var(--journal-muted)]">{{ item.solved }} / {{ item.total }} 题</div>
            <div class="mt-3 h-2 rounded-full" style="background: rgba(226,232,240,0.5)">
              <div
                class="h-2 rounded-full transition-all"
                :style="{ width: `${item.rate}%`, background: barColorMap[item.difficulty] }"
              />
            </div>
          </div>
        </div>
      </article>

      <!-- 解读 -->
      <div class="space-y-4">
        <article class="journal-panel rounded-[24px] border px-6 py-5">
          <div class="flex items-start gap-3">
            <div class="stat-icon stat-icon--success">
              <ShieldCheck class="h-5 w-5" />
            </div>
            <div>
              <div class="journal-eyebrow">Comfort Zone Check</div>
              <h4 class="mt-1 text-base font-semibold text-[var(--journal-ink)]">是否停留在舒适区？</h4>
              <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                若入门/简单覆盖率高而中等/困难覆盖率低，说明训练长期停留在舒适区，建议有意识地向更高难度迈进。
              </p>
            </div>
          </div>
        </article>

        <article class="journal-panel rounded-[24px] border px-6 py-5">
          <div class="flex items-start gap-3">
            <div class="stat-icon stat-icon--primary">
              <Layers2 class="h-5 w-5" />
            </div>
            <div>
              <div class="journal-eyebrow">Difficulty Structure</div>
              <h4 class="mt-1 text-base font-semibold text-[var(--journal-ink)]">难度结构分布</h4>
              <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                健康的训练结构应呈现倒三角：低难度覆盖率最高，逐级递减。若分布反常，说明训练路径需要调整。
              </p>
            </div>
          </div>
        </article>
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
    radial-gradient(circle at top right, rgba(253, 230, 138, 0.5), transparent 15rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-panel,
.journal-metric,
.journal-brief,
.journal-log {
  border-color: var(--journal-border);
}

.journal-panel,
.journal-metric {
  background: var(--journal-surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.journal-brief,
.journal-log {
  background: var(--journal-surface);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
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

.stat-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid rgba(226, 232, 240, 0.72);
  background: #f8fafc;
}

.stat-icon--success {
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.2);
  background: rgba(16, 185, 129, 0.08);
}

.stat-icon--primary {
  color: #4f46e5;
  border-color: rgba(79, 70, 229, 0.2);
  background: rgba(79, 70, 229, 0.08);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f1f5f9;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.85);
  --journal-surface-subtle: rgba(30, 41, 59, 0.6);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(120, 80, 20, 0.18), transparent 15rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}
</style>
