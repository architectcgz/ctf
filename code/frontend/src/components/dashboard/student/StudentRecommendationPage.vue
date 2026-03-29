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
  <div class="journal-shell space-y-6">
    <!-- Hero -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div>
          <div class="journal-eyebrow">Priority Focus</div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            补短板计划
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            根据当前薄弱维度给出优先训练顺序，建议先完成靠前题目，再回看能力画像确认是否抬升。
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
              <div class="journal-note-value">{{ weakDimensions.length > 0 ? weakDimensions.length + ' 项' : '暂无' }}</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">即将可做</div>
              <div class="journal-note-value">{{ topRecs.length }} 道</div>
            </div>
          </div>
        </article>
      </div>
    </section>

    <!-- 优先推荐前三 -->
    <section v-if="topRecs.length > 0" class="journal-panel rounded-[24px] border px-6 py-6">
      <div class="journal-eyebrow">Top Queue</div>
      <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">优先推荐</h3>
      <div class="mt-5 grid gap-4 md:grid-cols-3">
        <article
          v-for="(item, index) in topRecs"
          :key="item.challenge_id"
          class="journal-metric rounded-[18px] border px-4 py-4 cursor-pointer"
          @click="emit('openChallenge', item.challenge_id)"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--journal-muted)]">
              Queue {{ index + 1 }}
            </div>
            <Crosshair class="h-4 w-4 text-[var(--journal-accent)]" />
          </div>
          <div class="mt-3 text-base font-semibold leading-snug text-[var(--journal-ink)]">{{ item.title }}</div>
          <div class="mt-2 flex flex-wrap items-center gap-2">
            <span class="text-xs uppercase tracking-wide text-[var(--journal-muted)]">{{ item.category }}</span>
            <span class="h-1 w-1 rounded-full bg-slate-300" />
            <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="difficultyClass(item.difficulty)">
              {{ difficultyLabel(item.difficulty) }}
            </span>
          </div>
        </article>
      </div>
    </section>

    <!-- 完整推荐列表 -->
    <section class="journal-panel rounded-[24px] border px-6 py-6">
      <div class="flex items-start justify-between gap-4">
        <div>
          <div class="journal-eyebrow">Full List</div>
          <h3 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">推荐列表</h3>
        </div>
        <button
          class="journal-btn-outline"
          @click="emit('openChallenges')"
        >
          浏览全部
        </button>
      </div>

      <div
        v-if="recommendations.length === 0"
        class="mt-5 rounded-[22px] border border-dashed border-[var(--journal-border)] px-4 py-12 text-center text-sm text-[var(--journal-muted)]"
      >
        当前没有推荐题目，可以先去挑战列表探索新的方向。
      </div>

      <div v-else class="mt-5 space-y-3">
        <button
          v-for="(item, index) in recommendations"
          :key="item.challenge_id"
          class="journal-rec-item group w-full cursor-pointer rounded-[18px] border px-5 py-4 text-left"
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
                <span class="text-sm font-semibold text-[var(--journal-ink)]">{{ item.title }}</span>
                <span
                  class="rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="difficultyClass(item.difficulty)"
                >
                  {{ difficultyLabel(item.difficulty) }}
                </span>
                <span class="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs font-medium uppercase text-slate-500">
                  {{ item.category }}
                </span>
              </div>
              <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">{{ item.reason }}</p>
            </div>
            <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-[var(--journal-accent-strong)] opacity-0 transition group-hover:opacity-100" />
          </div>
        </button>
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

.journal-panel {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.journal-metric {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
  transition: all 0.2s ease-in-out;
}

.journal-metric:hover {
  transform: translateY(-2px);
  border-color: #6366f1;
  box-shadow: 0 14px 26px rgba(15, 23, 42, 0.06);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-rec-item {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
  transition: all 0.2s ease-in-out;
}

.journal-rec-item:hover {
  transform: translateY(-2px);
  border-color: #6366f1;
  box-shadow: 0 14px 26px rgba(15, 23, 42, 0.06);
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
  color: #fff;
}

.rec-index--rest {
  border: 1px solid rgba(99, 102, 241, 0.2);
  background: rgba(99, 102, 241, 0.07);
  color: var(--journal-accent);
}
</style>
