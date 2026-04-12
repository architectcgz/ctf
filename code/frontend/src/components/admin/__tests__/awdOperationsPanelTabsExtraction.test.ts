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
})
