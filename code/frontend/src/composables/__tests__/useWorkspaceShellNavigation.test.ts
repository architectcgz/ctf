import { describe, expect, it } from 'vitest'

import { createWorkspaceShellNavigation } from '@/composables/useWorkspaceShellNavigation'

describe('useWorkspaceShellNavigation', () => {
  it('builds student workspace modules for the shared shell', () => {
    const shell = createWorkspaceShellNavigation({
      path: '/student/dashboard',
      role: 'student',
    })

    expect(shell.isBackoffice).toBe(false)
    expect(shell.brandKicker).toBe('Student Space')
    expect(shell.brandTitle).toBe('攻防实训平台')
    expect(shell.modules.map((module) => module.label)).toEqual(['训练', '赛事', '账户'])
    expect(shell.modules[0].secondaryItems.map((item) => item.label)).toEqual([
      '仪表盘',
      '题目',
      '我的实例',
      '能力画像',
    ])
  })

  it('keeps academy and platform routes on the configured backoffice modules', () => {
    const shell = createWorkspaceShellNavigation({
      path: '/platform/overview',
      role: 'admin',
    })

    expect(shell.isBackoffice).toBe(true)
    expect(shell.brandKicker).toBe('ChallengeOps')
    expect(shell.brandTitle).toBe('后台工作台')
    expect(shell.modules.map((module) => module.label)).toEqual([
      '总览',
      '教学运营',
      '题库与资源',
      '赛事运维',
      '系统治理',
    ])
  })

  it('resolves active student module and secondary item from nested routes', () => {
    const detailShell = createWorkspaceShellNavigation({
      path: '/challenges/3',
      role: 'student',
    })

    expect(detailShell.activeModuleKey).toBe('training')
    expect(detailShell.activeSecondaryRouteName).toBe('Challenges')
    expect(detailShell.breadcrumb.moduleLabel).toBe('训练')
    expect(detailShell.breadcrumb.secondaryLabel).toBe('题目')
    expect(detailShell.breadcrumb.detailLabel).toBe('题目详情')
  })
})
