<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, Radar, ShieldAlert } from 'lucide-vue-next'

import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { formatDate } from '@/utils/format'

import type { StudentOverviewProps } from './overviewProps'
import { timelineSummary } from './utils'

const props = defineProps<StudentOverviewProps>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()

const quickRecommendations = computed(() => props.recommendations.slice(0, 3))
const recentTimeline = computed(() => props.timeline.slice(0, 3))
const headlineWeakness = computed(() => props.weakDimensions[0] || '暂无明显短板')
const primaryStats = computed(() => [
  { label: '总得分', value: props.progress.total_score ?? 0, hint: '累计训练积分' },
  { label: '已解题数', value: props.progress.total_solved ?? 0, hint: '成功提交并判定正确' },
  { label: '当前排名', value: `#${props.progress.rank ?? '-'}`, hint: '全站训练表现' },
  { label: '完成率', value: `${props.completionRate}%`, hint: '按当前题量估算' },
])
</script>

<template>
  <div class="command-shell space-y-6">
    <section class="grid gap-4 xl:grid-cols-[1.18fr_0.82fr]">
      <article class="command-hero overflow-hidden rounded-[30px] border px-6 py-6 md:px-8">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.26em] text-cyan-100/70">
          <span>Threat Board</span>
          <span class="rounded-full border border-white/14 px-3 py-1">{{ className || '自由训练' }}</span>
        </div>
        <h2 class="mt-4 max-w-3xl text-3xl font-semibold tracking-tight text-white md:text-[2.45rem]">
          {{ displayName }} 的攻防训练指挥台
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-cyan-50/78">
          这套方案强调“战情总览”的感觉。先读态势，再落点到推荐靶场，适合 CTF 平台主面板的高压和决策感。
        </p>

        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="command-status-card">
            <div class="command-status-label">当前排名</div>
            <div class="command-status-value">#{{ progress.rank ?? '-' }}</div>
            <div class="command-status-hint">在全站训练中的相对位置</div>
          </div>
          <div class="command-status-card">
            <div class="command-status-label">完成率</div>
            <div class="command-status-value">{{ completionRate }}%</div>
            <div class="command-status-hint">反映覆盖范围，而不是难度深度</div>
          </div>
          <div class="command-status-card">
            <div class="command-status-label">优先补位</div>
            <div class="command-status-value">{{ headlineWeakness }}</div>
            <div class="command-status-hint">建议以推荐队列作为今日入口</div>
          </div>
        </div>

        <div class="mt-6 flex flex-wrap gap-3">
          <ElButton type="primary" @click="emit('openChallenges')">继续训练</ElButton>
          <ElButton plain @click="emit('openSkillProfile')">查看能力画像</ElButton>
        </div>
      </article>

      <div class="grid gap-3">
        <article
          v-for="item in highlightItems"
          :key="item.label"
          class="command-highlight rounded-[26px] border px-5 py-5"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-cyan-100/62">{{ item.label }}</div>
              <div class="mt-3 text-[28px] font-semibold tracking-tight text-white">{{ item.value }}</div>
            </div>
            <div class="flex h-12 w-12 items-center justify-center rounded-2xl border border-cyan-300/18 bg-cyan-300/10 text-cyan-100">
              <component :is="item.icon" class="h-5 w-5" />
            </div>
          </div>
          <p class="mt-3 text-sm leading-6 text-slate-300/78">{{ item.description }}</p>
        </article>
      </div>
    </section>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <article
        v-for="item in primaryStats"
        :key="item.label"
        class="command-metric rounded-[22px] border px-5 py-4"
      >
        <div class="text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-400">{{ item.label }}</div>
        <div class="mt-3 text-[30px] font-semibold tracking-tight text-white">{{ item.value }}</div>
        <div class="mt-2 text-sm text-slate-300/72">{{ item.hint }}</div>
      </article>
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.06fr_0.94fr]">
      <article class="command-panel rounded-[28px] border px-6 py-6">
        <div class="flex items-center justify-between gap-4">
          <div>
            <div class="text-[11px] font-semibold uppercase tracking-[0.26em] text-cyan-100/62">Priority Queue</div>
            <h3 class="mt-2 text-2xl font-semibold text-white">优先训练队列</h3>
            <p class="mt-2 text-sm leading-6 text-slate-300/76">保持同一份推荐逻辑，只把展示做成战情卡片。</p>
          </div>
          <ShieldAlert class="hidden h-10 w-10 text-cyan-100/72 md:block" />
        </div>

        <div v-if="quickRecommendations.length === 0" class="mt-6 rounded-[22px] border border-dashed border-white/12 px-4 py-10 text-center text-sm text-slate-300/72">
          当前没有推荐题目，直接去挑战列表挑一道新题即可。
        </div>

        <div v-else class="mt-6 space-y-3">
          <button
            v-for="(item, index) in quickRecommendations"
            :key="item.challenge_id"
            type="button"
            class="command-rec-item flex w-full items-start gap-4 rounded-[24px] border px-4 py-4 text-left transition"
            @click="emit('openChallenge', item.challenge_id)"
          >
            <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-cyan-300/18 bg-cyan-300/10 text-lg font-semibold text-white">
              0{{ index + 1 }}
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div class="text-base font-semibold text-white">{{ item.title }}</div>
                  <div class="mt-1 text-xs uppercase tracking-[0.18em] text-slate-400">{{ item.category }}</div>
                </div>
                <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="difficultyClass(item.difficulty)">
                  {{ difficultyLabel(item.difficulty) }}
                </span>
              </div>
              <p class="mt-3 text-sm leading-6 text-slate-300/76">{{ item.reason }}</p>
            </div>
            <ArrowRight class="mt-1 h-4 w-4 shrink-0 text-cyan-100/78" />
          </button>
        </div>
      </article>

      <div class="grid gap-4">
        <article class="command-panel rounded-[28px] border px-6 py-6">
          <div class="flex items-center gap-3">
            <Radar class="h-5 w-5 text-cyan-100/84" />
            <h3 class="text-xl font-semibold text-white">待加强维度</h3>
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <span
              v-for="item in weakDimensions.slice(0, 4)"
              :key="item"
              class="rounded-full border border-amber-400/22 bg-amber-400/10 px-3 py-1 text-xs font-medium text-amber-100"
            >
              {{ item }}
            </span>
            <span
              v-if="weakDimensions.length === 0"
              class="rounded-full border border-emerald-400/22 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-100"
            >
              结构均衡
            </span>
          </div>
          <p class="mt-4 text-sm leading-6 text-slate-300/74">适合在继续训练前快速决定今天先解哪一类题。</p>
        </article>

        <article class="command-panel rounded-[28px] border px-6 py-6">
          <div class="flex items-center justify-between gap-4">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-cyan-100/62">Recent Signal</div>
              <h3 class="mt-2 text-xl font-semibold text-white">近期速览</h3>
            </div>
            <div class="text-sm text-slate-400">{{ timeline.length }} 条</div>
          </div>

          <div v-if="recentTimeline.length === 0" class="mt-5 rounded-[22px] border border-dashed border-white/12 px-4 py-10 text-center text-sm text-slate-300/72">
            当前还没有训练动态。
          </div>

          <div v-else class="mt-5 space-y-3">
            <article
              v-for="event in recentTimeline"
              :key="event.id"
              class="rounded-[22px] border border-white/10 bg-white/4 px-4 py-4"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="text-sm font-medium text-white">{{ event.title }}</div>
                <div class="text-xs text-slate-400">{{ formatDate(event.created_at) }}</div>
              </div>
              <div class="mt-2 text-sm leading-6 text-slate-300/74">{{ timelineSummary(event) }}</div>
            </article>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.command-shell {
  font-family: "Rajdhani", "DIN Alternate", "Bahnschrift", "Noto Sans SC", sans-serif;
}

.command-hero {
  border-color: rgba(34, 211, 238, 0.18);
  background:
    radial-gradient(circle at top left, rgba(34, 211, 238, 0.18), transparent 22rem),
    linear-gradient(135deg, rgba(7, 24, 39, 0.92), rgba(6, 78, 110, 0.7));
}

.command-highlight,
.command-panel,
.command-metric {
  border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.78), rgba(2, 6, 23, 0.88));
}

.command-status-card {
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 24px;
  background: rgba(2, 6, 23, 0.24);
  padding: 1rem 1.1rem;
}

.command-status-label {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: rgba(165, 243, 252, 0.72);
}

.command-status-value {
  margin-top: 0.8rem;
  font-size: 1.65rem;
  font-weight: 700;
  color: #fff;
}

.command-status-hint {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: rgba(226, 232, 240, 0.72);
}

.command-rec-item {
  border-color: rgba(255, 255, 255, 0.08);
  background: linear-gradient(180deg, rgba(8, 47, 73, 0.4), rgba(15, 23, 42, 0.62));
}

.command-rec-item:hover {
  border-color: rgba(34, 211, 238, 0.32);
  transform: translateY(-1px);
}
</style>
