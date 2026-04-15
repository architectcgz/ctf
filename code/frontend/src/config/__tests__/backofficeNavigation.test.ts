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
      'governance',
    ])
  })

  it('resolves secondary items from detail routes and preserves the owning entry active state', () => {
    const items = getVisibleBackofficeSecondaryItems('/platform/challenges/11/writeup', 'admin')

    expect(items.map((item) => item.label)).toEqual(['题目管理', '环境模板', '镜像管理'])
    expect(items.find((item) => item.active)?.routeName).toBe('ChallengeManage')
  })
})
