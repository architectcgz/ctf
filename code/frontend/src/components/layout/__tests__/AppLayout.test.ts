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

  it('stretches full-bleed route roots so wide screens do not expose main shell gaps', () => {
    expect(appLayoutSource).toContain('.workspace-page--bleed :deep(.workspace-route-root--bleed)')
    expect(appLayoutSource).toContain('flex: 1 1 auto;')
  })

  it('removes vertical main padding for full-bleed routes instead of canceling it with negative top margins', () => {
    expect(appLayoutSource).toContain('mainShellClass')
    expect(appLayoutSource).toContain('workspace-main--bleed')
    expect(appLayoutSource).toContain('padding-block: 0;')
    expect(appLayoutSource).toContain('padding-inline: 0;')
    expect(appLayoutSource).toContain('max-width: none;')
  })

  it('makes the topnav content column a flex stack so main can consume the remaining height', () => {
    expect(appLayoutSource).toContain('<div class="min-w-0 flex flex-1 flex-col">')
    expect(appLayoutSource).toContain('.workspace-main {')
    expect(appLayoutSource).toContain('flex: 1 1 auto;')
    expect(appLayoutSource).toContain('min-height: 0;')
  })
})
