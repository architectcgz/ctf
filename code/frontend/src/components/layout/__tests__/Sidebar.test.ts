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
})
