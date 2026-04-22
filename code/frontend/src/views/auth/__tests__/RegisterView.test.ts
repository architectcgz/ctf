import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import RegisterView from '@/views/auth/RegisterView.vue'
import registerViewSource from '@/views/auth/RegisterView.vue?raw'

const authMocks = vi.hoisted(() => ({
  register: vi.fn(),
}))

vi.mock('@/composables/useAuth', () => ({
  useAuth: () => authMocks,
}))

vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
}))

describe('RegisterView', () => {
  beforeEach(() => {
    authMocks.register.mockReset()
  })

  function mountRegisterView() {
    return mount(RegisterView)
  }

  it('应该渲染统一认证壳层和注册表单', async () => {
    const wrapper = mountRegisterView()

    await flushPromises()

    expect(wrapper.text()).toContain('CTF Platform Infrastructure')
    expect(wrapper.text()).toContain('训练空间')
    expect(wrapper.text()).toContain('教学协同')
    expect(wrapper.text()).toContain('系统值守')
    expect(wrapper.text()).toContain('注册账号')
    expect(wrapper.text()).toContain('已经有账号了')
    expect(wrapper.text()).toContain('返回登录')
    expect(wrapper.findAll('input')).toHaveLength(3)
  })

  it('填写必要字段后应触发注册', async () => {
    authMocks.register.mockResolvedValue(undefined)

    const wrapper = mountRegisterView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="new-password"]')
    const classNameInput = wrapper.findAll('input').at(2)

    expect(usernameInput.exists()).toBe(true)
    expect(passwordInput.exists()).toBe(true)
    expect(classNameInput?.exists()).toBe(true)

    await usernameInput.setValue('alice')
    await passwordInput.setValue('secure-pass')
    await classNameInput!.setValue('CTF-1')
    await wrapper.find('form').trigger('submit.prevent')

    expect(authMocks.register).toHaveBeenCalledWith({
      username: 'alice',
      password: 'secure-pass',
      class_name: 'CTF-1',
    })
  })

  it('注册按钮应使用原生 submit 类型以支持表单回车提交', async () => {
    const wrapper = mountRegisterView()

    await flushPromises()

    expect(wrapper.get('button[type="submit"]').attributes('type')).toBe('submit')
  })

  it('注册失败时应停留在当前页并展示错误信息', async () => {
    authMocks.register.mockRejectedValue(new Error('用户名已存在'))

    const wrapper = mountRegisterView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="new-password"]')

    await usernameInput.setValue('alice')
    await passwordInput.setValue('secure-pass')

    await expect(wrapper.get('form').trigger('submit.prevent')).resolves.toBeUndefined()
    await flushPromises()

    expect(authMocks.register).toHaveBeenCalledWith({
      username: 'alice',
      password: 'secure-pass',
      class_name: undefined,
    })
    expect(wrapper.text()).toContain('用户名已存在')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('注册进行中重复提交时只应发起一次请求', async () => {
    authMocks.register.mockImplementation(() => new Promise(() => {}))

    const wrapper = mountRegisterView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="new-password"]')

    await usernameInput.setValue('alice')
    await passwordInput.setValue('secure-pass')

    await wrapper.get('form').trigger('submit.prevent')
    await wrapper.get('form').trigger('submit.prevent')

    expect(authMocks.register).toHaveBeenCalledTimes(1)
    expect(authMocks.register).toHaveBeenCalledWith({
      username: 'alice',
      password: 'secure-pass',
      class_name: undefined,
    })
  })

  it('注册表单应切到共享控件原语而不是继续使用 Element Plus 表单', () => {
    expect(registerViewSource).toContain('class="ui-control-wrap"')
    expect(registerViewSource).toContain('class="ui-control"')
    expect(registerViewSource).toContain('class="ui-btn ui-btn--primary ui-btn--block auth-register-submit"')
    expect(registerViewSource).not.toContain('<ElForm')
    expect(registerViewSource).not.toContain('<ElFormItem')
    expect(registerViewSource).not.toContain('<ElInput')
    expect(registerViewSource).not.toContain('<ElButton')
  })
})
