<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, ShieldAlert } from 'lucide-vue-next'

import type { RecommendationItem } from '@/api/contracts'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'

const props = withDefaults(
  defineProps<{
    weakDimensions: string[]
    recommendations: RecommendationItem[]
    embedded?: boolean
  }>(),
  {
    embedded: false,
  }
)

const emit = defineEmits<{
  openChallenge: [challengeId: string]
  openChallenges: []
  openSkillProfile: []
}>()

const visibleWeakDimensions = computed(() => props.weakDimensions.slice(0, 3))
const headline = computed(() => visibleWeakDimensions.value[0] || '保持当前训练节奏')
const targetDifficulty = computed(() =>
  props.recommendations[0] ? difficultyLabel(props.recommendations[0].difficulty) : '待选择'
)
const summaryCards = computed(() => [
  {
    key: 'focus',
    label: '当前补强方向',
    value: headline.value,
    helper:
      visibleWeakDimensions.value.length > 0
        ? `先补 ${visibleWeakDimensions.value.join(' / ')}，把短板重新拉回稳定区间。`
        : '当前结构比较均衡，先保持训练连续性。',
  },
  {
    key: 'difficulty',
    label: '当前目标难度',
    value: targetDifficulty.value,
    helper:
      props.recommendations.length > 0
        ? '先从当前推荐队列开头进入，稳定抬高一档训练强度。'
        : '没有定向题目时，先去题库挑一题恢复训练手感。',
  },
  {
    key: 'queue',
    label: '当前行动队列',
    value: `${props.recommendations.length} 道`,
    helper:
      props.recommendations.length > 0
        ? '先做完这一组，再回来刷新下一批建议。'
        : '当前没有推荐任务，可以先浏览全部题目。',
  },
])
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
    <div class="recommendation-header">
      <div class="journal-eyebrow">Action Queue</div>
      <h1 class="journal-page-title workspace-tab-heading__title text-[var(--journal-ink)]">
        现在先练这几道
      </h1>
      <p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
        第一屏只回答接下来做什么，先把最适合当前阶段的题目按顺序推进。
      </p>

      <div class="mt-5 flex flex-wrap gap-2">
        <template v-if="visibleWeakDimensions.length > 0">
          <span
            v-for="dim in visibleWeakDimensions"
            :key="dim"
            class="inline-flex items-center gap-1.5 rounded-full border border-[var(--journal-accent)]/20 bg-[var(--journal-accent)]/8 px-3 py-1 text-xs font-semibold text-[var(--journal-accent-strong)]"
          >
            <ShieldAlert class="h-3 w-3" />
            {{ dim }}
          </span>
        </template>
        <span v-else class="journal-weak-tag journal-weak-tag--stable"> 暂无明显短板 </span>
      </div>

      <div class="recommendation-summary-strip mt-5 progress-strip metric-panel-grid metric-panel-default-surface">
        <article
          v-for="card in summaryCards"
          :key="card.key"
          class="recommendation-summary-card progress-card metric-panel-card"
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
      class="recommend-board mt-6 px-1 pt-5 md:px-2 md:pt-6"
      :class="{ 'recommend-board--embedded': embedded }"
    >
      <section class="recommend-section">
        <div v-if="recommendations.length > 0" class="recommend-toolbar">
          <p class="recommend-toolbar__copy">
            按当前顺序直接推进，做完这组再回来刷新下一批建议。
          </p>
          <div class="recommend-toolbar__actions">
            <button type="button" class="journal-btn-outline" @click="emit('openSkillProfile')">
              能力画像
            </button>
            <button type="button" class="journal-btn-primary" @click="emit('openChallenges')">
              浏览全部题目
            </button>
          </div>
        </div>

        <div
          v-if="recommendations.length === 0"
          class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-shell-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
        >
          当前没有推荐题目，可以先去题目列表探索新的方向。
          <div class="mt-4">
            <button type="button" class="journal-btn-primary" @click="emit('openChallenges')">
              浏览全部题目
            </button>
          </div>
        </div>

        <div v-else class="recommend-list mt-4">
          <button
            v-for="(item, index) in recommendations"
            :key="item.challenge_id"
            class="recommend-item group w-full cursor-pointer text-left"
            @click="emit('openChallenge', item.challenge_id)"
          >
            <div class="flex items-start gap-4">
              <div
                class="rec-index shrink-0"
                :class="index === 0 ? 'rec-index--top' : 'rec-index--rest'"
              >
                {{ index + 1 }}
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="text-sm font-semibold text-[var(--journal-ink)]">{{
                    item.title
                  }}</span>
                  <span
                    class="rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="difficultyClass(item.difficulty)"
                  >
                    {{ difficultyLabel(item.difficulty) }}
                  </span>
                  <span class="journal-category-chip">
                    {{ item.category }}
                  </span>
                </div>
                <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">{{ item.reason }}</p>
              </div>
              <ArrowRight
                class="mt-1 h-4 w-4 shrink-0 text-[var(--journal-accent-strong)] opacity-0 transition group-hover:opacity-100"
              />
            </div>
          </button>
        </div>
      </section>

    </div>
  </section>
