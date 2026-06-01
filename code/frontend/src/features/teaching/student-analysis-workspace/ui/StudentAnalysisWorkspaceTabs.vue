<template>
  <nav class="workspace-tabbar top-tabs" role="tablist" aria-label="学员分析标签页">
    <button
      v-for="(tab, index) in studentAnalysisWorkspaceTabs"
      :id="tab.buttonId"
      :key="tab.key"
      :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
      class="workspace-tab top-tab"
      :class="{ active: activeWorkspaceTab === tab.key }"
      type="button"
      role="tab"
      :tabindex="activeWorkspaceTab === tab.key ? 0 : -1"
      :aria-selected="activeWorkspaceTab === tab.key ? 'true' : 'false'"
      :aria-controls="tab.panelId"
      @click="emit('selectWorkspaceTab', tab.key)"
      @keydown="handleTabKeydown($event, index)"
    >
      {{ tab.label }}
    </button>
  </nav>
</template>

<script setup lang="ts">
import { useTabKeyboardNavigation } from '@/shared/lib/keyboard/useTabKeyboardNavigation'
import type { StudentAnalysisWorkspaceTab } from '../model/useStudentAnalysisPage'
import { studentAnalysisWorkspaceTabs } from './studentAnalysisWorkspaceTabs'

defineProps<{
  activeWorkspaceTab: StudentAnalysisWorkspaceTab
}>()

const emit = defineEmits<{
  selectWorkspaceTab: [tab: StudentAnalysisWorkspaceTab]
}>()

const workspaceTabOrder = studentAnalysisWorkspaceTabs.map(
  (tab) => tab.key
) as StudentAnalysisWorkspaceTab[]
const { setTabButtonRef, handleTabKeydown } =
  useTabKeyboardNavigation<StudentAnalysisWorkspaceTab>({
    orderedTabs: workspaceTabOrder,
    selectTab: (tab) => emit('selectWorkspaceTab', tab),
  })
</script>
