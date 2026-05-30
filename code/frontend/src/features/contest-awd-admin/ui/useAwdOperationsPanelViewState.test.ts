import { describe, expect, it } from 'vitest'
import { nextTick, ref } from 'vue'

import { computed } from 'vue'
import type { ContestDetailData } from '@/api/contracts'
import { useAwdOperationsPanelViewState } from './useAwdOperationsPanelViewState'
import type { AWDOperationsPanelKey } from './awdOperations.types'
import { isAwdRuntimeStageStatus } from '../model/useAwdContestStateFlags'

function buildContest(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'awd-1',
    title: '2026 AWD 联赛',
    description: '攻防赛',
    mode: 'awd',
    status: 'running',
    starts_at: '2026-03-18T09:00:00.000Z',
    ends_at: '2026-03-18T18:00:00.000Z',
    ...overrides,
  }
}

describe('useAwdOperationsPanelViewState', () => {
  it('应根据赛事状态收口运行态与可见 tab', () => {
    const selectedContest = ref<ContestDetailData | null>(buildContest())
    const runtimeStageReady = computed(() => isAwdRuntimeStageStatus(selectedContest.value?.status))

    const state = useAwdOperationsPanelViewState({
      runtimeStageReady,
      operationPanel: ref<AWDOperationsPanelKey | undefined>(undefined),
      hideContestSelector: ref(false),
      hideOperationTabs: ref(false),
      runtimeContent: ref(undefined),
    })

    expect(state.runtimeStageReady.value).toBe(true)
    expect(state.visibleOperationTabs.value.map((tab) => tab.key)).toEqual([
      'inspector',
      'instances',
    ])

    selectedContest.value = buildContest({ status: 'registering' })

    expect(state.runtimeStageReady.value).toBe(false)
    expect(state.visibleOperationTabs.value.map((tab) => tab.key)).toEqual(['inspector'])
  })

  it('在非受控模式下应允许切换面板，在受控模式下应跟随外部 prop', async () => {
    const operationPanel = ref<AWDOperationsPanelKey | undefined>(undefined)
    const state = useAwdOperationsPanelViewState({
      runtimeStageReady: computed(() => true),
      operationPanel,
      hideContestSelector: ref(false),
      hideOperationTabs: ref(false),
      runtimeContent: ref('all'),
    })

    state.selectPanel('instances')
    expect(state.activePanel.value).toBe('instances')

    operationPanel.value = 'inspector'
    await nextTick()
    expect(state.activePanel.value).toBe('inspector')

    state.selectPanel('instances')
    expect(state.activePanel.value).toBe('inspector')
  })

  it('应根据 runtime content 与 active panel 推导各阶段展示块', () => {
    const state = useAwdOperationsPanelViewState({
      runtimeStageReady: computed(() => true),
      operationPanel: ref<AWDOperationsPanelKey | undefined>('instances'),
      hideContestSelector: ref(true),
      hideOperationTabs: ref(true),
      runtimeContent: ref('instances'),
    })

    expect(state.shouldShowContestSelector.value).toBe(false)
    expect(state.shouldShowOperationTabs.value).toBe(false)
    expect(state.shouldShowRuntimeReadiness.value).toBe(false)
    expect(state.shouldShowRoundInspector.value).toBe(false)
    expect(state.shouldShowInstanceOrchestration.value).toBe(true)
  })
})
