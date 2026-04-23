import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const notificationDropdownComposableSource = readFileSync(
  resolve(process.cwd(), 'src/composables/useNotificationDropdown.ts'),
  'utf8'
)

describe('useNotificationDropdown', () => {
  it('uses theme tokens for notification accents instead of inline rgba or hex colors', () => {
    expect(notificationDropdownComposableSource).toContain("createNotificationTypeMeta(Info, '系统', 'var(--color-primary)', 'var(--color-primary-soft)')")
    expect(notificationDropdownComposableSource).toContain("createNotificationTypeMeta(Trophy, '竞赛', 'var(--color-warning)')")
    expect(notificationDropdownComposableSource).toContain("createNotificationTypeMeta(Flag, '训练', 'var(--color-success)')")
    expect(notificationDropdownComposableSource).toContain("createNotificationTypeMeta(GraduationCap, '团队', 'var(--color-brand-swatch-blue)')")
    expect(notificationDropdownComposableSource).toContain('color-mix(in srgb, ${accentColor} 26%, transparent)')
    expect(notificationDropdownComposableSource).toContain('color-mix(in srgb, ${accentColor} 22%, transparent)')
    expect(notificationDropdownComposableSource).not.toContain('rgba(210, 153, 34, 0.12)')
    expect(notificationDropdownComposableSource).not.toContain('rgba(210, 153, 34, 0.26)')
    expect(notificationDropdownComposableSource).not.toContain('rgba(63, 185, 80, 0.12)')
    expect(notificationDropdownComposableSource).not.toContain('rgba(63, 185, 80, 0.26)')
    expect(notificationDropdownComposableSource).not.toContain('#8b5cf6')
    expect(notificationDropdownComposableSource).not.toContain('#a78bfa')
  })
})
