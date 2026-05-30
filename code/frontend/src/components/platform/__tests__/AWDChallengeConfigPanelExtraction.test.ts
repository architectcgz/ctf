import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

import awdChallengeConfigDirectoryRowSource from '@/features/platform-contests/ui/AWDChallengeConfigDirectoryRow.vue?raw'
import awdChallengeConfigDirectorySectionSource from '@/features/platform-contests/ui/AWDChallengeConfigDirectorySection.vue?raw'
import awdChallengeConfigHeaderSource from '@/features/platform-contests/ui/AWDChallengeConfigHeader.vue?raw'
import awdChallengeConfigPanelSource from '@/features/platform-contests/ui/AWDChallengeConfigPanel.vue?raw'

const awdChallengeConfigCombinedSource = [
  awdChallengeConfigPanelSource,
  awdChallengeConfigHeaderSource,
  awdChallengeConfigDirectorySectionSource,
  awdChallengeConfigDirectoryRowSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/platform-contests/ui/awdChallengeConfigPanel.css'),
    'utf8'
  ),
].join('\n')

describe('AWDChallengeConfigPanel extraction', () => {
  it('AWDChallengeConfigPanel 应将 header、directory section 和 row 视图下沉到独立子组件', () => {
    expect(awdChallengeConfigPanelSource).toContain('<AWDChallengeConfigHeader')
    expect(awdChallengeConfigPanelSource).toContain('<AWDChallengeConfigDirectorySection')
    expect(awdChallengeConfigPanelSource).toContain('const directoryItems = computed')
    expect(awdChallengeConfigPanelSource).toContain('useAwdCheckResultPresentation')
    expect(awdChallengeConfigPanelSource).not.toContain('class="studio-pane-header"')
    expect(awdChallengeConfigPanelSource).not.toContain('class="studio-table-wrap"')
    expect(awdChallengeConfigPanelSource).not.toContain('<style scoped>')
  })

  it('提取后的子组件应继续承接 metric panel、directory shell 与 row action primitive', () => {
    expect(awdChallengeConfigHeaderSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(awdChallengeConfigDirectorySectionSource).toContain(
      'class="workspace-directory-section awd-config-directory"'
    )
    expect(awdChallengeConfigDirectorySectionSource).toContain('class="studio-table-wrap"')
    expect(awdChallengeConfigDirectoryRowSource).toContain(
      'class="ui-row-actions config-row__actions"'
    )
    expect(awdChallengeConfigDirectoryRowSource).toContain('class="ui-btn ui-btn--primary"')
    expect(awdChallengeConfigCombinedSource).toContain('.validation-pill.passed')
  })
})
