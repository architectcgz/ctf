import { describe, expect, it } from 'vitest'

import skillProfileSource from '../SkillProfile.vue?raw'

describe('skill profile tabs adoption', () => {
  it('SkillProfile 应复用 useUrlSyncedTabs，而不是继续本地维护 tab 按钮 ref 与键盘导航状态机', () => {
    expect(skillProfileSource).toContain(
      "import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'"
    )
    expect(skillProfileSource).toContain('useUrlSyncedTabs<SkillProfileTabKey>(')
    expect(skillProfileSource).not.toContain(
      'const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])'
    )
    expect(skillProfileSource).not.toContain(
      'function resolveTabFromLocation(): SkillProfileTabKey {'
    )
    expect(skillProfileSource).not.toContain(
      'function syncPanelToLocation(tabKey: SkillProfileTabKey): void {'
    )
    expect(skillProfileSource).not.toContain('function focusTab(index: number): void {')
    expect(skillProfileSource).not.toContain(
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )
  })
})
