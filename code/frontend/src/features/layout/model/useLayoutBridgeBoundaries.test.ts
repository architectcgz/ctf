import { describe, expect, it } from 'vitest'

import sessionBridgeSource from './useLayoutSessionActionsBridge.ts?raw'
import drawerBridgeSource from './useLayoutNotificationDrawerBridge.ts?raw'
import realtimeBridgeSource from './useLayoutNotificationRealtimeBridge.ts?raw'

describe('layout bridge boundaries', () => {
  describe('vue-router isolation', () => {
    it('session actions bridge 不应直接 import vue-router，登出后导航应通过回调注入', () => {
      expect(sessionBridgeSource).not.toMatch(/from\s+['"]vue-router['"]/)
      expect(sessionBridgeSource).not.toMatch(/\buseRouter\s*\(/)
      expect(sessionBridgeSource).not.toMatch(/\brouter\.push\s*\(/)
      // 确认回调注入接口存在
      expect(sessionBridgeSource).toContain('navigateToLogin')
      expect(sessionBridgeSource).toContain('export function useLayoutSessionActionsBridge')
    })

    it('notification drawer bridge 不应直接 import vue-router，通知导航应通过回调注入', () => {
      expect(drawerBridgeSource).not.toMatch(/from\s+['"]vue-router['"]/)
      expect(drawerBridgeSource).not.toMatch(/\buseRouter\s*\(/)
      expect(drawerBridgeSource).not.toMatch(/\brouter\.push\s*\(/)
      // 确认回调注入接口存在
      expect(drawerBridgeSource).toContain('goToNotifications')
      expect(drawerBridgeSource).toContain('goToNotificationDetail')
      expect(drawerBridgeSource).toContain('export function useLayoutNotificationDrawerBridge')
    })

    it('notification realtime bridge 不应直接 import vue-router', () => {
      expect(realtimeBridgeSource).not.toMatch(/from\s+['"]vue-router['"]/)
      expect(realtimeBridgeSource).toContain('export function useLayoutNotificationRealtimeBridge')
    })
  })

  describe('owner boundaries', () => {
    it('session actions bridge 应只桥接 auth feature，不直接调用 auth API', () => {
      expect(sessionBridgeSource).toContain("import { useAuth } from '@/features/auth'")
      expect(sessionBridgeSource).toContain('const { logout } = useAuth()')
      // 不应直接调 API
      expect(sessionBridgeSource).not.toMatch(/from\s+['"]@\/api\/auth['"]/)
      expect(sessionBridgeSource).not.toMatch(/\blogoutApi\b/)
    })

    it('notification drawer bridge 应只桥接 notification feature，不直接调 notification API', () => {
      expect(drawerBridgeSource).toContain("import { useNotificationDrawer } from '@/features/notifications'")
      // 不应直接调 API
      expect(drawerBridgeSource).not.toMatch(/from\s+['"]@\/api\/notification['"]/)
    })

    it('notification realtime bridge 应只桥接 notification feature', () => {
      expect(realtimeBridgeSource).toContain("import { useNotificationRealtime } from '@/features/notifications'")
    })
  })

  describe('layer isolation', () => {
    it('所有 bridge 不应 import pages 或 widgets 层', () => {
      const allSources = [sessionBridgeSource, drawerBridgeSource, realtimeBridgeSource]
      for (const source of allSources) {
        expect(source).not.toMatch(/from\s+['"]@\/pages\//)
        expect(source).not.toMatch(/from\s+['"]@\/widgets\//)
      }
    })

    it('bridge 不应渲染 Vue 组件或包含模板', () => {
      const allSources = [sessionBridgeSource, drawerBridgeSource, realtimeBridgeSource]
      for (const source of allSources) {
        expect(source).not.toContain('.vue')
        expect(source).not.toContain('defineComponent')
        expect(source).not.toContain('<template>')
      }
    })
  })
})