</template>

<style scoped>
.journal-soft-surface {
  --journal-soft-button-height: 34px;
  --journal-soft-button-padding: var(--space-1-5) var(--space-4);
  --journal-soft-button-size: 0.8rem;
  --journal-soft-button-primary-border: color-mix(in srgb, var(--journal-accent) 42%, transparent);
}

.recommendation-header {
  display: grid;
  gap: var(--space-3);
}

.recommendation-summary-strip {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.recommendation-summary-strip.metric-panel-default-surface {
  --metric-panel-border: var(--journal-soft-border);
  --metric-panel-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 14%, transparent),
      transparent 42%
    ),
    linear-gradient(
      165deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base))
    );
  --metric-panel-shadow: 0 10px 20px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent);
}

.recommendation-summary-card {
  min-height: 100%;
}

.recommend-item:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--journal-accent) 58%, white);
  outline-offset: 2px;
}

.recommend-board {
  border-top: 1px solid var(--journal-divider);
}

.recommend-board--embedded {
  margin-top: var(--space-5);
}

.recommend-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.recommend-toolbar__copy {
  margin: 0;
  font-size: var(--font-size-0-82);
  line-height: 1.7;
  color: var(--journal-muted);
}

.recommend-toolbar__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.recommend-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

.recommend-item {
  padding: var(--space-4) var(--space-4-5);
  transition: background 0.2s ease-in-out;
}

.recommend-item + .recommend-item {
  border-top: 1px solid var(--journal-divider);
}

.recommend-item:hover {
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
}

.rec-index {
  display: flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.625rem;
  font-size: var(--font-size-0-875);
  font-weight: 600;
  flex-shrink: 0;
}

.rec-index--top {
  background: var(--journal-accent);
  color: color-mix(in srgb, white 92%, var(--color-bg-base));
}

.rec-index--rest {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-category-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid var(--journal-soft-border);
  background: color-mix(in srgb, var(--journal-track) 82%, transparent);
  padding: var(--space-0-5) var(--space-2);
  font-size: var(--font-size-0-74);
  font-weight: 600;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-weak-tag {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: var(--space-1) var(--space-3);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.journal-weak-tag--stable {
  border-color: color-mix(in srgb, #16a34a 22%, transparent);
  background: color-mix(in srgb, #16a34a 10%, transparent);
  color: #15803d;
}

:global([data-theme='dark']) .recommend-list {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

@media (max-width: 900px) {
  .recommend-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .recommendation-summary-strip {
    --metric-panel-columns: repeat(1, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .journal-soft-surface {
    --journal-soft-button-height: 36px;
  }
}
</style>
