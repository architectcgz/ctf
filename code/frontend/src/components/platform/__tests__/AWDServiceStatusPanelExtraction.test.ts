import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdServiceRoundPerformanceTableSource from '@/features/awd-inspector/ui/AWDServiceRoundPerformanceTable.vue?raw'
import awdServiceStatusMatrixSource from '@/features/awd-inspector/ui/AWDServiceStatusMatrix.vue?raw'
import awdServiceStatusPanelSource from '@/features/awd-inspector/ui/AWDServiceStatusPanel.vue?raw'
import awdServiceStatusToolbarSource from '@/features/awd-inspector/ui/AWDServiceStatusToolbar.vue?raw'

const awdServiceStatusCombinedSource = [
  awdServiceStatusPanelSource,
  awdServiceStatusToolbarSource,
  awdServiceStatusMatrixSource,
  awdServiceRoundPerformanceTableSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/awd-inspector/ui/awdServiceStatusPanel.css'),
    'utf8'
  ),
].join('\n')

describe('AWDServiceStatusPanel extraction', () => {
  it('AWDServiceStatusPanel 应把 toolbar、matrix 和 round performance table 下沉到独立子组件', () => {
    expect(awdServiceStatusPanelSource).toContain('<AWDServiceStatusToolbar')
    expect(awdServiceStatusPanelSource).toContain('<AWDServiceStatusMatrix')
    expect(awdServiceStatusPanelSource).toContain('<AWDServiceRoundPerformanceTable')
    expect(awdServiceStatusPanelSource).toContain('const serviceStatusRows = computed')
    expect(awdServiceStatusPanelSource).not.toContain('class="matrix-toolbar"')
    expect(awdServiceStatusPanelSource).not.toContain('class="matrix-scroll custom-scrollbar"')
    expect(awdServiceStatusPanelSource).not.toContain('class="round-performance-area mt-12"')
    expect(awdServiceStatusPanelSource).not.toContain('<style scoped>')
  })

  it('提取后的子组件应继续承接 filter shell、matrix status card 与 round performance summary', () => {
    expect(awdServiceStatusToolbarSource).toContain('id="awd-service-filter-team"')
    expect(awdServiceStatusToolbarSource).toContain('id="awd-service-filter-status"')
    expect(awdServiceStatusToolbarSource).toContain('id="awd-service-filter-source"')
    expect(awdServiceStatusToolbarSource).toContain('id="awd-service-filter-alert"')
    expect(awdServiceStatusToolbarSource).toContain('id="awd-export-services"')
    expect(awdServiceStatusMatrixSource).toContain('class="status-box"')
    expect(awdServiceRoundPerformanceTableSource).toContain('Round Performance Summary')
    expect(awdServiceStatusCombinedSource).toContain('.status--up')
  })
})
