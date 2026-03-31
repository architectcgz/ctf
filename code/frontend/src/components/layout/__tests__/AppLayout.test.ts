import { describe, expect, it } from 'vitest'

import appLayoutSource from '../AppLayout.vue?raw'
import routerSource from '../../../router/index.ts?raw'

describe('AppLayout workspace shell', () => {
  it('owns full-bleed page spacing and drives it from route meta', () => {
    expect(appLayoutSource).toContain('<RouterView v-slot="{ Component }">')
    expect(appLayoutSource).toContain('workspace-page')
    expect(appLayoutSource).toContain('workspace-page--bleed')
    expect(appLayoutSource).toContain('pageShellClass')
    expect(appLayoutSource).toContain('workspace-route-root')
    expect(appLayoutSource).toContain('workspace-route-root--bleed')
    expect(routerSource).toContain("contentLayout: 'bleed'")
    expect((routerSource.match(/contentLayout: 'bleed'/g) ?? []).length).toBeGreaterThanOrEqual(30)
  })
})
