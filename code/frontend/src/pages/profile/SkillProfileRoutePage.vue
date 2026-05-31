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
    :active-tab="activePanel"
    :content-tabs="contentTabs"
    :set-tab-button-ref="setTabButtonRef"
    :handle-tab-keydown="handleTabKeydown"
    :build-challenge-route="buildChallengeRoute"
    @load-current-data="loadCurrentData"
    @select-tab="switchPanel"
    @update-selected-student-id="selectedStudentId = $event"
  />
</template>

<script setup lang="ts">
import { useTabKeyboardNavigation } from '@/shared/lib/keyboard/useTabKeyboardNavigation'
import {
  SkillProfileWorkspaceShell,
  type SkillProfilePanelKey,
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
  activePanel,
  radarIndicators,
  radarValues,
  loadCurrentData,
  switchPanel,
  buildChallengeRoute,
} = useSkillProfilePage()

const contentTabs: Array<{
  key: SkillProfilePanelKey
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

const contentTabOrder = contentTabs.map((tab) => tab.key) as SkillProfilePanelKey[]
const { setTabButtonRef, handleTabKeydown } = useTabKeyboardNavigation<SkillProfilePanelKey>({
  orderedTabs: contentTabOrder,
  selectTab: (tab) => void switchPanel(tab),
})
</script>
