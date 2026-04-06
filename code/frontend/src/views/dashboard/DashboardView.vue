<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FileChartColumnIncreasing, Rocket, ShieldAlert } from 'lucide-vue-next'

import { getMyProgress, getMyTimeline, getRecommendations, getSkillProfile } from '@/api/assessment'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TimelineEvent,
} from '@/api/contracts'
import StudentCategoryProgressPage from '@/components/dashboard/student/StudentCategoryProgressPage.vue'
import StudentDifficultyPage from '@/components/dashboard/student/StudentDifficultyPage.vue'
import StudentOverviewStyleEditorial from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue'
import StudentRecommendationPage from '@/components/dashboard/student/StudentRecommendationPage.vue'
import StudentTimelinePage from '@/components/dashboard/student/StudentTimelinePage.vue'
import { useAuthStore } from '@/stores/auth'
import { getWeakDimensions } from '@/utils/skillProfile'

import type { DashboardHighlightItem } from '@/components/dashboard/student/types'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const loading = ref(false)
const error = ref<string | null>(null)
const progress = ref<MyProgressData | null>(null)
const timeline = ref<TimelineEvent[]>([])
const recommendations = ref<RecommendationItem[]>([])
const skillProfile = ref<SkillProfileData | null>(null)

type DashboardPanelKey = 'category' | 'recommendation' | 'timeline' | 'difficulty'
const validPanelKeys = new Set<DashboardPanelKey>([
  'category',
  'recommendation',
  'timeline',
  'difficulty',
])

const displayName = computed(() => authStore.user?.name || authStore.user?.username || '选手')
const weakDimensions = computed(() => getWeakDimensions(skillProfile.value).slice(0, 3))
const recommendationCount = computed(() => recommendations.value.length)
const timelineCount = computed(() => timeline.value.length)
const categoryStats = computed(() => progress.value?.category_stats ?? [])
const difficultyStats = computed(() => progress.value?.difficulty_stats ?? [])
const completionRate = computed(() => {
  const solved = progress.value?.total_solved ?? 0
  const total = progress.value?.category_stats?.reduce((sum, item) => sum + item.total, 0) ?? 0
  if (!total) return 0
  return Math.round((solved / total) * 100)
})
const highlightItems = computed<DashboardHighlightItem[]>(() => [
  {
    label: '训练完成率',
    value: `${completionRate.value}%`,
    description: '按当前分类题量计算的整体覆盖率',
    icon: Rocket,
  },
  {
    label: '推荐任务',
    value: `${recommendationCount.value} 项`,
    description:
      weakDimensions.value.length > 0
        ? `优先补强 ${weakDimensions.value.join(' / ')}`
        : '当前没有明显短板',
    icon: ShieldAlert,
  },
  {
    label: '近期动态',
    value: `${timelineCount.value} 条`,
    description: '最近实例和提交记录的浓缩视图',
    icon: FileChartColumnIncreasing,
  },
])
const activePanel = computed<DashboardPanelKey | null>(() => {
  const panel = route.query.panel
  if (typeof panel === 'string' && validPanelKeys.has(panel as DashboardPanelKey)) {
    return panel as DashboardPanelKey
  }
  return null
})
const isOverview = computed(() => activePanel.value === null)
const panelTabs: Array<{ key: DashboardPanelKey | null; label: string; panelId: string; tabId: string }> =
  [
    { key: null, label: '总览', panelId: 'dashboard-panel-overview', tabId: 'dashboard-tab-overview' },
    {
      key: 'recommendation',
      label: '训练建议',
      panelId: 'dashboard-panel-recommendation',
      tabId: 'dashboard-tab-recommendation',
    },
    { key: 'category', label: '分类进度', panelId: 'dashboard-panel-category', tabId: 'dashboard-tab-category' },
    { key: 'timeline', label: '近期动态', panelId: 'dashboard-panel-timeline', tabId: 'dashboard-tab-timeline' },
    { key: 'difficulty', label: '难度分布', panelId: 'dashboard-panel-difficulty', tabId: 'dashboard-tab-difficulty' },
  ]

