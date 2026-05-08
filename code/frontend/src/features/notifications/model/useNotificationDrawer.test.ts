import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

const notificationDrawerComposableSource = readFileSync(
  resolve(process.cwd(), 'src/features/notifications/model/useNotificationDrawer.ts'),
  'utf8'
)

describe('useNotificationDrawer', () => {
  it('uses theme tokens for notification accents instead of inline rgba or hex colors', () => {
    expect(notificationDrawerComposableSource).toContain("createNotificationTypeMeta(Info, '系统', 'var(--color-primary)', 'var(--color-primary-soft)')")
    expect(notificationDrawerComposableSource).toContain("createNotificationTypeMeta(Trophy, '竞赛', 'var(--color-warning)')")
    expect(notificationDrawerComposableSource).toContain("createNotificationTypeMeta(Flag, '训练', 'var(--color-success)')")
    expect(notificationDrawerComposableSource).toContain("createNotificationTypeMeta(GraduationCap, '团队', 'var(--color-brand-swatch-blue)')")
    expect(notificationDrawerComposableSource).toContain('color-mix(in srgb, ${accentColor} 26%, transparent)')
    expect(notificationDrawerComposableSource).toContain('color-mix(in srgb, ${accentColor} 22%, transparent)')
    expect(notificationDrawerComposableSource).not.toContain('rgba(210, 153, 34, 0.12)')
    expect(notificationDrawerComposableSource).not.toContain('rgba(210, 153, 34, 0.26)')
    expect(notificationDrawerComposableSource).not.toContain('rgba(63, 185, 80, 0.12)')
    expect(notificationDrawerComposableSource).not.toContain('rgba(63, 185, 80, 0.26)')
    expect(notificationDrawerComposableSource).not.toContain('#8b5cf6')
    expect(notificationDrawerComposableSource).not.toContain('#a78bfa')
  })
})
