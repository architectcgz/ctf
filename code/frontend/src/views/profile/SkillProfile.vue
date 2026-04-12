<script setup lang="ts">
import { ref } from 'vue'
import { ChevronRight, Flame, Loader2, TriangleAlert } from 'lucide-vue-next'

import RadarChart from '@/components/charts/RadarChart.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useSkillProfilePage } from '@/composables/useSkillProfilePage'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'

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

type SkillProfileTabKey = 'analysis' | 'weakness' | 'recommendations'

const contentTabs: Array<{
  key: SkillProfileTabKey
  label: string
  buttonId: string
  panelId: string
}> = [
  {
    key: 'analysis',
    label: '能力维度分析',
    buttonId: 'skill-profile-tab-analysis',
    panelId: 'skill-profile-panel-analysis',
  },
  {
    key: 'weakness',
    label: '薄弱项提示',
    buttonId: 'skill-profile-tab-weakness',
    panelId: 'skill-profile-panel-weakness',
  },
  {
    key: 'recommendations',
    label: '推荐靶场',
    buttonId: 'skill-profile-tab-recommendations',
    panelId: 'skill-profile-panel-recommendations',
  },
]

const skillProfileTabSet = new Set<SkillProfileTabKey>(contentTabs.map((tab) => tab.key))

function resolveTabFromLocation(): SkillProfileTabKey {
  if (typeof window === 'undefined') return 'analysis'
  if (!window.location.pathname || window.location.pathname === '/') return 'analysis'
  const panel = new URLSearchParams(window.location.search).get('panel')
  if (panel && skillProfileTabSet.has(panel as SkillProfileTabKey)) {
    return panel as SkillProfileTabKey
  }
  return 'analysis'
}

function syncPanelToLocation(tabKey: SkillProfileTabKey): void {
  if (typeof window === 'undefined') return
  const url = new URL(window.location.href)
  url.searchParams.set('panel', tabKey)
  window.history.replaceState(window.history.state, '', `${url.pathname}${url.search}${url.hash}`)
}

const activeTab = ref<SkillProfileTabKey>(resolveTabFromLocation())
const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])

function selectTab(tabKey: SkillProfileTabKey): void {
  if (activeTab.value === tabKey) return
  activeTab.value = tabKey
  syncPanelToLocation(tabKey)
}

function setTabButtonRef(index: number, element: HTMLButtonElement | null): void {
  tabButtonRefs.value[index] = element
}

function focusTab(index: number): void {
  const normalizedIndex = (index + contentTabs.length) % contentTabs.length
  const nextTab = contentTabs[normalizedIndex]
  if (!nextTab) {
    return
  }
  selectTab(nextTab.key)
  tabButtonRefs.value[normalizedIndex]?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  switch (event.key) {
    case 'ArrowRight':
    case 'ArrowDown':
      event.preventDefault()
      focusTab(index + 1)
      break
    case 'ArrowLeft':
    case 'ArrowUp':
      event.preventDefault()
      focusTab(index - 1)
      break
    case 'Home':
      event.preventDefault()
      focusTab(0)
      break
    case 'End':
      event.preventDefault()
      focusTab(contentTabs.length - 1)
      break
  }
}
</script>

