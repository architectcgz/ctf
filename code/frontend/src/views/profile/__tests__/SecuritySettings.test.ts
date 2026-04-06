import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import SecuritySettings from '../SecuritySettings.vue'

const authApiMocks = vi.hoisted(() => ({
  changePassword: vi.fn(),
}))

const themeMocks = vi.hoisted(() => ({
  theme: { value: 'dark' },
  brand: { value: 'green' },
  availableBrands: [
    { value: 'green', label: '绿色', description: '更贴近学校 CTF 的默认技术主题' },
    { value: 'cyan', label: '青色', description: '保留当前较冷静的青蓝技术感' },
    { value: 'blue', label: '蓝色', description: '更传统、更中性的控制台色调' },
  ],
  setBrand: vi.fn(),
}))

vi.mock('@/api/auth', () => authApiMocks)
vi.mock('@/composables/useTheme', () => ({
  useTheme: () => themeMocks,
}))

describe('SecuritySettings', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()

    authApiMocks.changePassword.mockReset()
    authApiMocks.changePassword.mockResolvedValue(undefined)
    themeMocks.setBrand.mockReset()
    themeMocks.theme.value = 'dark'
    themeMocks.brand.value = 'green'
  })

  it('应该展示安全设置并支持修改密码', async () => {
    const wrapper = mount(SecuritySettings)
    await flushPromises()

    expect(wrapper.text()).toContain('安全设置')
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

  it('应该展示主题色选项并支持切换品牌主题', async () => {
    const wrapper = mount(SecuritySettings)
    await flushPromises()

    expect(wrapper.text()).toContain('主题色')
    expect(wrapper.text()).toContain('绿色')

    const greenOption = wrapper.find('input[value="green"]')
    expect(greenOption.exists()).toBe(true)

    await greenOption.setValue(true)

    expect(themeMocks.setBrand).toHaveBeenCalledWith('green')
  })
})
