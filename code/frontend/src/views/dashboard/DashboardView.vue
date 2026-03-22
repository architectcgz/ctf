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
import PageHeader from '@/components/common/PageHeader.vue'
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
type DashboardVariantKey = '1' | '2' | '3'

const validPanelKeys = new Set<DashboardPanelKey>([
  'category',
  'recommendation',
  'timeline',
  'difficulty',
])
const validVariantKeys = new Set<DashboardVariantKey>(['1', '2', '3'])

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
const activeVariant = computed<DashboardVariantKey>(() => {
  const variant = route.params.variant
  if (typeof variant === 'string' && validVariantKeys.has(variant as DashboardVariantKey)) {
    return variant as DashboardVariantKey
  }
  return '2'
})
const isOverview = computed(() => activePanel.value === null)
const panelCopyMap: Record<DashboardPanelKey, { title: string; description: string }> = {
  recommendation: {
    title: '训练建议',
    description: '基于当前画像推荐更适合补短板的题目，帮助你用更短路径补齐能力缺口。',
  },
  category: {
    title: '分类进度',
    description: '查看不同题型方向的完成比例，识别当前训练覆盖是否均衡。',
  },
  timeline: {
    title: '近期动态',
    description: '回看最近实例与提交记录，快速理解自己的训练节奏。',
  },
  difficulty: {
    title: '难度分布',
    description: '观察不同难度层级的完成情况，避免训练结构长期失衡。',
  },
}
const panelHeader = computed(() => {
  if (!activePanel.value) return null
  return panelCopyMap[activePanel.value]
})

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
</script>

<template>
  <div class="dashboard-view space-y-6">
    <PageHeader
      class="dashboard-view__header"
      eyebrow="Student Workspace"
      :title="
        isOverview
          ? `${displayName} 的训练仪表盘`
          : panelHeader?.title || `${displayName} 的训练仪表盘`
      "
      :description="
        isOverview
          ? '汇总当前得分、解题进度、近期训练动态与推荐靶场，帮助你快速判断下一步训练重点。'
          : panelHeader?.description || ''
      "
    >
      <template v-if="isOverview">
        <ElButton plain @click="router.push({ name: 'SkillProfile' })">能力画像</ElButton>
        <ElButton type="primary" @click="router.push({ name: 'Challenges' })">继续训练</ElButton>
      </template>
    </PageHeader>

    <div
      v-if="error"
      class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600"
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
      <StudentOverviewVariantSwitcher
        v-if="isOverview"
        :variant="activeVariant"
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
        v-else-if="activePanel === 'recommendation'"
        :weak-dimensions="weakDimensions"
        :recommendations="recommendations"
        @open-challenge="openChallenge"
        @open-challenges="openChallenges"
        @open-skill-profile="openSkillProfile"
      />
      <StudentCategoryProgressPage
        v-else-if="activePanel === 'category'"
        :category-stats="categoryStats"
        :completion-rate="completionRate"
        @open-challenges="openChallenges"
        @open-skill-profile="openSkillProfile"
      />
      <StudentTimelinePage v-else-if="activePanel === 'timeline'" :timeline="timeline" />
      <StudentDifficultyPage v-else :difficulty-stats="difficultyStats" />
    </template>
  </div>
</template>

<style scoped>
.dashboard-view :deep(.journal-hero),
.dashboard-view :deep(.journal-brief),
.dashboard-view :deep(.journal-panel),
.dashboard-view :deep(.journal-metric),
.dashboard-view :deep(.journal-note),
.dashboard-view :deep(.journal-rec-item),
.dashboard-view :deep(.journal-log),
.dashboard-view :deep(.command-hero),
.dashboard-view :deep(.command-highlight),
.dashboard-view :deep(.command-panel),
.dashboard-view :deep(.command-metric),
.dashboard-view :deep(.command-status-card),
.dashboard-view :deep(.command-rec-item),
.dashboard-view :deep(.signal-hero),
.dashboard-view :deep(.signal-panel),
.dashboard-view :deep(.signal-stat),
.dashboard-view :deep(.signal-rec-item),
.dashboard-view :deep(.signal-log),
.dashboard-view :deep(.signal-summary) {
  border-radius: 16px !important;
  overflow: hidden;
}

:global([data-theme='light']) .dashboard-view :deep(.journal-hero),
:global([data-theme='light']) .dashboard-view :deep(.journal-brief),
:global([data-theme='light']) .dashboard-view :deep(.journal-panel),
:global([data-theme='light']) .dashboard-view :deep(.journal-metric),
:global([data-theme='light']) .dashboard-view :deep(.journal-note),
:global([data-theme='light']) .dashboard-view :deep(.journal-rec-item),
:global([data-theme='light']) .dashboard-view :deep(.journal-log) {
  box-shadow: 0 10px 24px rgba(148, 163, 184, 0.12);
}

@media (max-width: 767px) {
  .dashboard-view :deep(.journal-hero),
  .dashboard-view :deep(.journal-brief),
  .dashboard-view :deep(.journal-panel),
  .dashboard-view :deep(.journal-metric),
  .dashboard-view :deep(.journal-note),
  .dashboard-view :deep(.journal-rec-item),
  .dashboard-view :deep(.journal-log),
  .dashboard-view :deep(.command-hero),
  .dashboard-view :deep(.command-highlight),
  .dashboard-view :deep(.command-panel),
  .dashboard-view :deep(.command-metric),
  .dashboard-view :deep(.command-status-card),
  .dashboard-view :deep(.command-rec-item),
  .dashboard-view :deep(.signal-hero),
  .dashboard-view :deep(.signal-panel),
  .dashboard-view :deep(.signal-stat),
  .dashboard-view :deep(.signal-rec-item),
  .dashboard-view :deep(.signal-log),
  .dashboard-view :deep(.signal-summary) {
    border-radius: 12px !important;
  }
}
</style>
