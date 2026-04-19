import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import SecuritySettings from '../SecuritySettings.vue'
import securitySettingsSource from '../SecuritySettings.vue?raw'

const authApiMocks = vi.hoisted(() => ({
  changePassword: vi.fn(),
}))

vi.mock('@/api/auth', () => authApiMocks)

describe('SecuritySettings', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()

    authApiMocks.changePassword.mockReset()
    authApiMocks.changePassword.mockResolvedValue(undefined)
  })

  it('应该展示安全设置并支持修改密码', async () => {
    const wrapper = mount(SecuritySettings)
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

  it('应该移除安全设置页中的主题色切换区块', async () => {
    const wrapper = mount(SecuritySettings)
    await flushPromises()

    expect(wrapper.text()).not.toContain('主题色')
    expect(wrapper.text()).not.toContain('绿色')
    expect(wrapper.text()).not.toContain('青色')
    expect(wrapper.text()).not.toContain('蓝色')
  })

  it('应该移除安全设置页级 shell 上遗留的 journal-eyebrow-text 修饰类', () => {
    expect(securitySettingsSource).toContain(
      'class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"'
    )
    expect(securitySettingsSource).not.toContain('journal-eyebrow-text')
  })

  it('应该把安全设置内容区的 soft eyebrow 收敛为局部 section kicker', () => {
    expect(securitySettingsSource).toContain('<div class="security-section-kicker">Password</div>')
    expect(securitySettingsSource).toContain('<div class="security-section-kicker">Tips</div>')
    expect(securitySettingsSource).not.toContain(
      '<div class="journal-eyebrow journal-eyebrow-soft">Password</div>'
    )
    expect(securitySettingsSource).not.toContain(
      '<div class="journal-eyebrow journal-eyebrow-soft">Tips</div>'
    )
  })

  it('应使用共享 ui-control 原语承载密码输入框', () => {
    expect(securitySettingsSource).toContain('class="ui-control-wrap"')
    expect(securitySettingsSource).toContain('class="ui-control"')
    expect(securitySettingsSource).not.toMatch(/^\.journal-input\s*\{/m)
    expect(securitySettingsSource).not.toMatch(/^\.journal-input:focus\s*\{/m)
    expect(securitySettingsSource).not.toMatch(/^\.journal-input--error\s*\{/m)
  })

  it('应把安全提示区的内文字色收敛为语义类', () => {
    expect(securitySettingsSource).toContain('security-side-status')
    expect(securitySettingsSource).toContain('security-side-copy')
    expect(securitySettingsSource).toContain('security-tip-copy')
    expect(securitySettingsSource).not.toContain(
      'class="flex items-center gap-2 text-sm font-medium text-[var(--journal-ink)]"'
    )
    expect(securitySettingsSource).not.toContain(
      'class="mt-3 text-sm leading-6 text-[var(--journal-muted)]"'
    )
    expect(securitySettingsSource).not.toContain(
      'class="mt-2 text-sm leading-6 text-[var(--journal-ink)]"'
    )
  })
})
