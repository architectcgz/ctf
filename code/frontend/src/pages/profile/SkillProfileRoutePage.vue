<template>
  <SkillProfileWorkspaceShell
    :is-teacher="isTeacher"
    :selected-student-id="selectedStudentId"
    :students="students"
    :loading="loading"
    :error="error"
    :skill-profile="skillProfile"
    :loading-recommendations="loadingRecommendations"
    :recommendations="recommendations"
    :weak-dimensions="weakDimensions"
    :challenges-route="challengesRoute"
    :radar-indicators="radarIndicators"
    :radar-values="radarValues"
    :active-tab="activeTab"
    :content-tabs="contentTabs"
    :set-tab-button-ref="setTabButtonRef"
    :handle-tab-keydown="handleTabKeydown"
    :build-challenge-route="buildChallengeRoute"
    @load-current-data="loadCurrentData"
    @select-tab="selectTab"
    @update-selected-student-id="selectedStudentId = $event"
  />
</template>

<script setup lang="ts">
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'
import {
  SkillProfileWorkspaceShell,
  useSkillProfilePage,
} from '@/features/skill-profile'

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
  challengesRoute,
  radarIndicators,
  radarValues,
  loadCurrentData,
  buildChallengeRoute,
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
    label: '六维分布分析',
    buttonId: 'skill-profile-tab-analysis',
    panelId: 'skill-profile-panel-analysis',
  },
  {
    key: 'weakness',
    label: '薄弱维度提示',
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

const contentTabOrder = contentTabs.map((tab) => tab.key) as SkillProfileTabKey[]
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } =
  useUrlSyncedTabs<SkillProfileTabKey>({
    orderedTabs: contentTabOrder,
    defaultTab: 'analysis',
  })
</script>
