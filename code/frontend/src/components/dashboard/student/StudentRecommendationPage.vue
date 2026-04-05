<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Crosshair, ShieldAlert, Sparkles } from 'lucide-vue-next'

import type { RecommendationItem } from '@/api/contracts'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'

const props = defineProps<{
  weakDimensions: string[]
  recommendations: RecommendationItem[]
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
  openChallenges: []
  openSkillProfile: []
}>()

const headline = computed(() => props.weakDimensions[0] || '当前训练结构较均衡')
const topRecs = computed(() => props.recommendations.slice(0, 3))
</script>

<template>
  <section class="journal-shell space-y-6 journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div>
          <div class="journal-eyebrow">Priority Focus</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            补短板计划
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
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
            <span
              v-else
              class="inline-flex items-center rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700"
            >
              暂无明显短板
            </span>
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

      <div class="recommend-board mt-6 px-1 pt-5 md:px-2 md:pt-6">
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
                    <span
                      class="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs font-medium uppercase text-slate-500"
                    >
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
            <button class="journal-btn-outline" @click="emit('openChallenges')">浏览全部</button>
          </div>

          <div
            v-if="recommendations.length === 0"
            class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
          >
            当前没有推荐题目，可以先去挑战列表探索新的方向。
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
                    <span
                      class="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs font-medium uppercase text-slate-500"
                    >
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
.journal-shell {
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
}

.journal-note {
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
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

.journal-btn-outline {
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

.recommend-board {
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.recommend-section + .recommend-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.recommend-list {
  border-radius: 22px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

.recommend-item {
  padding: 1rem 1.1rem;
  transition: background 0.2s ease-in-out;
}

.recommend-item + .recommend-item {
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
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
  font-size: 0.875rem;
  font-weight: 600;
  flex-shrink: 0;
}

.rec-index--top {
  background: var(--journal-accent);
  color: color-mix(in srgb, white 92%, var(--color-bg-base));
}

.rec-index--rest {
  border: 1px solid rgba(99, 102, 241, 0.2);
  background: rgba(99, 102, 241, 0.07);
  color: var(--journal-accent);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: color-mix(in srgb, var(--color-text-primary) 88%, var(--color-text-secondary));
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
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
:global([data-theme='dark']) .recommend-list {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}
</style>