<template>
  <section
    class="journal-shell journal-shell-user journal-eyebrow-text journal-hero flex min-h-full flex-1 flex-col space-y-6 rounded-[30px] border px-6 py-6 md:px-8"
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
      description="完成更多靶场题目后，系统将为你生成能力画像。"
      icon="Radar"
    />

    <div v-else class="flex flex-1 flex-col">
      <div>
        <div class="journal-eyebrow">Skill Profile</div>

        <nav class="top-tabs" role="tablist" aria-label="能力画像内容切换">
          <button
            v-for="(tab, index) in contentTabs"
            :id="tab.buttonId"
            :key="tab.key"
            :ref="(element) => setTabButtonRef(index, element as HTMLButtonElement | null)"
            class="top-tab"
            :class="{ active: activeTab === tab.key }"
            type="button"
            role="tab"
            :tabindex="activeTab === tab.key ? 0 : -1"
            :aria-selected="activeTab === tab.key ? 'true' : 'false'"
            :aria-controls="tab.panelId"
            @click="selectTab(tab.key)"
            @keydown="handleTabKeydown($event, index)"
          >
            {{ tab.label }}
          </button>
        </nav>
      </div>

      <div v-if="isTeacher" class="skill-teacher-panel">
        <div class="journal-eyebrow journal-eyebrow-soft">Teacher View</div>
        <h3 class="workspace-tab-heading__title">查看学员能力画像</h3>
        <label for="skill-student-select" class="skill-field-label mt-3 block">选择学员</label>
        <select
          id="skill-student-select"
          v-model="selectedStudentId"
          class="skill-student-select mt-2 w-full max-w-sm"
        >
          <option value="">我的能力画像</option>
          <option v-for="student in students" :key="student.id" :value="student.id">
            {{ student.name || student.username }} ({{ student.username }})
          </option>
        </select>
      </div>

      <div class="skill-board px-1 md:px-2">
        <section
          id="skill-profile-panel-analysis"
          class="tab-panel skill-section"
          role="tabpanel"
          aria-labelledby="skill-profile-tab-analysis"
          :aria-hidden="activeTab === 'analysis' ? 'false' : 'true'"
          v-show="activeTab === 'analysis'"
        >
          <div class="skill-analysis-stack">
            <div>
              <div class="skill-overview-head">
                <h1 class="journal-page-title workspace-page-title text-[var(--journal-ink)]">
                  能力画像
                </h1>
                <p class="skill-overview-copy workspace-page-copy">
                  查看当前能力维度表现，并根据薄弱项获取推荐靶场。
                </p>
                <div class="skill-overview-actions" role="group" aria-label="能力画像快捷操作">
                  <button type="button" class="journal-btn" @click="loadCurrentData">刷新</button>
                  <button
                    type="button"
                    class="journal-btn journal-btn--primary"
                    @click="goToChallenges"
                  >
                    去做题
                  </button>
                </div>
              </div>

              <div class="journal-eyebrow journal-eyebrow-soft">Radar Analysis</div>
              <h3 class="workspace-tab-heading__title">能力维度分析</h3>

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
                        <div
                          class="text-base font-semibold text-[var(--journal-ink)] md:text-[1.05rem]"
                        >
                          {{ dim.name }}
                        </div>
                        <div class="mt-1 text-[0.8rem] text-[var(--journal-muted)]">
                          当前维度表现
                        </div>
                      </div>
                      <div class="text-right">
                        <div
                          class="text-[1.9rem] font-semibold tracking-tight text-[var(--journal-ink)] tech-font md:text-[2.1rem]"
                        >
                          {{ dim.value }}
                        </div>
                        <div class="text-xs text-[var(--journal-muted)]">/ 100</div>
                      </div>
                    </article>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section
          id="skill-profile-panel-weakness"
          class="tab-panel skill-section"
          role="tabpanel"
          aria-labelledby="skill-profile-tab-weakness"
          :aria-hidden="activeTab === 'weakness' ? 'false' : 'true'"
          v-show="activeTab === 'weakness'"
        >
          <div class="skill-weak-wrap">
            <div class="journal-eyebrow journal-eyebrow-soft">Weak Points</div>
            <div
              class="mt-3 flex items-center gap-3 text-base font-semibold text-[var(--journal-ink)]"
            >
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
                <div class="mt-2 text-sm font-semibold text-[var(--journal-ink)]">
                  暂时没有明显短板
                </div>
              </div>
            </div>
          </div>
        </section>

        <section
          id="skill-profile-panel-recommendations"
          class="tab-panel skill-section"
          role="tabpanel"
          aria-labelledby="skill-profile-tab-recommendations"
          :aria-hidden="activeTab === 'recommendations' ? 'false' : 'true'"
          v-show="activeTab === 'recommendations'"
        >
          <div class="journal-eyebrow journal-eyebrow-soft">Recommendations</div>
          <h3 class="workspace-tab-heading__title">推荐靶场</h3>
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

          <div
            v-else-if="recommendations.length === 0"
            class="mt-6 text-sm text-[var(--journal-muted)]"
          >
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
                      :class="difficultyClass(item.difficulty)"
                    >
                      {{ difficultyLabel(item.difficulty) }}
                    </span>
                  </div>
                  <p class="mt-1 text-xs leading-5 text-[var(--journal-muted)]">
                    {{ item.reason }}
                  </p>
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
  --journal-shell-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--color-border-default) 70%, transparent);
  --journal-control-border: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --journal-divider: color-mix(in srgb, var(--color-border-default) 64%, transparent);
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --journal-shell-accent: var(--color-primary);
  --journal-shell-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-shell-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-shell-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-shell-hero-radial-strength: 12%;
  --journal-shell-hero-radial-size: 20rem;
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle, var(--color-bg-elevated)) 94%,
    var(--color-bg-base)
  );
  --journal-shell-dark-ink: var(--color-text-primary);
  --journal-shell-dark-muted: var(--color-text-secondary);
  --journal-shell-dark-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-dark-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-shell-dark-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 74%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-radial-strength: 16%;
  --page-top-tabs-gap: 1.2rem;
  --page-top-tabs-margin: 0 -0.5rem 1.5rem;
  --page-top-tabs-padding: 0 0.5rem;
  --page-top-tabs-border: color-mix(in srgb, var(--journal-soft-border) 92%, transparent);
  --page-top-tab-min-height: 3rem;
  --page-top-tab-padding: 0.4rem 0 0.75rem;
  --page-top-tab-font-size: var(--font-size-0-92);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
  --journal-user-button-height: 34px;
  --journal-user-button-radius: 12px;
  --journal-user-button-border: var(--journal-control-border);
  --journal-user-button-background: transparent;
  --journal-user-button-padding: 0.5rem 1rem;
  --journal-user-button-size: 0.875rem;
  --journal-user-button-weight: 500;
  --journal-user-button-color: var(--color-text-primary);
  --journal-user-button-hover-border: color-mix(
    in srgb,
    var(--journal-accent) 52%,
    var(--journal-control-border)
  );
  --journal-user-button-hover-background: transparent;
  --journal-user-button-hover-color: var(--journal-accent-strong);
  --journal-user-button-primary-border: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  --journal-user-button-primary-background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  --journal-user-button-primary-color: var(--journal-accent-strong);
  --journal-user-button-primary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 14%,
    transparent
  );
  --journal-user-tech-font: var(--font-family-mono);
  font-family: var(--font-family-sans);
}

