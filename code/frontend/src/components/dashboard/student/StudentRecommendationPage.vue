<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Crosshair, ShieldAlert, Sparkles } from 'lucide-vue-next'

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

const headline = computed(() => props.weakDimensions[0] || '当前训练结构较均衡')
const topRecs = computed(() => props.recommendations.slice(0, 3))
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
    <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
      <div>
        <div class="journal-eyebrow">Priority Focus</div>
        <h1 class="journal-page-title workspace-tab-heading__title text-[var(--journal-ink)]">
          补短板计划
        </h1>
        <p class="workspace-tab-copy max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          优先看最适合当前阶段的题目。
        </p>
        <div class="mt-5 flex flex-wrap gap-2">
          <template v-if="weakDimensions.length > 0">
            <span
              v-for="dim in weakDimensions.slice(0, 4)"
              :key="dim"
              class="inline-flex items-center gap-1.5 rounded-full border border-[var(--journal-accent)]/20 bg-[var(--journal-accent)]/8 px-3 py-1 text-xs font-semibold text-[var(--journal-accent-strong)]"
            >
              <ShieldAlert class="h-3 w-3" />
              {{ dim }}
            </span>
          </template>
          <span v-else class="journal-weak-tag journal-weak-tag--stable"> 暂无明显短板 </span>
        </div>
      </div>

      <article class="journal-brief rounded-[24px] border px-5 py-5">
        <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
          <Sparkles class="h-5 w-5 text-[var(--journal-accent)]" />
          推荐摘要
        </div>
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
          <div class="journal-note">
            <div class="journal-note-label">当前首要关注</div>
            <div class="journal-note-value">{{ headline }}</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">推荐队列</div>
            <div class="journal-note-value">{{ recommendations.length }} 项</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">薄弱维度</div>
            <div class="journal-note-value">
              {{ weakDimensions.length > 0 ? weakDimensions.length + ' 项' : '暂无' }}
            </div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">即将可做</div>
            <div class="journal-note-value">{{ topRecs.length }} 道</div>
          </div>
        </div>
      </article>
    </div>

    <div
      class="recommend-board mt-6 px-1 pt-5 md:px-2 md:pt-6"
      :class="{ 'recommend-board--embedded': embedded }"
    >
      <section v-if="topRecs.length > 0" class="recommend-section">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="journal-eyebrow journal-eyebrow-soft">Top Queue</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">优先推荐</h3>
          </div>
          <button class="journal-btn-outline" @click="emit('openSkillProfile')">看画像</button>
        </div>

        <div class="recommend-list mt-5">
          <button
            v-for="(item, index) in topRecs"
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
              <Crosshair class="mt-1 h-4 w-4 shrink-0 text-[var(--journal-accent)]" />
            </div>
          </button>
        </div>
      </section>

      <section class="recommend-section">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="journal-eyebrow journal-eyebrow-soft">Full List</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">推荐列表</h3>
          </div>
          <button class="journal-btn-primary" @click="emit('openChallenges')">浏览全部</button>
        </div>

        <div
          v-if="recommendations.length === 0"
          class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-shell-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
        >
          当前没有推荐题目，可以先去题目列表探索新的方向。
        </div>

        <div v-else class="recommend-list mt-5">
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

.journal-brief {
  border-color: var(--journal-shell-border);
  background: var(--journal-surface-subtle);
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

.recommend-section + .recommend-section {
  margin-top: var(--space-6);
  padding-top: var(--space-6);
  border-top: 1px solid var(--journal-divider);
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

@media (max-width: 767px) {
  .journal-soft-surface {
    --journal-soft-button-height: 36px;
  }
}
</style>
