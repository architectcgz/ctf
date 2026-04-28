import { describe, expect, it } from 'vitest'

import contestManageSource from '../ContestManage.vue?raw'

describe('ContestManage export flow extraction', () => {
  it('赛事目录不再直接挂载导出流程，结果导出应进入赛事运维处理', () => {
    expect(contestManageSource).toContain(
      "import ContestOrchestrationPage from '@/components/platform/contest/ContestOrchestrationPage.vue'"
    )
    expect(contestManageSource).not.toContain('useContestExportFlow')
    expect(contestManageSource).not.toContain('@export-contest')
  })
})
