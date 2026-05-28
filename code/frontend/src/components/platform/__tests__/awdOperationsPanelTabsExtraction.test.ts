import { describe, expect, it } from 'vitest'

import awdOperationsPanelSource from '@/features/contest-awd-admin/ui/AWDOperationsPanel.vue?raw'
import awdOperationsDialogHubSource from '@/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue?raw'
import awdOperationsPreRuntimeStageSource from '@/features/contest-awd-admin/ui/AWDOperationsPreRuntimeStage.vue?raw'
import awdOperationsRuntimeStageSource from '@/features/contest-awd-admin/ui/AWDOperationsRuntimeStage.vue?raw'
import awdOperationsTabsSource from '@/features/contest-awd-admin/ui/AWDOperationsTabs.vue?raw'

describe('awd operations panel tabs extraction', () => {
  it('AWDOperationsPanel 应复用 useTabKeyboardNavigation，而不是继续本地维护按钮 ref 与键盘导航状态机', () => {
    expect(awdOperationsPanelSource).toContain(
      "import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'"
    )
    expect(awdOperationsPanelSource).toContain('useTabKeyboardNavigation<AWDOperationsPanelKey>({')
    expect(awdOperationsPanelSource).not.toContain(
      'const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])'
    )
    expect(awdOperationsPanelSource).not.toContain(
      'function setTabButtonRef(index: number, element: HTMLButtonElement | null) {'
    )
    expect(awdOperationsPanelSource).not.toContain('function focusTabByIndex(index: number) {')
    expect(awdOperationsPanelSource).not.toContain(
      'function handlePanelKeydown(event: KeyboardEvent, index: number) {'
    )
    expect(awdOperationsPreRuntimeStageSource).toContain('<AWDOperationsTabs')
    expect(awdOperationsRuntimeStageSource).toContain('<AWDOperationsTabs')
    expect(awdOperationsTabsSource).toContain('class="studio-ops-tabs"')
  })

  it('AWDOperationsPanel 应将赛事选择器与未开赛运行壳层下沉到独立子组件，而不是继续在父组件内联整段结构', () => {
    expect(awdOperationsPanelSource).toContain('<AWDContestSelectorField')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsPreRuntimeStage')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsRuntimeStage')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsDialogHub')
    expect(awdOperationsPreRuntimeStageSource).toContain('name="pending"')
    expect(awdOperationsRuntimeStageSource).toContain('name="inspector"')
    expect(awdOperationsDialogHubSource).toContain('<AWDRoundCreateDialog')
    expect(awdOperationsDialogHubSource).toContain('<AWDReadinessOverrideDialog')
    expect(awdOperationsPanelSource).not.toContain('id="awd-runtime-shell-create-round"')
    expect(awdOperationsPanelSource).not.toContain('id="awd-runtime-shell-run-check"')
    expect(awdOperationsPanelSource).not.toContain('id="awd-contest-selector"')
    expect(awdOperationsPanelSource).not.toContain('class="section-header"')
    expect(awdOperationsPanelSource).not.toContain('class="studio-ops-tabs"')
    expect(awdOperationsPanelSource).not.toContain('<AWDRoundCreateDialog')
    expect(awdOperationsPanelSource).not.toContain('<AWDServiceCheckDialog')
    expect(awdOperationsPanelSource).not.toContain('<AWDAttackLogDialog')
    expect(awdOperationsPanelSource).not.toContain('<AWDReadinessOverrideDialog')
  })
})
