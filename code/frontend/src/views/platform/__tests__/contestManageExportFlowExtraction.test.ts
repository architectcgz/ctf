import { describe, expect, it } from 'vitest'

import contestManageSource from '@/pages/platform/contests/ContestManageRoutePage.vue?raw'

describe('ContestManage export flow extraction', () => {
  it('赛事目录不再直接挂载导出流程，结果导出应进入赛事运维处理', () => {
    expect(contestManageSource).toContain("from '@/features/platform/contests'")
    expect(contestManageSource).toContain('ContestOrchestrationPage')
    expect(contestManageSource).toContain('useContestManagePage')
    expect(contestManageSource).not.toContain('useContestExportFlow')
    expect(contestManageSource).not.toContain('@export-contest')
  })
})
