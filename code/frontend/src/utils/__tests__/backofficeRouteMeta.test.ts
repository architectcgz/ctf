import { describe, expect, it } from 'vitest'

import {
  getBackofficeActiveSecondaryRouteName,
  getBackofficeLayoutMode,
  isBackofficeRoute,
} from '../backofficeRouteMeta'

describe('backofficeRouteMeta', () => {
  it('marks academy and platform paths as backoffice routes', () => {
    expect(isBackofficeRoute('/academy/overview')).toBe(true)
    expect(isBackofficeRoute('/platform/challenges')).toBe(true)
    expect(isBackofficeRoute('/platform/users')).toBe(true)
    expect(isBackofficeRoute('/student/dashboard')).toBe(false)
  })

  it('returns the expected layout mode for backoffice and c-end routes', () => {
    expect(getBackofficeLayoutMode('/academy/classes')).toBe('backoffice')
    expect(getBackofficeLayoutMode('/student/dashboard')).toBe('default')
  })

  it('maps deep backoffice routes back to their owning secondary entry', () => {
    expect(getBackofficeActiveSecondaryRouteName('/platform/contests/77/edit')).toBe(
      'ContestManage'
    )
    expect(getBackofficeActiveSecondaryRouteName('/academy/awd-reviews/contest-1')).toBe(
      'TeacherAWDReviewIndex'
    )
  })
})
