import { describe, expect, it } from 'vitest'

import awdOperationsPanelSource from '../contest/AWDOperationsPanel.vue?raw'

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
  })

  it('AWDOperationsPanel 应将赛事选择器与未开赛运行壳层下沉到独立子组件，而不是继续在父组件内联整段结构', () => {
    expect(awdOperationsPanelSource).toContain('<AWDContestSelectorField')
    expect(awdOperationsPanelSource).toContain('<AWDRuntimePendingState')
    expect(awdOperationsPanelSource).not.toContain('id="awd-runtime-shell-create-round"')
    expect(awdOperationsPanelSource).not.toContain('id="awd-runtime-shell-run-check"')
    expect(awdOperationsPanelSource).not.toContain('id="awd-contest-selector"')
  })
})
