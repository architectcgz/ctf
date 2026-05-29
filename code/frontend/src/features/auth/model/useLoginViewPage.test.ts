import { describe, expect, it } from 'vitest'

import { DEFAULT_LOGIN_REDIRECT, resolveLoginRedirectTarget } from './useLoginViewPage'

describe('useLoginViewPage helpers', () => {
  it('根路径 redirect 时应回退到默认目标', () => {
    expect(resolveLoginRedirectTarget(DEFAULT_LOGIN_REDIRECT, '/academy/overview')).toBe(
      '/academy/overview'
    )
  })

  it('携带显式 redirect 时应保留原目标', () => {
    expect(resolveLoginRedirectTarget('/contests/1', '/academy/overview')).toBe('/contests/1')
  })
})
