import { describe, expect, it } from 'vitest'

import {
  ADMIN_AWD_PAGES,
  STUDENT_AWD_PAGES,
  TEACHER_AWD_PAGES,
  buildAdminAwdPath,
  buildStudentAwdPath,
  buildTeacherAwdPath,
} from '@/modules/awd/navigation'

describe('awd navigation definitions', () => {
  it('covers all 19 documented awd pages', () => {
    expect(STUDENT_AWD_PAGES).toHaveLength(5)
    expect(ADMIN_AWD_PAGES).toHaveLength(9)
    expect(TEACHER_AWD_PAGES).toHaveLength(5)
  })

  it('builds role-specific awd paths', () => {
    expect(buildStudentAwdPath('42', 'overview')).toBe('/contests/42/awd/overview')
    expect(buildAdminAwdPath('9', 'traffic')).toBe('/platform/contests/9/awd/traffic')
    expect(buildTeacherAwdPath('5', 'export')).toBe('/academy/awd-reviews/5/export')
  })

  it('keeps stable page keys for all role navigation groups', () => {
    expect(STUDENT_AWD_PAGES.map((item) => item.key)).toEqual([
      'overview',
      'services',
      'targets',
      'attacks',
      'collab',
    ])
    expect(ADMIN_AWD_PAGES.map((item) => item.key)).toEqual([
      'overview',
      'readiness',
      'rounds',
      'services',
      'attacks',
      'traffic',
      'alerts',
      'instances',
      'replay',
    ])
    expect(TEACHER_AWD_PAGES.map((item) => item.key)).toEqual([
      'overview',
      'teams',
      'services',
      'replay',
      'export',
    ])
  })
})
