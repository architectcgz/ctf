import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('../browser', () => ({
  redirectTo: vi.fn(),
}))

import {
  isAuthFlowRequest,
  redirectToErrorStatusPage,
  resolveErrorStatusPage,
  resolveErrorStatusRetryTarget,
  shouldRedirectToErrorStatusPage,
} from '../errorStatusPage'
import { redirectTo } from '../browser'

describe('error status page helpers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.history.replaceState({}, '', '/')
  })

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
    expect(isAuthFlowRequest('/auth/cas/login')).toBe(false)
    expect(isAuthFlowRequest('/auth/cas/callback')).toBe(false)
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

  it('redirects server errors to the status page with the current route as retry target', () => {
    window.history.replaceState({}, '', '/student/dashboard?panel=difficulty#chart')

    expect(redirectToErrorStatusPage(500, '/api/dashboard')).toBe(true)
    expect(redirectTo).toHaveBeenCalledWith('/500?from=%2Fstudent%2Fdashboard%3Fpanel%3Ddifficulty%23chart')
  })

  it('reads the retry target from the error page query string', () => {
    expect(resolveErrorStatusRetryTarget('?from=%2Fchallenges%2F5')).toBe('/challenges/5')
    expect(resolveErrorStatusRetryTarget('?from=https%3A%2F%2Fevil.example')).toBeNull()
    expect(resolveErrorStatusRetryTarget('')).toBeNull()
  })
})
