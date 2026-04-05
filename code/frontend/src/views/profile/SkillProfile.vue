<script setup lang="ts">
import { ChevronRight, Flame, Loader2, TriangleAlert } from 'lucide-vue-next'

import RadarChart from '@/components/charts/RadarChart.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useSkillProfilePage } from '@/composables/useSkillProfilePage'
import { difficultyLabel } from '@/utils/challenge'

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
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col space-y-6 rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div v-if="loading" class="space-y-6">
      <div class="space-y-6">
        <div class="h-12 animate-pulse rounded-2xl bg-[var(--journal-surface)]/90"></div>
        <div class="grid gap-6 xl:grid-cols-[minmax(0,1.06fr)_minmax(300px,0.94fr)]">
          <div class="h-80 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
          <div class="h-80 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
        </div>
        <div class="h-56 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
        <div class="h-56 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
      </div>
    </div>

    <div v-else-if="error" class="py-8 text-center">
      <TriangleAlert class="mx-auto h-10 w-10 text-[var(--color-danger)]" />
      <p class="mt-3 text-sm text-[var(--color-danger)]">{{ error }}</p>
      <button type="button" class="journal-btn journal-btn--primary mt-4" @click="loadCurrentData">
        重试
      </button>
    </div>

    <!-- 空状态 -->
    <AppEmpty
      v-else-if="!skillProfile"
      title="暂无能力画像数据"
      description="完成更多靶场挑战后，系统将为你生成能力画像。"
      icon="Radar"
    />

    <div v-else class="flex flex-1 flex-col">
      <div>
        <div class="journal-eyebrow">Skill Profile</div>
        <h2
          class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
        >
          能力画像
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          查看能力画像和推荐靶场。
        </p>

        <div class="mt-6 flex flex-wrap gap-3">
          <button type="button" class="journal-btn" @click="loadCurrentData">刷新</button>
          <button type="button" class="journal-btn journal-btn--primary" @click="goToChallenges">
            去做题
          </button>
        </div>
      </div>

      <div v-if="isTeacher" class="skill-teacher-panel mt-6">
        <div class="journal-eyebrow journal-eyebrow-soft">Teacher View</div>
        <h3 class="mt-3 text-base font-semibold text-[var(--journal-ink)]">查看学员能力画像</h3>
        <select
          v-model="selectedStudentId"
          class="skill-student-select mt-3 w-full max-w-sm"
        >
          <option value="">我的能力画像</option>
          <option v-for="student in students" :key="student.id" :value="student.id">
            {{ student.name || student.username }} ({{ student.username }})
          </option>
        </select>
      </div>

      <div class="skill-board mt-6 px-1 pt-5 md:px-2 md:pt-6">
        <section class="skill-section">
          <div class="skill-analysis-stack">
            <div>
              <div class="journal-eyebrow journal-eyebrow-soft">Radar Analysis</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">能力维度分析</h3>

              <div class="skill-dimension-wrap mt-5">
                <div class="skill-dimension-list mt-2">
                  <div class="skill-dimension-chart">
                    <div class="skill-dimension-chart__frame">
                      <div class="skill-dimension-chart__inner">
                        <RadarChart
                          :indicators="radarIndicators"
                          :values="radarValues"
                          name="维度得分"
                          height-class="h-[30rem] md:h-[34rem] xl:h-[38rem]"
                          :label-font-size="20"
                          :axis-name-gap="12"
                          radius="74%"
                          center-y="50%"
                        />
                      </div>
                    </div>
                  </div>

                  <div class="skill-dimension-legend">
                    <article
                      v-for="dim in skillProfile.dimensions"
                      :key="dim.name"
                      class="skill-dimension-legend__item"
                    >
                      <div class="min-w-0">
                        <div class="text-base font-semibold text-[var(--journal-ink)] md:text-[1.05rem]">{{ dim.name }}</div>
                        <div class="mt-1 text-[0.8rem] text-[var(--journal-muted)]">当前维度表现</div>
                      </div>
                      <div class="text-right">
                        <div class="text-[1.9rem] font-semibold tracking-tight text-[var(--journal-ink)] tech-font md:text-[2.1rem]">
                          {{ dim.value }}
                        </div>
                        <div class="text-xs text-[var(--journal-muted)]">/ 100</div>
                      </div>
                    </article>
                  </div>
                </div>
              </div>

              <div class="skill-weak-wrap">
                <div class="journal-eyebrow journal-eyebrow-soft">Weak Points</div>
                <div class="mt-3 flex items-center gap-3 text-base font-semibold text-[var(--journal-ink)]">
                  <Flame class="h-5 w-5 text-[var(--journal-accent)]" />
                  薄弱项提示
                </div>
                <div v-if="weakDimensions.length > 0" class="skill-weak-list mt-5">
                  <div v-for="dim in weakDimensions.slice(0, 4)" :key="dim" class="skill-weak-item">
                    <div class="journal-note-label">建议加强</div>
                    <div class="mt-2 text-sm font-semibold text-[var(--journal-ink)]">{{ dim }}</div>
                  </div>
                </div>
                <div v-else class="skill-weak-list mt-5">
                  <div class="skill-weak-item">
                    <div class="journal-note-label">当前状态</div>
                    <div class="mt-2 text-sm font-semibold text-[var(--journal-ink)]">暂时没有明显短板</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="skill-section">
          <div class="journal-eyebrow journal-eyebrow-soft">Recommendations</div>
          <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">推荐靶场</h3>
          <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
            优先从当前最匹配的题目开始。
          </p>

          <div
            v-if="loadingRecommendations"
            class="mt-6 flex items-center gap-3 text-sm text-[var(--journal-muted)]"
          >
            <Loader2 class="h-4 w-4 animate-spin" />
            加载推荐中…
          </div>

          <div v-else-if="recommendations.length === 0" class="mt-6 text-sm text-[var(--journal-muted)]">
            暂无推荐靶场，完成更多题目后会自动生成。
          </div>

          <div v-else class="skill-recommend-list mt-5">
            <button
              v-for="item in recommendations"
              :key="item.challenge_id"
              type="button"
              class="skill-recommend-item w-full text-left"
              @click="goToChallenge(item.challenge_id)"
            >
              <div class="flex items-center justify-between gap-4">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
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
      </div>
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
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