async function loadDashboard(): Promise<void> {
  const role = authStore.user?.role
  if (role === 'teacher') {
    await router.replace({ name: 'TeacherDashboard' })
    return
  }
  if (role === 'admin') {
    await router.replace({ name: 'AdminDashboard' })
    return
  }

  loading.value = true
  error.value = null
  try {
    const [progressPayload, timelinePayload, recommendationPayload, profilePayload] =
      await Promise.all([getMyProgress(), getMyTimeline(), getRecommendations(), getSkillProfile()])

    progress.value = progressPayload
    timeline.value = timelinePayload.slice(0, 6)
    recommendations.value = recommendationPayload.slice(0, 4)
    skillProfile.value = profilePayload
  } catch (err) {
    console.error('加载学生仪表盘失败:', err)
    error.value = '加载仪表盘失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})

function openChallenges(): void {
  router.push({ name: 'Challenges' })
}

function openSkillProfile(): void {
  router.push({ name: 'SkillProfile' })
}

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

function isPanelActive(panelKey: DashboardPanelKey | null): boolean {
  return panelKey === null ? isOverview.value : activePanel.value === panelKey
}

function switchPanel(panelKey: DashboardPanelKey | null): void {
  const nextQuery = { ...route.query }
  if (panelKey === null) {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panelKey
  }
  void router.replace({ query: nextQuery })
}

function focusTabByIndex(index: number): void {
  const safeIndex = Math.max(0, Math.min(index, panelTabs.length - 1))
  const targetTab = panelTabs[safeIndex]
  if (!targetTab) return
  document.getElementById(targetTab.tabId)?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft' && event.key !== 'Home' && event.key !== 'End') {
    return
  }

  event.preventDefault()

  if (event.key === 'Home') {
    switchPanel(panelTabs[0].key)
    focusTabByIndex(0)
    return
  }

  if (event.key === 'End') {
    const endIndex = panelTabs.length - 1
    switchPanel(panelTabs[endIndex].key)
    focusTabByIndex(endIndex)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + panelTabs.length) % panelTabs.length
  switchPanel(panelTabs[nextIndex].key)
  focusTabByIndex(nextIndex)
}
</script>

<template>
  <section class="workspace-shell">
    <nav class="top-tabs" role="tablist" aria-label="学生仪表盘视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.tabId"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: isPanelActive(tab.key) }"
        :aria-selected="isPanelActive(tab.key) ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="isPanelActive(tab.key) ? 0 : -1"
        @click="switchPanel(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <div class="workspace-grid">
      <main class="content-pane">
        <div
          v-if="error"
          class="workspace-alert"
          role="alert"
          aria-live="polite"
        >
          {{ error }}
          <button type="button" class="workspace-alert-action" @click="loadDashboard">重试</button>
        </div>

        <div v-if="loading" class="dashboard-loading-grid">
          <div
            v-for="index in 4"
            :key="index"
            class="dashboard-loading-item"
          />
        </div>

        <template v-else-if="progress">
          <StudentOverviewStyleEditorial
            id="dashboard-panel-overview"
            class="tab-panel"
            :class="{ active: isOverview }"
            role="tabpanel"
            aria-labelledby="dashboard-tab-overview"
            :aria-hidden="isOverview ? 'false' : 'true'"
            v-show="isOverview"
            embedded
            :display-name="displayName"
            :class-name="authStore.user?.class_name"
            :progress="progress"
            :completion-rate="completionRate"
            :highlight-items="highlightItems"
            :recommendations="recommendations"
            :timeline="timeline"
            :weak-dimensions="weakDimensions"
            :skill-dimensions="skillProfile?.dimensions ?? []"
            @open-challenge="openChallenge"
            @open-challenges="openChallenges"
            @open-skill-profile="openSkillProfile"
          />

          <StudentRecommendationPage
            id="dashboard-panel-recommendation"
            class="tab-panel"
            :class="{ active: activePanel === 'recommendation' }"
            role="tabpanel"
            aria-labelledby="dashboard-tab-recommendation"
            :aria-hidden="activePanel === 'recommendation' ? 'false' : 'true'"
            v-show="activePanel === 'recommendation'"
            embedded
            :weak-dimensions="weakDimensions"
            :recommendations="recommendations"
            @open-challenge="openChallenge"
            @open-challenges="openChallenges"
            @open-skill-profile="openSkillProfile"
          />

          <StudentCategoryProgressPage
            id="dashboard-panel-category"
            class="tab-panel"
            :class="{ active: activePanel === 'category' }"
            role="tabpanel"
            aria-labelledby="dashboard-tab-category"
            :aria-hidden="activePanel === 'category' ? 'false' : 'true'"
            v-show="activePanel === 'category'"
            embedded
            :category-stats="categoryStats"
            :completion-rate="completionRate"
            @open-challenges="openChallenges"
            @open-skill-profile="openSkillProfile"
          />

          <StudentTimelinePage
            id="dashboard-panel-timeline"
            class="tab-panel"
            :class="{ active: activePanel === 'timeline' }"
            role="tabpanel"
            aria-labelledby="dashboard-tab-timeline"
            :aria-hidden="activePanel === 'timeline' ? 'false' : 'true'"
            v-show="activePanel === 'timeline'"
            embedded
            :timeline="timeline"
          />

          <StudentDifficultyPage
            id="dashboard-panel-difficulty"
            class="tab-panel"
            :class="{ active: activePanel === 'difficulty' }"
            role="tabpanel"
            aria-labelledby="dashboard-tab-difficulty"
            :aria-hidden="activePanel === 'difficulty' ? 'false' : 'true'"
            v-show="activePanel === 'difficulty'"
            embedded
            :difficulty-stats="difficultyStats"
          />
        </template>
      </main>
    </div>
  </section>
