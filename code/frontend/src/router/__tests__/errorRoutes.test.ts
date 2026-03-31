import { describe, expect, it } from 'vitest'

import { routes } from '../index'

describe('error routes', () => {
  it('registers explicit status-code error pages and redirects unknown paths to /404', () => {
    const topLevelPaths = routes.map((route) => route.path)

    expect(topLevelPaths).toContain('/401')
    expect(topLevelPaths).toContain('/403')
    expect(topLevelPaths).toContain('/404')
    expect(topLevelPaths).toContain('/429')
    expect(topLevelPaths).toContain('/500')
    expect(topLevelPaths).toContain('/502')
    expect(topLevelPaths).toContain('/503')
    expect(topLevelPaths).toContain('/504')

    const catchAllRoute = routes.find((route) => route.path === '/:pathMatch(.*)*')
    expect(catchAllRoute?.redirect).toBe('/404')
  })
})
