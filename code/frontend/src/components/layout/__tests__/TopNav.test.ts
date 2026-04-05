import { describe, expect, it } from 'vitest'

import topNavSource from '../TopNav.vue?raw'

describe('TopNav layout language', () => {
  it('removes the route protocol label and keeps the header as an action bar', () => {
    expect(topNavSource).not.toContain('route://')
    expect(topNavSource).toContain('class="topnav-main flex min-w-0 items-center gap-3 md:gap-4"')
    expect(topNavSource).toContain('class="topnav-actions flex shrink-0 items-center gap-3"')
    expect(topNavSource).toContain('class="topnav-tool-cluster"')
  })

  it('compresses the user area to name plus role', () => {
    expect(topNavSource).toContain('class="topnav-user-card flex items-center gap-3 px-2.5 py-1.5 sm:px-3"')
    expect(topNavSource).toContain('class="topnav-user-name truncate text-sm font-semibold text-text-primary"')
    expect(topNavSource).not.toContain('topnav-user-meta')
    expect(topNavSource).toContain('topnav-logout')
  })
})
