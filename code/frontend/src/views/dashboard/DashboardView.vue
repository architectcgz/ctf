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
import StudentOverviewVariantSwitcher from '@/components/dashboard/student/StudentOverviewVariantSwitcher.vue'
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
</script>

<template>
  <div class="dashboard-view space-y-6">
    <div
      v-if="error"
      class="rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 px-5 py-4 text-sm text-[var(--color-danger)]"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="loadDashboard">重试</button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div
        v-for="index in 4"
        :key="index"
        class="h-32 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]"
      />
    </div>

    <template v-else-if="progress">
      <section class="dashboard-tab-rail rounded-2xl border px-4 py-3">
        <div class="dashboard-tab-list" role="tablist" aria-label="学生仪表盘视图切换">
          <button
            v-for="tab in panelTabs"
            :id="tab.tabId"
            :key="tab.tabId"
            type="button"
            role="tab"
            class="dashboard-tab"
            :class="{ 'dashboard-tab--active': isPanelActive(tab.key) }"
            :aria-selected="isPanelActive(tab.key)"
            :aria-controls="tab.panelId"
            @click="switchPanel(tab.key)"
          >
            {{ tab.label }}
          </button>
        </div>
      </section>

      <div
        v-if="isOverview"
        id="dashboard-panel-overview"
        role="tabpanel"
        aria-labelledby="dashboard-tab-overview"
      >
        <StudentOverviewVariantSwitcher
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
      </div>

      <div
        v-else-if="activePanel === 'recommendation'"
        id="dashboard-panel-recommendation"
        role="tabpanel"
        aria-labelledby="dashboard-tab-recommendation"
      >
        <StudentRecommendationPage
          :weak-dimensions="weakDimensions"
          :recommendations="recommendations"
          @open-challenge="openChallenge"
          @open-challenges="openChallenges"
          @open-skill-profile="openSkillProfile"
        />
      </div>

      <div
        v-else-if="activePanel === 'category'"
        id="dashboard-panel-category"
        role="tabpanel"
        aria-labelledby="dashboard-tab-category"
      >
        <StudentCategoryProgressPage
          :category-stats="categoryStats"
          :completion-rate="completionRate"
          @open-challenges="openChallenges"
          @open-skill-profile="openSkillProfile"
        />
      </div>

      <div
        v-else-if="activePanel === 'timeline'"
        id="dashboard-panel-timeline"
        role="tabpanel"
        aria-labelledby="dashboard-tab-timeline"
      >
        <StudentTimelinePage :timeline="timeline" />
      </div>

      <div
        v-else
        id="dashboard-panel-difficulty"
        role="tabpanel"
        aria-labelledby="dashboard-tab-difficulty"
      >
        <StudentDifficultyPage :difficulty-stats="difficultyStats" />
      </div>
    </template>
  </div>
</template>

<style scoped>
.dashboard-tab-rail {
  border-color: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
}

.dashboard-tab-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.dashboard-tab {
  min-height: 34px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 86%, transparent);
  border-radius: 10px;
  padding: 0.375rem 0.875rem;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  background: color-mix(in srgb, var(--color-bg-surface) 94%, transparent);
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease;
}

.dashboard-tab:hover {
  border-color: color-mix(in srgb, var(--color-primary) 52%, var(--color-border-default));
  color: var(--color-text-primary);
}

.dashboard-tab--active {
  border-color: color-mix(in srgb, var(--color-primary) 56%, transparent);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  color: var(--color-primary-hover);
}

.dashboard-tab:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 60%, white);
  outline-offset: 2px;
}

@media (max-width: 767px) {
  .dashboard-tab {
    min-height: 36px;
  }
}
</style>
