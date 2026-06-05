import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { flushPromises, mount } from '@vue/test-utils'

import SecuritySettings from '@/pages/profile/SecuritySettingsRoutePage.vue'
import securitySettingsSource from '@/pages/profile/SecuritySettingsRoutePage.vue?raw'
import securitySettingsWorkspaceShellSource from '@/features/profile/ui/SecuritySettingsWorkspaceShell.vue?raw'

const authApiMocks = vi.hoisted(() => ({
  changePassword: vi.fn(),
}))

vi.mock('@/api/auth', () => authApiMocks)

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/login',
        component: { template: '<div>login</div>' },
      },
      {
        path: '/settings/security',
        component: SecuritySettings,
      },
    ],
  })
}

function mountSecuritySettings() {
  const router = createTestRouter()
  return {
    router,
    wrapper: mount(SecuritySettings, {
      global: {
        plugins: [router],
      },
    }),
  }
}

describe('SecuritySettings', () => {
  const securitySettingsWorkspaceSource = [
    securitySettingsSource,
    securitySettingsWorkspaceShellSource,
  ].join('\n')

  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()

    authApiMocks.changePassword.mockReset()
    authApiMocks.changePassword.mockResolvedValue(undefined)
  })

  it('应该展示安全设置并支持修改密码', async () => {
    const { wrapper } = mountSecuritySettings()
    await flushPromises()

    expect(wrapper.text()).toContain('安全设置')
    expect(wrapper.text()).toContain('更新账号密码并检查当前安全策略。')
    expect(wrapper.get('h1').classes()).toContain('workspace-page-title')
    expect(wrapper.find('.workspace-page-copy').exists()).toBe(true)
    expect(wrapper.find('.security-topbar').exists()).toBe(true)
    expect(wrapper.find('.security-summary').exists()).toBe(true)
    expect(wrapper.find('.security-summary-title').text()).toContain('安全概况')
    expect(wrapper.text()).toContain('密码修改')

    const passwordInputs = wrapper.findAll('input[type="password"]')
    await passwordInputs[0].setValue('Password123')
    await passwordInputs[1].setValue('Password456')
    await passwordInputs[2].setValue('Password456')

    const submitButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('更新密码'))
    expect(submitButton).toBeTruthy()

    await submitButton!.trigger('click')
    await flushPromises()

    expect(authApiMocks.changePassword).toHaveBeenCalledWith({
      old_password: 'Password123',
      new_password: 'Password456',
    })
  })

  it('改密成功后应清除认证状态并跳转登录页', async () => {
    const { wrapper, router } = mountSecuritySettings()
    await flushPromises()
    await router.isReady()

    const passwordInputs = wrapper.findAll('input[type="password"]')
    await passwordInputs[0].setValue('Password123')
    await passwordInputs[1].setValue('Password456')
    await passwordInputs[2].setValue('Password456')

    const submitButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('更新密码'))
    expect(submitButton).toBeTruthy()

    const pushSpy = vi.spyOn(router, 'push')
    await submitButton!.trigger('click')
    await flushPromises()

    expect(authApiMocks.changePassword).toHaveBeenCalledTimes(1)
    expect(pushSpy).toHaveBeenCalledWith('/login')
  })

  it('路由页应仅负责组合，不直接耦合密码提交与校验流程', () => {
    expect(securitySettingsSource).toContain('useSecuritySettingsPage')
    expect(securitySettingsSource).not.toContain("from '@/api/auth'")
    expect(securitySettingsSource).not.toContain('validatePasswordForm')
    expect(securitySettingsSource).toContain("from '@/features/profile'")
    expect(securitySettingsSource).toContain('SecuritySettingsWorkspaceShell')
    expect(securitySettingsWorkspaceSource).not.toContain('<PageHeader')
    expect(securitySettingsWorkspaceSource).toContain('安全概况')
    expect(securitySettingsWorkspaceSource).toContain('submitPasswordChange')
    expect(securitySettingsWorkspaceSource).toContain('<component :is="stat.icon" class="h-4 w-4" />')
  })

  it('密码修改进行中重复提交表单时只应提交一次', async () => {
    authApiMocks.changePassword.mockImplementation(() => new Promise(() => {}))

    const { wrapper } = mountSecuritySettings()
    await flushPromises()

    const passwordInputs = wrapper.findAll('input[type="password"]')
    await passwordInputs[0].setValue('Password123')
    await passwordInputs[1].setValue('Password456')
    await passwordInputs[2].setValue('Password456')

    const submitButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('更新密码'))
    expect(submitButton).toBeTruthy()

    await wrapper.get('form').trigger('submit.prevent')
    await wrapper.get('form').trigger('submit.prevent')

    expect(authApiMocks.changePassword).toHaveBeenCalledTimes(1)
    expect(authApiMocks.changePassword).toHaveBeenCalledWith({
      old_password: 'Password123',
      new_password: 'Password456',
    })
  })

  it('应该移除安全设置页中的主题色切换区块', async () => {
    const { wrapper } = mountSecuritySettings()
    await flushPromises()

    expect(wrapper.text()).not.toContain('主题色')
    expect(wrapper.text()).not.toContain('绿色')
    expect(wrapper.text()).not.toContain('青色')
    expect(wrapper.text()).not.toContain('蓝色')
  })
})
