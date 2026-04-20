import { describe, expect, it } from 'vitest'

import studentAwdCollabSource from '@/modules/awd/views/student/StudentAwdCollabView.vue?raw'
import studentAwdOverviewSource from '@/modules/awd/views/student/StudentAwdOverviewView.vue?raw'
import studentAwdServicesSource from '@/modules/awd/views/student/StudentAwdServicesView.vue?raw'
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

  it('学生 AWD 页面族应统一接入共享 shell，并保持旧面板完全移除', () => {
    expect(studentAwdOverviewSource).toContain('StudentAwdWorkspaceLayout')
    expect(studentAwdServicesSource).toContain('StudentAwdWorkspaceLayout')
    expect(studentAwdCollabSource).toContain('StudentAwdWorkspaceLayout')
    expect(studentAwdServicesSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(studentAwdOverviewSource).not.toContain('ContestAWDWorkspacePanel')
    expect(studentAwdServicesSource).not.toContain('ContestAWDWorkspacePanel')
    expect(studentAwdCollabSource).not.toContain('ContestAWDWorkspacePanel')
  })
})
