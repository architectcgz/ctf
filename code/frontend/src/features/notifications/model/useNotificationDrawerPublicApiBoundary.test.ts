import { describe, expect, it } from 'vitest'

import notificationsModelIndexSource from './index.ts?raw'
import useNotificationDrawerSource from './useNotificationDrawer.ts?raw'

describe('notification drawer public API boundaries', () => {
  it('model/index.ts 应通过 barrel export 暴露 useNotificationDrawer', () => {
    expect(notificationsModelIndexSource).toContain("export * from './useNotificationDrawer'")
  })

  it('useNotificationDrawer 应通过 actions 参数接收导航回调，不直接 import vue-router', () => {
    expect(useNotificationDrawerSource).not.toMatch(/from\s+['"]vue-router['"]/)
    expect(useNotificationDrawerSource).not.toMatch(/\buseRouter\s*\(/)
    expect(useNotificationDrawerSource).not.toMatch(/\brouter\.push\s*\(/)
  })

  it('useNotificationDrawer 应接收 actions 参数，由调用方注入 goToNotifications / goToNotificationDetail', () => {
    expect(useNotificationDrawerSource).toContain('actions: UseNotificationDrawerActions')
    expect(useNotificationDrawerSource).toContain('goToNotifications')
    expect(useNotificationDrawerSource).toContain('goToNotificationDetail')
  })

  it('useNotificationDrawer 返回的接口应包含 open / toggleOpen / close / items / unreadCount', () => {
    expect(useNotificationDrawerSource).toContain('return {')
    expect(useNotificationDrawerSource).toContain('open,')
    expect(useNotificationDrawerSource).toContain('toggleOpen,')
    expect(useNotificationDrawerSource).toContain('close,')
    expect(useNotificationDrawerSource).toContain('items,')
    expect(useNotificationDrawerSource).toContain('unreadCount,')
  })

  it('NotificationTypeMeta 应来自 entities/notification，不在 feature 或 shared 中重复定义', () => {
    expect(useNotificationDrawerSource).toContain("import type { NotificationTypeMeta } from '@/entities/notification'")
    // 不应在 feature 内部重新定义
    expect(useNotificationDrawerSource).not.toContain('export interface NotificationTypeMeta')
    expect(useNotificationDrawerSource).not.toContain('interface NotificationTypeMeta')
  })
})
