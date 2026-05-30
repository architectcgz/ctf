import { describe, expect, it } from 'vitest'

import contestProjectorSource from '@/pages/platform/contests/ContestProjectorRoutePage.vue?raw'

describe('ContestProjector page state extraction', () => {
  it('应通过 feature page model 获取大屏生命周期和展示派生状态', () => {
    expect(contestProjectorSource).toContain("from '@/features/contest-projector'")
    expect(contestProjectorSource).toContain('useContestProjectorPage')
    expect(contestProjectorSource).toContain('ContestProjectorAttackMap')
    expect(contestProjectorSource).toContain('ContestProjectorFocusOverlay')
    expect(contestProjectorSource).not.toContain("from '@/components/platform/contest/projector/")
    expect(contestProjectorSource).not.toContain("from '@/composables/useToast'")
    expect(contestProjectorSource).not.toContain('const projectorStageRef = ref')
    expect(contestProjectorSource).not.toContain('const contestTitle = computed')
    expect(contestProjectorSource).not.toContain('function focusPanel(')
    expect(contestProjectorSource).not.toContain('async function toggleFullscreen(')
    expect(contestProjectorSource).not.toContain('onMounted(() =>')
    expect(contestProjectorSource).not.toContain('onUnmounted(() =>')
  })
})
