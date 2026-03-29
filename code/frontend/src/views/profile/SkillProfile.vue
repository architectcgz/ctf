<script setup lang="ts">
import { ChevronRight, Flame, Loader2, Target, TriangleAlert } from 'lucide-vue-next'

import RadarChart from '@/components/charts/RadarChart.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useSkillProfilePage } from '@/composables/useSkillProfilePage'
import { difficultyLabel } from '@/utils/challenge'
import { formatDate } from '@/utils/format'

const {
  isTeacher,
  selectedStudentId,
  students,
  loading,
  error,
  skillProfile,
  loadingRecommendations,
  recommendations,
  weakDimensions,
  radarIndicators,
  radarValues,
  loadCurrentData,
  goToChallenge,
  goToChallenges,
} = useSkillProfilePage()

const difficultyColorMap: Record<string, string> = {
  beginner: '#10b981',
  easy: '#22d3ee',
  medium: '#f59e0b',
  hard: '#f97316',
  insane: '#ef4444',
}
</script>

<template>
  <div class="journal-shell space-y-6">
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Skill Profile</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            能力画像
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            在这里查看你的能力画像和推荐题目。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="journal-btn" @click="loadCurrentData">刷新</button>
            <button type="button" class="journal-btn journal-btn--primary" @click="goToChallenges">
              去做题
            </button>
          </div>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <Target class="h-5 w-5 text-[var(--journal-accent)]" />
            当前画像概况
          </div>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div class="journal-note">
              <div class="journal-note-label">维度数量</div>
              <div class="journal-note-value">{{ skillProfile?.dimensions?.length ?? 0 }}</div>
              <div class="journal-note-helper">当前被纳入画像的技术维度</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">薄弱项提示</div>
              <div class="journal-note-value">{{ weakDimensions.length }}</div>
              <div class="journal-note-helper">优先补强的薄弱维度数量</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">推荐靶场</div>
              <div class="journal-note-value">{{ recommendations.length }}</div>
              <div class="journal-note-helper">当前可直接进入的推荐题目数</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">最近同步</div>
              <div class="journal-note-value text-sm">
                {{ skillProfile?.updated_at ? formatDate(skillProfile.updated_at) : '待同步' }}
              </div>
              <div class="journal-note-helper">画像刷新后会同步更新这个时间</div>
            </div>
          </div>
        </article>
      </div>

      <div
        v-if="isTeacher"
        class="mt-6 rounded-[18px] border border-[var(--journal-border)] bg-[var(--journal-surface-subtle)] px-4 py-4"
      >
        <div class="journal-eyebrow">Teacher View</div>
        <h3 class="mt-3 text-base font-semibold text-[var(--journal-ink)]">查看学员能力画像</h3>
        <select
          v-model="selectedStudentId"
          class="mt-3 w-full max-w-sm rounded-[14px] border border-[var(--journal-border)] bg-[var(--journal-surface)] px-4 py-2.5 text-sm text-[var(--journal-ink)] outline-none transition focus:border-[var(--journal-accent)]"
        >
          <option value="">我的能力画像</option>
          <option v-for="student in students" :key="student.id" :value="student.id">
            {{ student.name || student.username }} ({{ student.username }})
          </option>
        </select>
      </div>
    </section>

    <!-- 加载骨架 -->
    <div v-if="loading" class="space-y-6">
      <div class="journal-hero rounded-[30px] border px-6 py-6">
        <div class="h-8 w-40 animate-pulse rounded-2xl bg-[var(--journal-surface)]"></div>
        <div class="mt-4 h-80 animate-pulse rounded-2xl bg-[var(--journal-surface)]"></div>
      </div>
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <div
          v-for="i in 3"
          :key="i"
          class="h-28 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"
        ></div>
      </div>
      <div class="journal-hero rounded-[30px] border px-6 py-6">
        <div class="h-6 w-24 animate-pulse rounded-2xl bg-[var(--journal-surface)]"></div>
        <div class="mt-4 space-y-3">
          <div class="h-16 animate-pulse rounded-[20px] bg-[var(--journal-surface)]"></div>
          <div class="h-16 animate-pulse rounded-[20px] bg-[var(--journal-surface)]"></div>
          <div class="h-16 animate-pulse rounded-[20px] bg-[var(--journal-surface)]"></div>
        </div>
      </div>
    </div>

    <!-- 错误状态 -->
    <section v-else-if="error" class="journal-hero rounded-[30px] border px-6 py-8 text-center">
      <TriangleAlert class="mx-auto h-10 w-10 text-[var(--color-danger)]" />
      <p class="mt-3 text-sm text-[var(--color-danger)]">{{ error }}</p>
      <button type="button" class="journal-btn journal-btn--primary mt-4" @click="loadCurrentData">
        重试
      </button>
    </section>

    <!-- 空状态 -->
    <AppEmpty
      v-else-if="!skillProfile"
      title="暂无能力画像数据"
      description="完成更多靶场挑战后，系统将为你生成能力画像。"
      icon="Radar"
    />

    <!-- 主内容 -->
    <template v-else>
      <!-- Hero：雷达图 -->
      <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
        <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
          <div>
            <div class="journal-eyebrow">Radar Analysis</div>
            <h2
              class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
            >
              能力维度分析
            </h2>
            <p class="mt-3 max-w-xl text-sm leading-7 text-[var(--journal-muted)]">
              这里能看到各个维度的大致情况。
            </p>
            <p v-if="skillProfile.updated_at" class="mt-3 text-xs text-[var(--journal-muted)]">
              最近更新：{{ formatDate(skillProfile.updated_at) }}
            </p>
            <div
              class="mt-4 rounded-[16px] border border-[var(--journal-border)] bg-[var(--journal-surface)] px-4 py-3"
            >
              <div class="text-sm text-[var(--journal-muted)]">
                最近同步：{{ formatDate(skillProfile.updated_at) }}
              </div>
            </div>
          </div>

          <!-- 薄弱项简报 -->
          <article
            v-if="weakDimensions.length > 0"
            class="journal-brief rounded-[24px] border px-5 py-5"
          >
            <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
              <Flame class="h-5 w-5 text-[var(--journal-accent)]" />
              薄弱项提示
            </div>
            <div class="mt-4 space-y-2">
              <div
                v-for="dim in weakDimensions.slice(0, 3)"
                :key="dim"
                class="journal-note rounded-[16px] border px-4 py-3"
              >
                <div class="journal-note-label">建议加强</div>
                <div class="journal-note-value">{{ dim }}</div>
              </div>
            </div>
          </article>
          <article v-else class="journal-brief rounded-[24px] border px-5 py-5">
            <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
              <Target class="h-5 w-5 text-[var(--journal-accent)]" />
              能力覆盖
            </div>
            <p class="mt-4 text-sm leading-6 text-[var(--journal-muted)]">暂时没有明显短板。</p>
          </article>
        </div>

        <div class="mt-6">
          <RadarChart :indicators="radarIndicators" :values="radarValues" name="能力画像" />
        </div>
      </section>

      <!-- 各维度 bento -->
      <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="dim in skillProfile.dimensions"
          :key="dim.name"
          class="journal-metric rounded-[24px] border px-5 py-5"
        >
          <div class="journal-eyebrow">{{ dim.name }}</div>
          <div
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] tech-font"
          >
            {{ dim.value }}
          </div>
          <div class="mt-1 text-xs text-[var(--journal-muted)]">/ 100</div>
          <div class="mt-4 h-1.5 rounded-full" style="background: rgba(226, 232, 240, 0.5)">
            <div
              class="h-1.5 rounded-full transition-all duration-700"
              :style="{ width: dim.value + '%', background: 'var(--journal-accent)' }"
            />
          </div>
        </article>
      </section>

      <!-- 推荐靶场 -->
      <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
        <div class="journal-eyebrow">Recommendations</div>
        <h2 class="mt-3 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
          推荐靶场
        </h2>
        <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
          这里会列出当前更适合你的题目。
        </p>

        <div
          v-if="loadingRecommendations"
          class="mt-6 flex items-center gap-3 text-sm text-[var(--journal-muted)]"
        >
          <Loader2 class="h-4 w-4 animate-spin" />
          加载推荐中…
        </div>

        <div
          v-else-if="recommendations.length === 0"
          class="mt-6 text-sm text-[var(--journal-muted)]"
        >
          暂无推荐靶场，完成更多题目后将自动生成。
        </div>

        <div v-else class="mt-6 space-y-3">
          <button
            v-for="item in recommendations"
            :key="item.challenge_id"
            type="button"
            class="journal-rec-item w-full rounded-[20px] border px-5 py-4 text-left transition"
            @click="goToChallenge(item.challenge_id)"
          >
            <div class="flex items-center justify-between gap-4">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-semibold text-[var(--journal-ink)]">{{
                    item.title
                  }}</span>
                  <span
                    class="shrink-0 rounded-full px-2 py-0.5 text-[11px] font-semibold"
                    :style="{
                      background: (difficultyColorMap[item.difficulty] || '#94a3b8') + '22',
                      color: difficultyColorMap[item.difficulty] || '#94a3b8',
                    }"
                  >
                    {{ difficultyLabel(item.difficulty) }}
                  </span>
                </div>
                <p class="mt-1 text-xs leading-5 text-[var(--journal-muted)]">{{ item.reason }}</p>
              </div>
              <ChevronRight class="h-4 w-4 shrink-0 text-[var(--journal-accent-strong)]" />
            </div>
          </button>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: #ffffff;
  --journal-surface-subtle: rgba(248, 250, 252, 0.92);
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.98), rgba(241, 245, 249, 0.95));
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-metric {
  border-color: var(--journal-border);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.journal-metric:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 35%, transparent);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.08);
}

.journal-brief {
  border-color: var(--journal-border);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-note {
  border-color: var(--journal-border);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.9));
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-rec-item {
  border-color: var(--journal-border);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
  transition:
    border-color 0.2s,
    background 0.2s,
    box-shadow 0.2s;
}

.journal-rec-item:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 40%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
  box-shadow: 0 8px 20px rgba(79, 70, 229, 0.08);
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-ready {
  background: #10b981;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
  animation: pulse-dot 2s infinite;
}

@keyframes pulse-dot {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.tech-font {
  font-family: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
}

.journal-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 12px;
  border: 1px solid var(--color-border-default);
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
  background: transparent;
  transition:
    border-color 0.2s,
    color 0.2s;
  cursor: pointer;
}

.journal-btn:hover {
  border-color: var(--journal-accent);
  color: var(--journal-accent);
}

.journal-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-btn--primary:hover {
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
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
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}
</style>
