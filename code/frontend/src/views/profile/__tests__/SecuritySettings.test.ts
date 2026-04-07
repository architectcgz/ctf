import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import SecuritySettings from '../SecuritySettings.vue'

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
})
