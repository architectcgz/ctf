<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
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
import StudentOverviewPage from '@/components/dashboard/student/StudentOverviewPage.vue'
import StudentRecommendationPage from '@/components/dashboard/student/StudentRecommendationPage.vue'
import StudentTimelinePage from '@/components/dashboard/student/StudentTimelinePage.vue'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
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

type DashboardPanelKey = 'overview' | 'category' | 'recommendation' | 'timeline' | 'difficulty'

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
const panelTabs: Array<{
  key: DashboardPanelKey
  label: string
  panelId: string
  tabId: string
}> = [
  {
    key: 'overview',
    label: '训练总览',
    panelId: 'dashboard-panel-overview',
    tabId: 'dashboard-tab-overview',
  },
  {
    key: 'recommendation',
    label: '训练队列',
    panelId: 'dashboard-panel-recommendation',
    tabId: 'dashboard-tab-recommendation',
  },
  {
    key: 'category',
    label: '分类补强',
    panelId: 'dashboard-panel-category',
    tabId: 'dashboard-tab-category',
  },
  {
    key: 'timeline',
    label: '训练记录',
    panelId: 'dashboard-panel-timeline',
    tabId: 'dashboard-tab-timeline',
  },
  {
    key: 'difficulty',
    label: '强度推进',
    panelId: 'dashboard-panel-difficulty',
    tabId: 'dashboard-tab-difficulty',
  },
]

const panelTabOrder = panelTabs.map((tab) => tab.key) as DashboardPanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<DashboardPanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'overview',
})

async function loadDashboard(): Promise<void> {
  const role = authStore.user?.role
  if (role === 'teacher') {
    await router.replace({ name: 'TeacherDashboard' })
    return
  }
  if (role === 'admin') {
    await router.replace({ name: 'PlatformOverview' })
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

function openCategoryChallenges(category: string): void {
  router.push({ name: 'Challenges', query: { category } })
}

function openDifficultyChallenges(difficulty: string): void {
  router.push({ name: 'Challenges', query: { difficulty } })
}

function openSkillProfile(): void {
  router.push({ name: 'SkillProfile' })
}

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

function resolveDashboardPanelComponent(panelKey: DashboardPanelKey): Component {
  switch (panelKey) {
    case 'overview':
      return StudentOverviewPage
    case 'recommendation':
      return StudentRecommendationPage
    case 'category':
      return StudentCategoryProgressPage
    case 'timeline':
      return StudentTimelinePage
    case 'difficulty':
      return StudentDifficultyPage
  }
}

function resolveDashboardPanelBindings(panelKey: DashboardPanelKey): Record<string, unknown> {
  switch (panelKey) {
    case 'overview':
      return {
        embedded: true,
        displayName: displayName.value,
        className: authStore.user?.class_name,
        progress: progress.value,
        completionRate: completionRate.value,
        highlightItems: highlightItems.value,
        recommendations: recommendations.value,
        timeline: timeline.value,
        weakDimensions: weakDimensions.value,
        skillDimensions: skillProfile.value?.dimensions ?? [],
        onOpenChallenge: openChallenge,
        onOpenChallenges: openChallenges,
        onOpenSkillProfile: openSkillProfile,
      }
    case 'recommendation':
      return {
        embedded: true,
        weakDimensions: weakDimensions.value,
        recommendations: recommendations.value,
        onOpenChallenge: openChallenge,
        onOpenChallenges: openChallenges,
        onOpenSkillProfile: openSkillProfile,
      }
    case 'category':
      return {
        embedded: true,
        categoryStats: categoryStats.value,
        completionRate: completionRate.value,
        onOpenChallenges: openChallenges,
        onOpenCategoryChallenges: openCategoryChallenges,
        onOpenSkillProfile: openSkillProfile,
      }
    case 'timeline':
      return {
        embedded: true,
        timeline: timeline.value,
      }
    case 'difficulty':
      return {
        embedded: true,
        difficultyStats: difficultyStats.value,
        onOpenChallenges: openChallenges,
        onOpenDifficultyChallenges: openDifficultyChallenges,
      }
  }
}
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <nav
      class="top-tabs"
      role="tablist"
      aria-label="学生仪表盘视图切换"
    >
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.tabId"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: activePanel === tab.key }"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="activePanel === tab.key ? 0 : -1"
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
          <button
            type="button"
            class="workspace-alert-action"
            @click="loadDashboard"
          >
            重试
          </button>
        </div>

        <div
          v-if="loading"
          class="dashboard-loading-grid"
        >
          <div
            v-for="index in 4"
            :key="index"
            class="dashboard-loading-item"
          />
        </div>

        <template v-else-if="progress">
          <component
            :is="resolveDashboardPanelComponent(tab.key)"
            v-for="tab in panelTabs"
            v-show="activePanel === tab.key"
            :id="tab.panelId"
            :key="tab.panelId"
            class="tab-panel"
            :class="{ active: activePanel === tab.key }"
            role="tabpanel"
            :aria-labelledby="tab.tabId"
            :aria-hidden="activePanel === tab.key ? 'false' : 'true'"
            v-bind="resolveDashboardPanelBindings(tab.key)"
          />
        </template>
      </main>
    </div>
  </section>
</template>

<style scoped>
.workspace-shell {
  --workspace-line-soft: var(--journal-border);
  --workspace-faint: var(--journal-muted);
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: var(--journal-accent-strong);
  --workspace-brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --workspace-page: var(--color-bg-base);
  --workspace-shell-bg: var(--journal-surface);
  --workspace-danger: var(--color-danger);
  --workspace-shadow-shell: var(--journal-shell-hero-shadow, 0 22px 50px var(--color-shadow-soft));
  --workspace-font-sans: var(--font-family-sans);
  --journal-track: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 94%,
    var(--color-bg-base)
  );
  flex: 1 1 auto;
}

.content-pane {
  min-height: 0;
}

.tab-panel {
  min-height: 0;
}

.workspace-alert {
  margin-bottom: 18px;
  padding: 16px 18px;
  border: 1px solid color-mix(in srgb, var(--workspace-danger) 24%, var(--workspace-line-soft));
  border-radius: 18px;
  background: color-mix(in srgb, var(--workspace-danger) 6%, transparent);
  font-size: var(--font-size-13);
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
