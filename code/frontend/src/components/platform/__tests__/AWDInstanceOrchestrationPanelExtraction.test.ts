import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdInstanceOrchestrationHeaderSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationHeader.vue?raw'
import awdInstanceOrchestrationMatrixSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationMatrix.vue?raw'
import awdInstanceOrchestrationPanelSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue?raw'
import awdInstanceOrchestrationRowSource from '@/features/contest-awd-admin/ui/AWDInstanceOrchestrationRow.vue?raw'

const awdInstanceOrchestrationCombinedSource = [
  awdInstanceOrchestrationPanelSource,
  awdInstanceOrchestrationHeaderSource,
  awdInstanceOrchestrationMatrixSource,
  awdInstanceOrchestrationRowSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdInstanceOrchestrationPanel.css'),
    'utf8'
  ),
].join('\n')

describe('AWDInstanceOrchestrationPanel extraction', () => {
  it('AWDInstanceOrchestrationPanel 应把 header、matrix 和 row 视图下沉到独立子组件', () => {
    expect(awdInstanceOrchestrationPanelSource).toContain('<AWDInstanceOrchestrationHeader')
    expect(awdInstanceOrchestrationPanelSource).toContain('<AWDInstanceOrchestrationMatrix')
    expect(awdInstanceOrchestrationPanelSource).toContain('const headerSummary = computed')
    expect(awdInstanceOrchestrationPanelSource).toContain('const rowViews = computed')
    expect(awdInstanceOrchestrationPanelSource).not.toContain('class="orchestration-header"')
    expect(awdInstanceOrchestrationPanelSource).not.toContain('class="orchestration-table-wrap"')
    expect(awdInstanceOrchestrationPanelSource).not.toContain('<style scoped>')
  })

  it('提取后的子组件应继续承接 action primitive、matrix shell 和状态展示', () => {
    expect(awdInstanceOrchestrationHeaderSource).toContain('class="ops-btn ops-btn--neutral"')
    expect(awdInstanceOrchestrationHeaderSource).toContain('class="ops-btn ops-btn--primary"')
    expect(awdInstanceOrchestrationMatrixSource).toContain('class="orchestration-table-wrap"')
    expect(awdInstanceOrchestrationRowSource).toContain('class="cell-start-btn"')
    expect(awdInstanceOrchestrationRowSource).toContain('class="row-start-btn"')
    expect(awdInstanceOrchestrationCombinedSource).toContain('.instance-status--running')
  })
})
