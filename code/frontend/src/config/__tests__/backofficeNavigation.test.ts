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
    expect(items.find((item) => item.active)?.routeName).toBe('AdminStudentManagement')
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

    expect(items.map((item) => item.label)).toEqual(['题目管理', '环境模板', '镜像管理'])
    expect(items.find((item) => item.active)?.routeName).toBe('ChallengeManage')
  })

  it('maps admin event operations routes back to 赛事运维 and marks the matched secondary item active', () => {
    expect(getBackofficeModuleByPath('/admin/contest-ops/environment')?.key).toBe('contestOps')

    const items = getVisibleBackofficeSecondaryItems('/admin/contest-ops/traffic', 'admin')

    expect(items.map((item) => item.label)).toEqual([
      '环境管理',
      '流量监控',
      '大屏投射',
      '排行榜',
    ])
    expect(items.find((item) => item.active)?.routeName).toBe('AdminContestOpsTraffic')
  })
})
