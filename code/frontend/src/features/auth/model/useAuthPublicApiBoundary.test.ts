import { describe, expect, it } from 'vitest'

import authModelIndexSource from './index.ts?raw'
import useAuthSource from './useAuth.ts?raw'

describe('auth feature public API boundaries', () => {
  it('model/index.ts 应通过 barrel export 暴露 useAuth', () => {
    expect(authModelIndexSource).toContain("export * from './useAuth'")
  })

  it('useAuth 应暴露 login / register / logout 三个方法', () => {
    expect(useAuthSource).toContain('return { login, register, logout }')
  })

  it('useAuth 应通过 authStore 管理状态，不应绕过 store 直接写 localStorage 等副作用', () => {
    expect(useAuthSource).toContain('authStore.setAuth(')
    expect(useAuthSource).toContain('authStore.logout()')
  })

  it('useAuth 不应直接 import vue-router，导航应由调用方通过回调处理', () => {
    expect(useAuthSource).not.toMatch(/from\s+['"]vue-router['"]/)
    expect(useAuthSource).not.toMatch(/\buseRouter\s*\(/)
    expect(useAuthSource).not.toMatch(/\brouter\.push\s*\(/)
  })

  it('logout 不应抛出让调用方崩溃的异常', () => {
    expect(useAuthSource).toContain('try {')
    expect(useAuthSource).toContain('} catch {')
    expect(useAuthSource).toContain('authStore.logout()')
  })
})
