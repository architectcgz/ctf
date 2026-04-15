import { describe, expect, it } from 'vitest'

import classManagementPageSource from '@/components/teacher/class-management/ClassManagementPage.vue?raw'

describe('class management tabs adoption', () => {
  it('ClassManagementPage 应复用 useUrlSyncedTabs，而不是继续本地维护 URL 同步与键盘导航状态机', () => {
    expect(classManagementPageSource).toContain(
      "import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'"
    )
    expect(classManagementPageSource).toContain('useUrlSyncedTabs<WorkspaceTab>(')
    expect(classManagementPageSource).not.toContain(
      'function resolveTabFromLocation(): WorkspaceTab {'
    )
    expect(classManagementPageSource).not.toContain(
      'function syncPanelToLocation(tab: WorkspaceTab): void {'
    )
    expect(classManagementPageSource).not.toContain(
      'const tabButtonRefs: Partial<Record<WorkspaceTab, HTMLButtonElement | null>> = {}'
    )
    expect(classManagementPageSource).not.toContain('function focusTab(tab: WorkspaceTab): void {')
    expect(classManagementPageSource).not.toContain(
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )
  })
})
