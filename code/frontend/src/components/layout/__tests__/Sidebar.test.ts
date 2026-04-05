import { describe, expect, it } from 'vitest'

import sidebarSource from '../Sidebar.vue?raw'

describe('Sidebar desktop layout', () => {
  it('stretches the desktop nav to align its bottom edge with the content area', () => {
    expect(sidebarSource).toMatch(
      /<aside\s+class="[^"]*sidebar-shell-desktop[^"]*min-h-screen[^"]*self-stretch[^"]*"/s
    )
    expect(sidebarSource).toMatch(
      /<nav class="[^"]*flex[^"]*min-h-full[^"]*flex-col[^"]*space-y-7[^"]*">/s
    )
  })

  it('uses a flatter console navigation system instead of stacked card buttons', () => {
    expect(sidebarSource).toContain('class="sidebar-brand-button flex min-w-0 items-center gap-3 px-2.5 py-2 text-left transition"')
    expect(sidebarSource).toContain('sidebar-nav-scroll')
    expect(sidebarSource).toContain('sidebar-group-title--collapsed')
    expect(sidebarSource).toContain('.sidebar-item-active::before,')
    expect(sidebarSource).toContain('.sidebar-item-button--collapsed::before')
  })
})
