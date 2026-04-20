import { describe, expect, it } from 'vitest'

import awdWorkspaceSource from '@/components/contests/ContestAWDWorkspacePanel.vue?raw'
import contestDetailSource from '@/views/contests/ContestDetail.vue?raw'

describe('contest student action primitives', () => {
  it('ContestDetail 应接入共享 ui 按钮与输入控件原语', () => {
    expect(contestDetailSource).toMatch(/class="ui-control-wrap(?:\s+[^\"]+)?"/)
    expect(contestDetailSource).toContain('class="ui-control"')
    expect(contestDetailSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestDetailSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestDetailSource).not.toMatch(/^\.contest-btn\s*\{/m)
    expect(contestDetailSource).not.toMatch(/^\.contest-btn--primary\s*\{/m)
    expect(contestDetailSource).not.toMatch(/^\.contest-btn--ghost\s*\{/m)
    expect(contestDetailSource).not.toMatch(/^\.flag-submit__input\s*\{/m)
    expect(contestDetailSource).not.toMatch(/^\.flag-submit__input:focus\s*\{/m)
  })

  it('AWD 工作台也应复用共享 ui 按钮与输入控件原语', () => {
    expect(awdWorkspaceSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(awdWorkspaceSource).toContain('class="ui-btn ui-btn--primary"')
    expect(awdWorkspaceSource).toMatch(/class="ui-control-wrap(?:\s+[^\"]+)?"/)
    expect(awdWorkspaceSource).toContain('class="ui-control"')
    expect(awdWorkspaceSource).not.toMatch(/^\.contest-btn\s*\{/m)
    expect(awdWorkspaceSource).not.toMatch(/^\.contest-btn--primary\s*\{/m)
    expect(awdWorkspaceSource).not.toMatch(/^\.contest-btn--ghost\s*\{/m)
    expect(awdWorkspaceSource).not.toMatch(/^\.flag-submit__input\s*\{/m)
    expect(awdWorkspaceSource).not.toMatch(/^\.awd-target-select\s*\{/m)
  })
})
