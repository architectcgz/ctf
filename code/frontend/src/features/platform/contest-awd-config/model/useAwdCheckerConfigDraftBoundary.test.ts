import { describe, expect, it } from 'vitest'

import draftSource from './useAwdCheckerConfigDraft.ts?raw'
import supportSource from './awdCheckerConfigSupport.ts?raw'

describe('useAwdCheckerConfigDraft boundary', () => {
  it('应组合 draft hydration 与 tcp step actions，避免主模块内联草稿回填和步骤交互实现', () => {
    expect(draftSource).toContain("from './useAwdCheckerDraftHydration'")
    expect(draftSource).toContain("from './useAwdTcpStepActions'")
    expect(draftSource).not.toContain('function assignHTTPDraft(')
    expect(draftSource).not.toContain('function assignTCPDraft(')
    expect(draftSource).not.toContain('function assignScriptDraft(')
    expect(draftSource).not.toContain('function toggleTCPCheckerStep(')
  })

  it('checker support 应拆分 script/tcp 构建逻辑到独立模块', () => {
    expect(supportSource).toContain("from './awdCheckerScriptConfigSupport'")
    expect(supportSource).toContain("from './awdCheckerTcpConfigSupport'")
    expect(supportSource).not.toContain('function buildScriptCheckerConfig(')
    expect(supportSource).not.toContain('function buildTCPStandardCheckerConfig(')
  })
})