</template>

<style scoped>
.workspace-shell {
  --journal-ink: var(--color-text-primary);
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --workspace-shell: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-danger: var(--color-danger);
  --workspace-shadow-shell: 0 24px 84px color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
  --workspace-font-sans:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  min-height: 100%;
  flex: 1 1 auto;
  border: 1px solid var(--workspace-line-soft);
  border-radius: 28px;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--workspace-brand) 6%, transparent), transparent 26rem),
    linear-gradient(180deg, color-mix(in srgb, var(--workspace-shell) 96%, var(--workspace-page)), var(--workspace-shell));
  box-shadow: var(--workspace-shadow-shell);
  overflow: clip;
  font-family: var(--workspace-font-sans);
  color: var(--journal-ink);
}

.top-tabs {
  display: flex;
  gap: 28px;
  padding: 0 28px;
  margin-top: 10px;
  border-bottom: 1px solid var(--workspace-line-soft);
  overflow-x: auto;
  scrollbar-width: none;
}

.top-tabs::-webkit-scrollbar {
  display: none;
}

.top-tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  min-height: 52px;
  padding: 10px 0 13px;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--workspace-faint);
  font: 600 15px/1 var(--workspace-font-sans);
  white-space: nowrap;
  cursor: pointer;
  transition:
    color 160ms ease,
    border-color 160ms ease;
}

.top-tab:hover,
.top-tab.active,
.top-tab:focus-visible {
  color: var(--workspace-brand-ink);
  border-bottom-color: var(--workspace-brand);
  outline: none;
}

.workspace-grid {
  display: grid;
  grid-template-columns: 1fr;
}

.content-pane {
  min-width: 0;
  min-height: 0;
  padding: 28px;
}

.tab-panel {
  display: none;
  min-height: 0;
}

.tab-panel.active {
  display: block;
  animation: tabPanelIn 180ms ease both;
}

.workspace-alert {
  margin-bottom: 18px;
  padding: 16px 18px;
  border: 1px solid color-mix(in srgb, var(--workspace-danger) 24%, var(--workspace-line-soft));
  border-radius: 18px;
  background: color-mix(in srgb, var(--workspace-danger) 6%, transparent);
  font-size: 13px;
  line-height: 1.7;
  color: var(--journal-ink);
}

.workspace-alert-action {
  margin-left: 10px;
  border: 0;
  background: transparent;
  font-weight: 600;
  text-decoration: underline;
  color: inherit;
  cursor: pointer;
}

.dashboard-loading-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.dashboard-loading-item {
  height: 7.5rem;
  border-radius: 18px;
  background: var(--journal-track);
  animation: dashboardPulse 1.1s ease-in-out infinite;
}

@keyframes tabPanelIn {
  from {
    opacity: 0;
    transform: translateY(3px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes dashboardPulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 860px) {
  .top-tabs {
    gap: 18px;
    padding: 0 18px;
  }

  .content-pane {
    padding: 18px;
  }
}

@media (max-width: 640px) {
  .dashboard-loading-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
