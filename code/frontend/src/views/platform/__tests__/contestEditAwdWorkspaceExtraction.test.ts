import { describe, expect, it } from 'vitest'

import contestEditSource from '@/views/platform/ContestEdit.vue?raw'

describe('ContestEdit AWD workspace extraction', () => {
  it('应将 AWD 工作区状态与操作抽到独立 composable', () => {
    expect(contestEditSource).toContain(
      "import { useContestEditAwdWorkspace } from '@/composables/useContestEditAwdWorkspace'"
    )
    expect(contestEditSource).toContain('} = useContestEditAwdWorkspace({')
  })
})
