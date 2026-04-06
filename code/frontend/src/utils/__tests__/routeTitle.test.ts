import { describe, expect, it } from 'vitest'

import { resolveRouteTitle } from '../routeTitle'

describe('resolveRouteTitle', () => {
  it('应该识别学生端仪表盘路径与子面板标题', () => {
    expect(resolveRouteTitle({ path: '/student/dashboard' })).toBe('仪表盘')
    expect(
      resolveRouteTitle({
        path: '/student/dashboard',
        query: { panel: 'timeline' },
      })
    ).toBe('近期动态')
  })
})
