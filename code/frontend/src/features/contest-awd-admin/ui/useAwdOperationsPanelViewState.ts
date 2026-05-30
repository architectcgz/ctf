import { computed, ref, watch, type Ref } from 'vue'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'

import type {
  AWDOperationsPanelKey,
  AWDOperationsRuntimeContent,
  AWDOperationsTabItem,
} from './awdOperations.types'

interface UseAwdOperationsPanelViewStateOptions {
  runtimeStageReady: Readonly<Ref<boolean>>
  operationPanel: Readonly<Ref<AWDOperationsPanelKey | undefined>>
  hideContestSelector: Readonly<Ref<boolean | undefined>>
  hideOperationTabs: Readonly<Ref<boolean | undefined>>
  runtimeContent: Readonly<Ref<AWDOperationsRuntimeContent | undefined>>
}

const operationTabs: readonly AWDOperationsTabItem[] = [
  {
    key: 'inspector',
    label: '轮次态势',
    tabId: 'awd-ops-tab-inspector',
    panelId: 'awd-ops-panel-inspector',
  },
  {
    key: 'instances',
    label: '实例编排',
    tabId: 'awd-ops-tab-instances',
    panelId: 'awd-ops-panel-instances',
  },
] as const

const operationTabOrder = operationTabs.map((tab) => tab.key) as AWDOperationsPanelKey[]

export function useAwdOperationsPanelViewState({
  runtimeStageReady,
  operationPanel,
  hideContestSelector,
  hideOperationTabs,
  runtimeContent,
}: UseAwdOperationsPanelViewStateOptions) {
  const shouldShowContestSelector = computed(() => !hideContestSelector.value)
  const activePanel = ref<AWDOperationsPanelKey>(operationPanel.value ?? 'inspector')
  const visibleOperationTabs = computed(() =>
    runtimeStageReady.value ? operationTabs : operationTabs.filter((tab) => tab.key === 'inspector')
  )
  const shouldShowOperationTabs = computed(() => !hideOperationTabs.value)
  const resolvedRuntimeContent = computed(() => runtimeContent.value ?? 'all')
  const shouldShowRuntimeReadiness = computed(
    () => resolvedRuntimeContent.value === 'all' || resolvedRuntimeContent.value === 'readiness'
  )
  const shouldShowRoundInspector = computed(
    () =>
      activePanel.value === 'inspector' &&
      (resolvedRuntimeContent.value === 'all' ||
        resolvedRuntimeContent.value === 'round-inspector')
  )
  const shouldShowInstanceOrchestration = computed(
    () =>
      activePanel.value === 'instances' &&
      (resolvedRuntimeContent.value === 'all' || resolvedRuntimeContent.value === 'instances')
  )

  watch(operationPanel, (panel) => {
    if (panel) {
      activePanel.value = panel
    }
  })

  function selectPanel(panel: AWDOperationsPanelKey) {
    if (operationPanel.value) {
      return
    }
    activePanel.value = panel
  }

  const { setTabButtonRef, handleTabKeydown } = useTabKeyboardNavigation<AWDOperationsPanelKey>({
    orderedTabs: operationTabOrder,
    selectTab: selectPanel,
  })

  function registerTabButton(key: AWDOperationsPanelKey, element: HTMLButtonElement | null) {
    setTabButtonRef(key, element)
  }

  function handlePanelTabKeydown(event: KeyboardEvent, index: number) {
    handleTabKeydown(event, index)
  }

  return {
    shouldShowContestSelector,
    runtimeStageReady,
    activePanel,
    visibleOperationTabs,
    shouldShowOperationTabs,
    runtimeContent: resolvedRuntimeContent,
    shouldShowRuntimeReadiness,
    shouldShowRoundInspector,
    shouldShowInstanceOrchestration,
    selectPanel,
    registerTabButton,
    handlePanelTabKeydown,
  }
}
