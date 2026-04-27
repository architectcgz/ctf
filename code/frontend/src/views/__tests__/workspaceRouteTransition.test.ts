import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const styleSource = readFileSync(resolve(__dirname, '../../style.css'), 'utf8')

describe('workspace route transition styles', () => {
  it('defines a restrained global route transition with reduced-motion support', () => {
    expect(styleSource).toContain('--workspace-route-motion-enter-duration: 280ms;')
    expect(styleSource).toContain('--workspace-route-motion-leave-duration: 180ms;')
    expect(styleSource).toContain('.app-route-enter-active')
    expect(styleSource).toContain('.app-route-leave-active')
    expect(styleSource).toContain('.workspace-route-enter-active')
    expect(styleSource).toContain('.workspace-route-leave-active')
    expect(styleSource).toContain('transform: translate3d(0, var(--space-3), 0);')
    expect(styleSource).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