.journal-hero {
  border-color: var(--journal-shell-border);
  border-radius: 16px !important;
  overflow: hidden;
}

.skill-teacher-panel {
  margin-top: var(--workspace-tab-panel-gap-top-tight);
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
  padding: 1rem 1.1rem;
}

.skill-field-label {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.skill-student-select {
  border-radius: 14px;
  border: 1px solid var(--journal-control-border);
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
  padding: 0.7rem 0.95rem;
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
  outline: none;
  transition: border-color 0.2s;
}

.skill-student-select:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 52%, var(--journal-control-border));
}

.skill-student-select:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--journal-accent) 58%, white);
  outline-offset: 2px;
}

.skill-overview-head {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  margin-bottom: 1.75rem;
}

.skill-overview-copy {
  max-width: 42rem;
  color: var(--journal-muted);
}

.skill-overview-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.skill-analysis-stack {
  display: block;
}

.skill-dimension-wrap {
  margin-top: 1.25rem;
}

.skill-recommend-list,
.skill-weak-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
}

.skill-weak-item,
.skill-recommend-item {
  padding: 1rem 1.1rem;
}

.skill-weak-item + .skill-weak-item,
.skill-recommend-item + .skill-recommend-item {
  border-top: 1px solid var(--journal-divider);
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
      color-mix(
        in srgb,
        var(--journal-surface-subtle, var(--color-bg-elevated)) 96%,
        var(--color-bg-base)
      )
    ),
    linear-gradient(135deg, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent);
  border: 1px solid var(--journal-shell-border);
}

.skill-dimension-chart__frame::after {
  inset: 18px;
  clip-path: polygon(25% 6%, 75% 6%, 100% 50%, 75% 94%, 25% 94%, 0 50%);
  border: 1px solid
    color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 78%, transparent);
  background:
    radial-gradient(
      circle at 50% 45%,
      color-mix(in srgb, var(--journal-accent) 12%, transparent),
      transparent 60%
    ),
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
    border-top: 1px solid var(--journal-divider);
  }

  .skill-dimension-legend__item:nth-child(2n) {
    border-left: 1px solid var(--journal-divider);
  }
}

@media (max-width: 639px) {
  .skill-dimension-legend__item + .skill-dimension-legend__item {
    border-top: 1px solid var(--journal-divider);
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
    border-top: 1px solid var(--journal-divider);
  }

  .skill-weak-item:nth-child(2n) {
    border-left: 1px solid var(--journal-divider);
  }
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

.skill-recommend-item:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--journal-accent) 58%, white);
  outline-offset: 2px;
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
  border-color: color-mix(in srgb, var(--journal-muted) 20%, transparent);
}

@media (max-width: 767px) {
  .journal-shell {
    --journal-user-button-height: 36px;
  }
}
</style>