.journal-note-label {
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
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

.skill-teacher-panel {
  border-radius: 22px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  padding: 1rem 1.1rem;
}

.skill-student-select {
  border-radius: 14px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  padding: 0.7rem 0.95rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition: border-color 0.2s;
}

.skill-student-select:focus {
  border-color: var(--journal-accent);
}

.skill-board {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.skill-section + .skill-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.skill-analysis-stack {
  display: block;
}

.skill-dimension-wrap {
  margin-top: 1.25rem;
}

.skill-weak-wrap {
  margin-top: 1.75rem;
  padding-top: 1.5rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.skill-recommend-list,
.skill-weak-list {
  border-radius: 22px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.skill-weak-item,
.skill-recommend-item {
  padding: 1rem 1.1rem;
}

.skill-weak-item + .skill-weak-item,
.skill-recommend-item + .skill-recommend-item {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.skill-dimension-list {
  display: grid;
  gap: 1.25rem;
}

.skill-dimension-chart__frame {
  position: relative;
  margin: 0 auto;
  width: min(100%, 520px);
  aspect-ratio: 1.04;
  overflow: visible;
}

.skill-dimension-chart__frame::before,
.skill-dimension-chart__frame::after {
  content: '';
  position: absolute;
  pointer-events: none;
}

.skill-dimension-chart__frame::before {
  inset: 0;
  clip-path: polygon(25% 6%, 75% 6%, 100% 50%, 75% 94%, 25% 94%, 0 50%);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 94%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 96%, var(--color-bg-base))
    ),
    linear-gradient(135deg, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent);
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.skill-dimension-chart__frame::after {
  inset: 18px;
  clip-path: polygon(25% 6%, 75% 6%, 100% 50%, 75% 94%, 25% 94%, 0 50%);
  border: 1px solid color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 78%, transparent);
  background:
    radial-gradient(circle at 50% 45%, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 60%),
    color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 76%, var(--color-bg-base));
}

.skill-dimension-chart__inner {
  position: absolute;
  inset: 18px;
  z-index: 1;
}

.skill-dimension-legend {
  display: grid;
  gap: 0;
}

.skill-dimension-legend__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.15rem 1.2rem;
}

@media (min-width: 640px) {
  .skill-dimension-legend {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .skill-dimension-legend__item {
    min-height: 100%;
  }

  .skill-dimension-legend__item:nth-child(n + 3) {
    border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  }

  .skill-dimension-legend__item:nth-child(2n) {
    border-left: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  }
}

@media (max-width: 639px) {
  .skill-dimension-legend__item + .skill-dimension-legend__item {
    border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  }
}

@media (min-width: 1024px) {
  .skill-dimension-list {
    grid-template-columns: minmax(0, 1.18fr) minmax(260px, 0.82fr);
    align-items: center;
  }
}

@media (min-width: 768px) {
  .skill-weak-list {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .skill-weak-item + .skill-weak-item {
    border-top: 0;
  }

  .skill-weak-item:nth-child(n + 3) {
    border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  }

  .skill-weak-item:nth-child(2n) {
    border-left: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
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
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 16%, transparent), transparent 20rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
}

:global([data-theme='dark']) .skill-teacher-panel,
:global([data-theme='dark']) .skill-recommend-list,
:global([data-theme='dark']) .skill-weak-list,
:global([data-theme='dark']) .skill-student-select {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

.skill-recommend-item {
  transition: background 0.2s;
}

.skill-recommend-item:hover {
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
}

.tech-font {
  font-family: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
}

:global([data-theme='dark']) .skill-dimension-chart__frame {
}

:global([data-theme='dark']) .skill-dimension-chart__frame::before {
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    ),
    linear-gradient(135deg, color-mix(in srgb, var(--journal-accent) 18%, transparent), transparent);
}

:global([data-theme='dark']) .skill-dimension-chart__frame::after {
  background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  border-color: rgba(148, 163, 184, 0.2);
}

</style>
