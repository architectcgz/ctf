import { describe, expect, it } from 'vitest'

import contestManageSource from '../ContestManage.vue?raw'

describe('ContestManage export flow extraction', () => {
  it('应将赛事结果导出与轮询下载逻辑抽到独立 composable', () => {
    expect(contestManageSource).toContain(
      "import { useContestExportFlow } from '@/composables/useContestExportFlow'"
    )
    expect(contestManageSource).toContain('const { handleExportContest } = useContestExportFlow()')
  })
})
