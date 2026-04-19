import { describe, expect, it } from 'vitest'

import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'

describe('contest orchestration tabs adoption', () => {
  it('ContestOrchestrationPage 应复用 useUrlSyncedTabs，而不是继续本地维护 URL 同步与键盘导航状态机', () => {
    expect(contestOrchestrationSource).toContain(
      "import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'"
    )
    expect(contestOrchestrationSource).toContain('useUrlSyncedTabs<ContestPanelKey>(')
    expect(contestOrchestrationSource).not.toContain(
      'function resolvePanelFromLocation(): ContestPanelKey {'
    )
    expect(contestOrchestrationSource).not.toContain(
      'function syncPanelToLocation(panelKey: ContestPanelKey): void {'
    )
    expect(contestOrchestrationSource).not.toContain(
      'const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])'
    )
    expect(contestOrchestrationSource).not.toContain(
      'function focusTabByIndex(index: number): void {'
    )
    expect(contestOrchestrationSource).not.toContain(
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )
  })
})
