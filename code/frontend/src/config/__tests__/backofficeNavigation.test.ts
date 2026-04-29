import { describe, expect, it } from 'vitest'

import {
  getBackofficeModuleByPath,
  getVisibleBackofficeModules,
  getVisibleBackofficeSecondaryItems,
} from '../backofficeNavigation'

describe('backofficeNavigation', () => {
  it('maps platform challenge detail routes back to the 题库与资源 module', () => {
    expect(getBackofficeModuleByPath('/platform/challenges/11/topology')?.key).toBe('resources')
    expect(getBackofficeModuleByPath('/platform/challenges/imports/import-1')?.key).toBe(
      'resources'
    )
  })

  it('maps teacher detail routes back to 教学运营', () => {
    expect(
      getBackofficeModuleByPath('/academy/classes/class-a/students/student-1/review-archive')?.key
    ).toBe('operations')
  })

  it('maps admin class management back to 教学运营 and exposes the platform class entry', () => {
    expect(getBackofficeModuleByPath('/platform/classes')?.key).toBe('operations')

    const items = getVisibleBackofficeSecondaryItems('/platform/classes', 'admin')

    expect(items.map((item) => item.label)).toEqual([
      '班级管理',
      '学生管理',
      'AWD复盘',
      '实例管理',
    ])
    expect(items.find((item) => item.active)?.path).toBe('/platform/classes')
  })

  it('keeps admin teaching detail routes inside the platform teaching workspace', () => {
    expect(getBackofficeModuleByPath('/platform/classes/class-a/students/student-1')?.key).toBe(
      'operations'
    )
    expect(getBackofficeModuleByPath('/platform/awd-reviews/contest-1')?.key).toBe('operations')
    expect(getBackofficeModuleByPath('/platform/instances')?.key).toBe('operations')

    const items = getVisibleBackofficeSecondaryItems(
      '/platform/classes/class-a/students/student-1/review-archive',
      'admin'
    )

    expect(items.map((item) => item.path)).toEqual([
      '/platform/classes',
      '/platform/students',
      '/platform/awd-reviews',
      '/platform/instances',
    ])
    expect(items.find((item) => item.active)?.routeName).toBe('PlatformStudentManagement')
  })

  it('filters visible modules by role', () => {
    expect(getVisibleBackofficeModules('teacher').map((item) => item.key)).toEqual([
      'overview',
      'operations',
      'resources',
    ])
    expect(getVisibleBackofficeModules('admin').map((item) => item.key)).toEqual([
      'overview',
      'operations',
      'resources',
      'contestOps',
      'governance',
    ])
  })

  it('resolves secondary items from detail routes and preserves the owning entry active state', () => {
    const items = getVisibleBackofficeSecondaryItems('/platform/challenges/11/writeup', 'admin')

    expect(items.map((item) => item.label)).toEqual([
      '题目管理',
      'AWD题库',
      '镜像管理',
    ])
    expect(items.find((item) => item.active)?.routeName).toBe('ChallengeManage')
  })

  it('keeps AWD challenge import inside the AWD challenge library entry', () => {
    const items = getVisibleBackofficeSecondaryItems('/platform/awd-challenges/imports', 'admin')

    expect(items.find((item) => item.active)?.routeName).toBe('PlatformAwdChallengeLibrary')
    expect(getBackofficeModuleByPath('/platform/awd-challenges/imports')?.key).toBe('resources')
  })

  it('maps admin event operations routes back to 赛事运维 and marks the matched secondary item active', () => {
    expect(getBackofficeModuleByPath('/platform/contest-ops/contests')?.key).toBe('contestOps')
    expect(getBackofficeModuleByPath('/platform/contests/contest-1/manage')?.key).toBe('contestOps')

    const items = getVisibleBackofficeSecondaryItems('/platform/contest-ops/contests', 'admin')

    expect(items.map((item) => item.label)).toEqual(['竞赛列表', '大屏展示'])
    expect(items.find((item) => item.active)?.routeName).toBe('PlatformContestOpsIndex')

    const manageItems = getVisibleBackofficeSecondaryItems('/platform/contests/contest-1/manage', 'admin')
    expect(manageItems.find((item) => item.active)?.routeName).toBe('PlatformContestOpsIndex')

    const projectorItems = getVisibleBackofficeSecondaryItems('/platform/contest-ops/projector', 'admin')
    expect(projectorItems.find((item) => item.active)?.routeName).toBe('PlatformContestProjector')
  })
})
