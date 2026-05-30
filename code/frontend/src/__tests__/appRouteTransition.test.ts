import { describe, expect, it } from 'vitest'

import appSource from '@/App.vue?raw'

describe('app route transition', () => {
  it('wraps top-level routes without re-animating every nested app route', () => {
    expect(appSource).toContain('<RouterView v-slot="{ Component, route: resolvedRoute }">')
    expect(appSource).toContain('name="app-route"')
    expect(appSource).toContain('mode="out-in"')
    expect(appSource).toContain(':key="resolvedRoute.matched[0]?.path || resolvedRoute.path"')
  })
})
