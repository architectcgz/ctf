<script setup lang="ts">
import { computed } from 'vue'
import { Flame } from 'lucide-vue-next'

import { difficultyLabel } from '@/utils/challenge'

import { orderDifficultyActionItems, selectDifficultyPriority } from './utils'

interface DifficultyStat {
  difficulty: string
  total: number
  solved: number
}

interface RankedDifficultyStat extends DifficultyStat {
  rate: number
  remaining: number
  order: number
}

const props = withDefaults(
  defineProps<{
    difficultyStats: DifficultyStat[]
    embedded?: boolean
  }>(),
  {
    embedded: false,
  }
)

const emit = defineEmits<{
  openChallenges: []
  openDifficultyChallenges: [difficulty: string]
}>()

const barColorMap: Record<string, string> = {
  beginner: '#10b981',
  easy: '#22d3ee',
  medium: '#f59e0b',
  hard: '#f97316',
  insane: '#ef4444',
}

const orderedStats = computed<RankedDifficultyStat[]>(() =>
  orderDifficultyActionItems(props.difficultyStats)
)
const primaryDifficulty = computed<RankedDifficultyStat | null>(() =>
  selectDifficultyPriority(props.difficultyStats)
)
const hasDifficultyStats = computed(() => orderedStats.value.length > 0)
const solvedDifficultyCount = computed(() => orderedStats.value.filter((item) => item.solved > 0).length)
const headlineTitle = computed(() =>
  primaryDifficulty.value
    ? `先推这一档强度：${difficultyLabel(primaryDifficulty.value.difficulty)}`
    : '先开始建立强度节奏'
)
const summaryCards = computed(() => [
  {
    key: 'focus',
    label: '当前完成率',
    value: primaryDifficulty.value ? `${primaryDifficulty.value.rate}%` : '待建立',
    helper: primaryDifficulty.value
      ? `${difficultyLabel(primaryDifficulty.value.difficulty)} 还有 ${primaryDifficulty.value.remaining} 道题待补，先把这一档推稳。`
      : '先做出第一批难度分布，这里就会告诉你下一步该推哪一档。',
  },
  {
    key: 'coverage',
    label: '当前覆盖层级',
    value: hasDifficultyStats.value ? `${solvedDifficultyCount.value} / ${orderedStats.value.length} 档` : '待建立',
    helper: hasDifficultyStats.value
      ? '看已经摸到的难度层级有多少，判断训练是不是还停在熟悉区间。'
      : '还没有难度数据时，先从题库里做几道题把强度分布跑出来。',
  },
  {
    key: 'pace',
    label: '推进节奏',
    value: primaryDifficulty.value ? `先补 ${difficultyLabel(primaryDifficulty.value.difficulty)}` : '先形成样本',
    helper: primaryDifficulty.value
      ? '先补完成率更低的一档，再继续往上提强度，训练节奏会更稳。'
      : '先把训练样本积累起来，再决定应该从哪一档开始往上推。',
  },
])

function difficultyActionCopy(item: RankedDifficultyStat): string {
  if (item.remaining <= 0) {
    return '这一档已经补齐，可以在后面用作热身或维持当前手感。'
  }

  if (primaryDifficulty.value?.difficulty === item.difficulty) {
    return `还有 ${item.remaining} 道题待补，先把这一档推稳，再继续往更高强度走。`
  }

  if (primaryDifficulty.value && item.order < primaryDifficulty.value.order) {
    return `这一档已经在前面打过底，必要时可以回来热身，但当前不需要继续堆太多。`
  }

  return `先把前一档补稳，再顺着这一档继续抬升强度，避免直接跳档造成训练断层。`
}

