import { describe, expect, it } from 'vitest'

import {
  isAuthFlowRequest,
  resolveErrorStatusPage,
  shouldRedirectToErrorStatusPage,
} from '../errorStatusPage'

describe('error status page helpers', () => {
  it('maps supported HTTP statuses to dedicated error pages', () => {
    expect(resolveErrorStatusPage(401)).toBe('/401')
    expect(resolveErrorStatusPage(429)).toBe('/429')
    expect(resolveErrorStatusPage(500)).toBe('/500')
    expect(resolveErrorStatusPage(502)).toBe('/502')
    expect(resolveErrorStatusPage(503)).toBe('/503')
    expect(resolveErrorStatusPage(504)).toBe('/504')
    expect(resolveErrorStatusPage(418)).toBeNull()
  })

  it('does not treat auth entry APIs as unauthorized page redirects', () => {
    expect(isAuthFlowRequest('/auth/login')).toBe(true)
    expect(isAuthFlowRequest('/auth/register')).toBe(true)
    expect(isAuthFlowRequest('/auth/cas/login')).toBe(true)
    expect(isAuthFlowRequest('/auth/cas/callback')).toBe(true)
    expect(isAuthFlowRequest('/auth/profile')).toBe(false)
    expect(isAuthFlowRequest('/challenges')).toBe(false)
  })

  it('redirects only for supported status pages outside auth entry flows', () => {
    expect(shouldRedirectToErrorStatusPage(401, '/auth/login')).toBe(false)
    expect(shouldRedirectToErrorStatusPage(401, '/auth/register')).toBe(false)
    expect(shouldRedirectToErrorStatusPage(401, '/auth/profile')).toBe(true)
    expect(shouldRedirectToErrorStatusPage(429, '/scoreboard')).toBe(true)
    expect(shouldRedirectToErrorStatusPage(500, '/notifications')).toBe(true)
    expect(shouldRedirectToErrorStatusPage(404, '/notifications')).toBe(false)
  })
})
