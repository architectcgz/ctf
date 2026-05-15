<script setup lang="ts">
import { computed } from 'vue'
import { BarChart2, Flame, Target } from 'lucide-vue-next'

import {
  ChallengeDifficultyText,
  getChallengeDifficultyColor,
  toChallengeDifficulty,
} from '@/entities/challenge'
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
  beginner: 'var(--challenge-difficulty-pill-beginner)',
  easy: 'var(--challenge-difficulty-pill-easy)',
  medium: 'var(--challenge-difficulty-pill-medium)',
  hard: 'var(--challenge-difficulty-pill-hard)',
  insane: 'var(--challenge-difficulty-pill-insane)',
}

const orderedStats = computed<RankedDifficultyStat[]>(() =>
  orderDifficultyActionItems(props.difficultyStats)
)
const primaryDifficulty = computed<RankedDifficultyStat | null>(() =>
  selectDifficultyPriority(props.difficultyStats)
)
const hasDifficultyStats = computed(() => orderedStats.value.length > 0)
const hasPendingDifficulty = computed(() =>
  orderedStats.value.some((item) => item.remaining > 0)
)
const solvedDifficultyCount = computed(
  () => orderedStats.value.filter((item) => item.solved > 0).length
)
const headlineTitle = computed(() =>
  !hasDifficultyStats.value
    ? '先开始建立强度节奏'
    : !hasPendingDifficulty.value
      ? '强度层级已补齐'
      : `先推这一档强度：${difficultyLabel(primaryDifficulty.value!.difficulty)}`
)
const headlineCopy = computed(() =>
  !hasDifficultyStats.value
    ? '先积累几道题的样本，这里会开始告诉你下一步该推哪一档。'
    : !hasPendingDifficulty.value
      ? '所有难度档位都已补齐，接下来可以从题库继续挑题维持手感。'
      : '先补当前最该推进的一档，再决定要不要继续抬强度。'
)
const summaryCards = computed(() => [
  {
    key: 'focus',
    label: '当前完成率',
    value: primaryDifficulty.value
      ? `${primaryDifficulty.value.rate}%`
      : hasDifficultyStats.value
        ? '100%'
        : '待建立',
    icon: Target,
    helper: primaryDifficulty.value
      ? `${difficultyLabel(primaryDifficulty.value.difficulty)} 还有 ${primaryDifficulty.value.remaining} 道题待补，先把这一档推稳。`
      : hasDifficultyStats.value
        ? '所有难度档位都已补齐，可以继续按兴趣挑题维持手感。'
        : '先做出第一批难度分布，这里就会告诉你下一步该推哪一档。',
  },
  {
    key: 'coverage',
    label: '当前覆盖层级',
    value: hasDifficultyStats.value
      ? `${solvedDifficultyCount.value} / ${orderedStats.value.length} 档`
      : '待建立',
    icon: BarChart2,
    helper: hasDifficultyStats.value
      ? hasPendingDifficulty.value
        ? '看已经摸到的难度层级有多少，判断训练是不是还停在熟悉区间。'
        : '所有难度层级都已经跑通，可以继续把训练节奏保持住。'
      : '还没有难度数据时，先从题库里做几道题把强度分布跑出来。',
  },
  {
    key: 'pace',
    label: '推进节奏',
    value: primaryDifficulty.value
      ? `先补 ${difficultyLabel(primaryDifficulty.value.difficulty)}`
      : hasDifficultyStats.value
        ? '保持节奏'
        : '先形成样本',
    icon: Flame,
    helper: primaryDifficulty.value
      ? '先补完成率更低的一档，再继续往上提强度，训练节奏会更稳。'
      : hasDifficultyStats.value
        ? '所有档位都已补齐后，就按题库兴趣继续推进即可。'
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

function difficultyPillValue(difficulty: string) {
  return toChallengeDifficulty(difficulty)
}

function difficultyActionItemStyle(difficulty: string, isPrimary: boolean): Record<string, string> {
  const normalizedDifficulty = toChallengeDifficulty(difficulty)
  if (!normalizedDifficulty) {
    return {}
  }

  const color = getChallengeDifficultyColor(normalizedDifficulty)

  return {
    '--journal-soft-panel-item-border': `color-mix(in srgb, ${color} ${isPrimary ? 40 : 30}%, var(--journal-shell-border))`,
    '--journal-soft-panel-item-background': `color-mix(in srgb, ${color} ${isPrimary ? 10 : 6}%, var(--journal-surface-subtle))`,
  }
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
      <div class="workspace-panel-header difficulty-header">
        <div class="workspace-panel-header__intro">
          <div class="workspace-overline">Difficulty</div>
          <h1 class="journal-page-title workspace-page-title journal-soft-page-title">
            {{ headlineTitle }}
          </h1>
          <p class="workspace-page-copy max-w-2xl">
            {{ headlineCopy }}
          </p>
        </div>

        <div
          class="workspace-panel-header__actions flex flex-wrap gap-3"
          role="group"
          aria-label="难度进度快捷操作"
        >
          <button class="journal-btn-primary" @click="openPrimaryDifficulty">
            {{
              primaryDifficulty
                ? `先做${difficultyLabel(primaryDifficulty.difficulty)}`
                : hasDifficultyStats
                  ? '继续训练'
                  : '去训练'
            }}
          </button>
          <button class="journal-btn-outline" @click="emit('openChallenges')">浏览全部题目</button>
        </div>

        <div
          class="workspace-panel-header__summary difficulty-summary-strip progress-strip metric-panel-grid metric-panel-default-surface"
        >
          <article
            v-for="card in summaryCards"
            :key="card.key"
            class="difficulty-summary-card progress-card metric-panel-card"
          >
            <div class="journal-note-label progress-card-label metric-panel-label">
              <span>{{ card.label }}</span>
              <component :is="card.icon" class="h-4 w-4" />
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
      <div class="workspace-panel-divider" aria-hidden="true" />

      <div class="difficulty-board px-1 pt-0 md:px-2 md:pt-0">
        <section class="difficulty-section">
          <div v-if="hasDifficultyStats" class="difficulty-toolbar">
            <p class="difficulty-toolbar__copy">
              {{
                hasPendingDifficulty
                  ? '难度顺序固定，主推档位已高亮，按列表从上往下看就够了。'
                  : '所有难度档位都已补齐，按列表查看当前训练分布就行。'
              }}
            </p>
          </div>

          <div v-if="!hasDifficultyStats" class="journal-soft-empty-state mt-5">
            当前还没有难度统计数据，先完成几道题再回来查看。
          </div>

          <div v-else class="difficulty-action-list journal-soft-panel-shell mt-5">
            <article
              v-for="(item, index) in orderedStats"
              :key="item.difficulty"
              class="difficulty-action-item journal-soft-panel-item"
              :style="
                difficultyActionItemStyle(
                  item.difficulty,
                  primaryDifficulty?.difficulty === item.difficulty
                )
              "
              :class="{
                'difficulty-action-item--primary':
                  primaryDifficulty?.difficulty === item.difficulty,
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
                    <ChallengeDifficultyText
                      v-if="difficultyPillValue(item.difficulty)"
                      :difficulty="difficultyPillValue(item.difficulty)!"
                    />
                    <span v-else class="difficulty-action-item__name">{{
                      difficultyLabel(item.difficulty)
                    }}</span>
                    <span class="difficulty-action-item__rate">{{ item.rate }}%</span>
                    <span class="difficulty-action-item__count"
                      >{{ item.solved }}/{{ item.total }}</span
                    >
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
                  class="difficulty-action-item__cta"
                  :class="{
                    'journal-btn-primary': primaryDifficulty?.difficulty === item.difficulty,
                    'journal-btn-secondary': primaryDifficulty?.difficulty !== item.difficulty,
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
          <div class="difficulty-note journal-soft-panel-shell">
            <Flame class="difficulty-note__icon h-4 w-4" />
            <p class="difficulty-note__copy">
              {{
                primaryDifficulty
                  ? `当前最需要补的是${difficultyLabel(primaryDifficulty.difficulty)}。先补当前断档，再按层级继续往上推。`
                  : hasDifficultyStats
                    ? '所有难度档位都已补齐，接下来可以从题库继续挑题维持手感。'
                    : '先积累一批真实训练样本，这一页才会开始根据你的强度分布安排下一步动作。'
              }}
            </p>
          </div>
        </section>
      </div>
    </div>
  </section>
</template>

<style scoped>
.difficulty-summary-strip {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.difficulty-summary-strip.metric-panel-default-surface {
  --metric-panel-border: var(--journal-soft-border);
  --metric-panel-shadow: 0 10px 20px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent);
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
  --journal-soft-panel-shell-padding: var(--space-4) var(--space-4-5);
}

.difficulty-action-item + .difficulty-action-item {
  margin-top: var(--space-3);
}

.difficulty-action-item {
  --journal-soft-panel-item-padding: var(--space-2) var(--space-2-5);
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