function openPrimaryDifficulty(): void {
  if (primaryDifficulty.value) {
    emit('openDifficultyChallenges', primaryDifficulty.value.difficulty)
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
        : 'journal-shell journal-hero rounded-[30px] border px-6 py-6 md:px-8'
    "
  >
    <div class="difficulty-header">
      <div class="journal-eyebrow">Intensity Workspace</div>
      <h1 class="journal-page-title workspace-tab-heading__title text-[var(--journal-ink)]">
        {{ headlineTitle }}
      </h1>
      <p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
        {{
          hasDifficultyStats
            ? '先把当前最值得推进的一档补稳，再决定下一步要不要继续抬高训练强度。'
            : '先完成几道题，把第一批强度分布跑出来，这里就会开始按节奏给出下一步动作。'
        }}
      </p>

      <div class="mt-5 flex flex-wrap gap-3" role="group" aria-label="难度进度快捷操作">
        <button class="journal-btn-primary" @click="openPrimaryDifficulty">
          {{ primaryDifficulty ? `先做${difficultyLabel(primaryDifficulty.difficulty)}` : '去训练' }}
        </button>
        <button class="journal-btn-outline" @click="emit('openChallenges')">浏览全部题目</button>
      </div>

      <div class="difficulty-summary-strip mt-5 progress-strip metric-panel-grid metric-panel-default-surface">
        <article
          v-for="card in summaryCards"
          :key="card.key"
          class="difficulty-summary-card progress-card metric-panel-card"
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
      class="difficulty-board mt-6 px-1 pt-5 md:px-2 md:pt-6"
      :class="{ 'difficulty-board--embedded': embedded }"
    >
      <section class="difficulty-section">
        <div v-if="hasDifficultyStats" class="difficulty-toolbar">
          <p class="difficulty-toolbar__copy">
            难度顺序固定，主推档位已高亮，按列表从上往下看就够了。
          </p>
        </div>

        <div
          v-if="!hasDifficultyStats"
          class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-shell-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
        >
          当前还没有难度统计数据，先完成几道题再回来查看。
        </div>

        <div v-else class="difficulty-action-list mt-5">
          <article
            v-for="(item, index) in orderedStats"
            :key="item.difficulty"
            class="difficulty-action-item"
            :class="{
              'difficulty-action-item--primary': primaryDifficulty?.difficulty === item.difficulty,
            }"
            :data-test="`difficulty-action-${item.difficulty}`"
            :aria-current="primaryDifficulty?.difficulty === item.difficulty ? 'step' : undefined"
          >
            <div class="difficulty-action-item__body">
              <div class="difficulty-action-rank">
                {{ String(index + 1).padStart(2, '0') }}
              </div>
              <div class="difficulty-action-item__content">
                <div class="difficulty-action-item__meta">
                  <span class="difficulty-action-item__name">{{ difficultyLabel(item.difficulty) }}</span>
                  <span class="difficulty-action-item__rate">{{ item.rate }}%</span>
                  <span class="difficulty-action-item__count">{{ item.solved }}/{{ item.total }}</span>
                </div>
                <p class="difficulty-action-item__copy">
                  {{ difficultyActionCopy(item) }}
                </p>
                <div class="difficulty-track">
                  <div
                    class="difficulty-track-fill h-2 rounded-full"
                    :style="{ width: `${item.rate}%`, background: barColorMap[item.difficulty] }"
                  />
                </div>
              </div>
              <button
                type="button"
                class="journal-btn-primary difficulty-action-item__cta"
                :class="{
                  'difficulty-action-item__cta--secondary':
                    primaryDifficulty?.difficulty !== item.difficulty,
                }"
                @click="emit('openDifficultyChallenges', item.difficulty)"
              >
                去做这一档
              </button>
            </div>
          </article>
        </div>
      </section>

      <section v-if="hasDifficultyStats" class="difficulty-section difficulty-section--compact">
        <div class="difficulty-note">
          <Flame class="difficulty-note__icon h-4 w-4" />
          <p class="difficulty-note__copy">
            {{
              primaryDifficulty
                ? `当前最需要补的是${difficultyLabel(primaryDifficulty.difficulty)}。先补当前断档，再按层级继续往上推。`
                : '先积累一批真实训练样本，这一页才会开始根据你的强度分布安排下一步动作。'
            }}
          </p>
        </div>
      </section>
    </div>
  </section>
</template>

<style scoped>
.difficulty-header {
  display: grid;
  gap: var(--space-3);
}

.difficulty-summary-strip {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.difficulty-summary-strip.metric-panel-default-surface {
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

.difficulty-board {
  border-top: 1px solid var(--journal-divider);
}

.difficulty-board--embedded {
  margin-top: var(--space-5);
}

.difficulty-toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.difficulty-toolbar__copy {
  margin: 0;
  font-size: var(--font-size-0-82);
  line-height: 1.7;
  color: var(--journal-muted);
}

.difficulty-section + .difficulty-section {
  margin-top: var(--space-6);
  padding-top: var(--space-6);
  border-top: 1px solid var(--journal-divider);
}

.difficulty-action-list,
.difficulty-note {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-4) var(--space-4-5);
}

.difficulty-action-item + .difficulty-action-item {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--journal-divider);
}

.difficulty-action-item {
  border-radius: 18px;
  padding: var(--space-2) var(--space-2-5);
}

.difficulty-action-item--primary {
  background: color-mix(in srgb, var(--journal-accent) 7%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.difficulty-action-item__body {
  display: grid;
  grid-template-columns: 2.75rem minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-4);
}

.difficulty-action-rank {
  display: flex;
  min-width: 2.75rem;
  height: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.95rem;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, transparent);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
  font-weight: 700;
  letter-spacing: 0.08em;
}

.difficulty-action-item--primary .difficulty-action-rank {
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
  color: var(--journal-accent-strong);
}

.difficulty-action-item__content {
  min-width: 0;
}

.difficulty-action-item__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem;
}

.difficulty-action-item__name {
  font-size: var(--font-size-0-82);
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--journal-ink);
}

.difficulty-action-item__copy {
  margin-top: 0.7rem;
  font-size: var(--font-size-0-82);
  line-height: 1.7;
  color: var(--journal-muted);
}

.difficulty-action-item__rate,
.difficulty-action-item__count {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.25rem 0.6rem;
  font-size: var(--font-size-0-74);
  font-weight: 600;
}

.difficulty-action-item__rate {
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-accent-strong);
}

.difficulty-action-item__count {
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, transparent);
  color: var(--journal-muted);
}

.difficulty-action-item__cta {
  flex-shrink: 0;
  min-width: 6rem;
}

.difficulty-action-item__cta--secondary {
  border-color: var(--journal-control-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  color: var(--journal-ink);
}

.difficulty-note {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
}

.difficulty-note__icon {
  margin-top: 0.125rem;
  flex-shrink: 0;
  color: var(--journal-accent);
}

.difficulty-note__copy {
  margin: 0;
  font-size: var(--font-size-0-82);
  line-height: 1.8;
  color: var(--journal-muted);
}

.difficulty-track {
  margin-top: 0.8rem;
  height: 0.5rem;
  overflow: hidden;
  border-radius: 999px;
  background: var(--journal-track);
}

:global([data-theme='dark']) .difficulty-action-list,
:global([data-theme='dark']) .difficulty-note {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

@media (max-width: 900px) {
  .difficulty-summary-strip {
    --metric-panel-columns: 1fr;
  }

  .difficulty-action-item__body {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }

  .difficulty-action-rank {
    width: 2.5rem;
  }

  .difficulty-action-item__cta {
    width: 100%;
    justify-content: center;
  }
}
</style>
