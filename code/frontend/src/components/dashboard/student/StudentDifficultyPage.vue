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
  <section class="journal-shell space-y-6 journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Difficulty Ladder</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            难度层级总览
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            看不同难度的完成情况和下一步训练重点。
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
                当前覆盖率最低：{{ nextFocus.rate }}%（{{ nextFocus.solved }}/{{
                  nextFocus.total
                }}），适合优先突破。
              </div>
            </div>
            <div v-else class="text-sm text-[var(--journal-muted)]">
              暂无数据，完成更多题目后会给出建议。
            </div>
          </div>
        </article>
      </div>

      <div class="difficulty-board mt-6 px-1 pt-5 md:px-2 md:pt-6">
        <section class="difficulty-section">
          <div class="journal-eyebrow journal-eyebrow-soft">Difficulty Layer View</div>
          <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">难度层级视图</h3>

          <div
            v-if="orderedStats.length === 0"
            class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-soft-border)] bg-[var(--journal-surface-subtle)]/52 px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
          >
            暂无难度统计数据。
          </div>

          <div v-else class="difficulty-list mt-5">
            <div
              v-for="item in orderedStats"
              :key="item.difficulty"
              class="difficulty-item"
            >
              <div class="flex items-center justify-between gap-4">
                <div class="min-w-0">
                  <div class="text-sm font-semibold text-[var(--journal-ink)]">
                    {{ difficultyLabel(item.difficulty) }}
                  </div>
                  <div class="mt-1 text-xs text-[var(--journal-muted)]">
                    {{ item.solved }} / {{ item.total }} 题
                  </div>
                </div>
                <div class="text-right">
                  <div class="text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
                    {{ item.rate }}%
                  </div>
                </div>
              </div>
              <div class="mt-4 h-2 rounded-full bg-[var(--journal-track)]">
                <div
                  class="h-2 rounded-full transition-all"
                  :style="{ width: `${item.rate}%`, background: barColorMap[item.difficulty] }"
                />
              </div>
            </div>
          </div>
        </section>

        <section class="difficulty-section">
          <div class="grid gap-6 xl:grid-cols-[minmax(0,1.02fr)_minmax(320px,0.98fr)]">
            <div>
              <div class="journal-eyebrow journal-eyebrow-soft">Difficulty Interpretation</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">训练解读</h3>

              <div class="difficulty-insight-list mt-5">
                <article class="difficulty-insight-item">
                  <div class="flex items-start gap-3">
                    <div class="stat-icon stat-icon--success">
                      <ShieldCheck class="h-5 w-5" />
                    </div>
                    <div>
                      <div class="text-sm font-semibold text-[var(--journal-ink)]">
                        是否停留在舒适区？
                      </div>
                      <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                        若入门和简单覆盖率高，而中等和困难偏低，说明训练长期停留在舒适区，适合主动抬高强度。
                      </p>
                    </div>
                  </div>
                </article>

                <article class="difficulty-insight-item">
                  <div class="flex items-start gap-3">
                    <div class="stat-icon stat-icon--primary">
                      <Layers2 class="h-5 w-5" />
                    </div>
                    <div>
                      <div class="text-sm font-semibold text-[var(--journal-ink)]">难度结构分布</div>
                      <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                        健康的训练结构通常是低难度覆盖更高，再逐级递减。如果分布反常，说明路径需要调整。
                      </p>
                    </div>
                  </div>
                </article>
              </div>
            </div>

            <aside class="difficulty-focus">
              <div class="journal-eyebrow journal-eyebrow-soft">Focus Card</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">当前突破口</h3>
              <div class="mt-5 rounded-[22px] border border-[var(--journal-soft-border)] bg-[var(--journal-surface)]/70 p-4">
                <div v-if="nextFocus" class="space-y-3">
                  <div class="journal-note-label">优先难度</div>
                  <div class="text-lg font-semibold text-[var(--journal-ink)]">
                    {{ difficultyLabel(nextFocus.difficulty) }}
                  </div>
                  <div class="text-sm leading-6 text-[var(--journal-muted)]">
                    先把这个层级补到更稳定，再逐步推高整体训练强度。
                  </div>
                </div>
                <div v-else class="text-sm text-[var(--journal-muted)]">等待训练数据。</div>
              </div>
            </aside>
          </div>
        </section>
      </div>
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--journal-border) 68%, transparent);
  --journal-divider: color-mix(in srgb, var(--journal-border) 56%, transparent);
  --journal-track: color-mix(in srgb, var(--color-bg-elevated) 58%, var(--journal-surface-subtle));
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
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
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base)));
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
  background: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 34%, transparent);
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.difficulty-board {
  border-top: 1px dashed var(--journal-divider);
}

.difficulty-section + .difficulty-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px dashed var(--journal-divider);
}

.difficulty-list,
.difficulty-insight-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.difficulty-item,
.difficulty-insight-item {
  padding: 1rem 1.1rem;
}

.difficulty-item + .difficulty-item,
.difficulty-insight-item + .difficulty-insight-item {
  border-top: 1px dashed var(--journal-divider);
}

.difficulty-focus {
  position: relative;
}

.stat-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid var(--journal-soft-border);
  background: var(--journal-surface-subtle);
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

@media (min-width: 1280px) {
  .difficulty-focus {
    padding-left: 1.5rem;
  }

  .difficulty-focus::before {
    content: '';
    position: absolute;
    left: -0.75rem;
    top: 0;
    bottom: 0;
    border-left: 1px dashed var(--journal-divider);
  }
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(99, 102, 241, 0.18), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .journal-note,
:global([data-theme='dark']) .difficulty-list,
:global([data-theme='dark']) .difficulty-insight-list,
:global([data-theme='dark']) .difficulty-focus .rounded-\[22px\] {
  background: rgba(15, 23, 42, 0.42);
}
</style>
