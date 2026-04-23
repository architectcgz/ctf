import { describe, expect, it } from 'vitest'

import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'

describe('class management tabs adoption', () => {
  it('ClassStudentsPage 应复用 useUrlSyncedTabs，而不是继续本地维护 URL 同步与键盘导航状态机', () => {
    expect(classStudentsPageSource).toContain(
      "import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'"
    )
    expect(classStudentsPageSource).toContain('useUrlSyncedTabs<WorkspaceTab>(')
    expect(classStudentsPageSource).not.toContain(
      'function resolveTabFromLocation(): WorkspaceTab {'
    )
    expect(classStudentsPageSource).not.toContain(
      'function syncPanelToLocation(tab: WorkspaceTab): void {'
    )
    expect(classStudentsPageSource).not.toContain(
      'const tabButtonRefs: Partial<Record<WorkspaceTab, HTMLButtonElement | null>> = {}'
    )
    expect(classStudentsPageSource).not.toContain('function focusTab(tab: WorkspaceTab): void {')
    expect(classStudentsPageSource).not.toContain(
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )
  })
})
