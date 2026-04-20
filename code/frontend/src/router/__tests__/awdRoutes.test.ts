import { describe, expect, it } from 'vitest'

import { routes } from '@/router'

function getRootChildren() {
  const root = routes.find((route) => route.path === '/')
  return root?.children ?? []
}

function findChild(path: string) {
  return getRootChildren().find((route) => route.path === path)
}

describe('awd workspace routes', () => {
  it('covers all 19 documented awd page routes', () => {
    const awdPageRoutes = [
      ['contests/:id/awd/overview', 'ContestAwdOverview'],
      ['contests/:id/awd/services', 'ContestAwdServices'],
      ['contests/:id/awd/targets', 'ContestAwdTargets'],
      ['contests/:id/awd/attacks', 'ContestAwdAttacks'],
      ['contests/:id/awd/collab', 'ContestAwdCollab'],
      ['academy/awd-reviews/:contestId/overview', 'TeacherAwdOverview'],
      ['academy/awd-reviews/:contestId/teams', 'TeacherAwdTeams'],
      ['academy/awd-reviews/:contestId/services', 'TeacherAwdServices'],
      ['academy/awd-reviews/:contestId/replay', 'TeacherAwdReplay'],
      ['academy/awd-reviews/:contestId/export', 'TeacherAwdExport'],
      ['platform/contests/:id/awd/overview', 'AdminAwdOverview'],
      ['platform/contests/:id/awd/readiness', 'AdminAwdReadiness'],
      ['platform/contests/:id/awd/rounds', 'AdminAwdRounds'],
      ['platform/contests/:id/awd/services', 'AdminAwdServices'],
      ['platform/contests/:id/awd/attacks', 'AdminAwdAttacks'],
      ['platform/contests/:id/awd/traffic', 'AdminAwdTraffic'],
      ['platform/contests/:id/awd/alerts', 'AdminAwdAlerts'],
      ['platform/contests/:id/awd/instances', 'AdminAwdInstances'],
      ['platform/contests/:id/awd/replay', 'AdminAwdReplay'],
    ] as const

    expect(awdPageRoutes).toHaveLength(19)

    awdPageRoutes.forEach(([path, name]) => {
      expect(findChild(path)?.name).toBe(name)
    })
  })

  it('redirects student and admin awd base entries to the new overview pages', () => {
    const studentRedirect = findChild('contests/:id/awd')?.redirect
    const adminRedirect = findChild('platform/contests/:id/awd')?.redirect

    expect(typeof studentRedirect).toBe('function')
    expect(typeof adminRedirect).toBe('function')

    if (typeof studentRedirect !== 'function' || typeof adminRedirect !== 'function') {
      throw new Error('AWD redirect routes should use function redirects')
    }

    const from = { params: {}, query: {}, hash: '' } as never

    expect(
      studentRedirect({ params: { id: '42' }, query: { tab: 'services' }, hash: '#logs' } as never, from)
    ).toEqual({
      name: 'ContestAwdOverview',
      params: { id: '42' },
      query: { tab: 'services' },
      hash: '#logs',
    })

    expect(
      adminRedirect({ params: { id: '9' }, query: { stage: 'ops' }, hash: '#traffic' } as never, from)
    ).toEqual(
      {
        name: 'AdminAwdOverview',
        params: { id: '9' },
        query: { stage: 'ops' },
        hash: '#traffic',
      }
    )
  })
})
