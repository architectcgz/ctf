import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import LoginView from '@/views/auth/LoginView.vue'
import loginViewSource from '@/views/auth/LoginView.vue?raw'

const authMocks = vi.hoisted(() => ({
  login: vi.fn(),
}))
const routeState = vi.hoisted(() => ({
  query: {
    redirect: undefined as string | undefined,
  },
}))

vi.mock('@/composables/useAuth', () => ({
  useAuth: () => authMocks,
}))
vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
  useRoute: () => routeState,
}))

describe('LoginView', () => {
  beforeEach(() => {
    authMocks.login.mockReset()
    routeState.query.redirect = undefined
  })

  function mountLoginView() {
    return mount(LoginView)
  }

  it('不应渲染 CAS 登录入口', async () => {
    const wrapper = mountLoginView()

    await flushPromises()

    expect(wrapper.text()).toContain('教学平台入口')
    expect(wrapper.text()).toContain('训练空间')
    expect(wrapper.text()).toContain('教学协同')
    expect(wrapper.text()).toContain('系统值守')
    expect(wrapper.text()).not.toContain('CAS 统一认证')
    expect(wrapper.text()).not.toContain('使用 CAS 统一认证登录')
  })

  it('用户名输入框按回车时应触发登录', async () => {
    authMocks.login.mockResolvedValue(undefined)

    const wrapper = mountLoginView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="current-password"]')

    expect(usernameInput.exists()).toBe(true)
    expect(passwordInput.exists()).toBe(true)

    await usernameInput.setValue('alice')
    await passwordInput.setValue('saved-password')
    await usernameInput.trigger('keyup.enter')

    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'saved-password' },
      undefined
    )
  })

  it('登录按钮应使用原生 submit 类型以支持表单回车提交', async () => {
    const wrapper = mountLoginView()

    await flushPromises()

    expect(wrapper.get('button').attributes('type')).toBe('submit')
  })

  it('登录表单应切到共享控件原语而不是继续使用 Element Plus 表单', () => {
    expect(loginViewSource).toContain('class="ui-control-wrap"')
    expect(loginViewSource).toContain('class="ui-control"')
    expect(loginViewSource).toContain(
      'class="ui-btn ui-btn--primary ui-btn--block auth-login-form__submit"'
    )
    expect(loginViewSource).not.toContain('<ElForm')
    expect(loginViewSource).not.toContain('<ElFormItem')
    expect(loginViewSource).not.toContain('<ElInput')
    expect(loginViewSource).not.toContain('<ElButton')
  })

  it('密码由浏览器自动填充时，用户名输入框按回车也应触发登录', async () => {
    authMocks.login.mockResolvedValue(undefined)

    const wrapper = mountLoginView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="current-password"]')

    expect(usernameInput.exists()).toBe(true)
    expect(passwordInput.exists()).toBe(true)

    await usernameInput.setValue('alice')
    ;(passwordInput.element as HTMLInputElement).value = 'browser-saved-password'
    await usernameInput.trigger('keyup.enter')

    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'browser-saved-password' },
      undefined
    )
  })

  it('携带 redirect 参数时应继续把目标路径传给登录逻辑', async () => {
    routeState.query.redirect = '/teacher/dashboard'
    authMocks.login.mockResolvedValue(undefined)

    const wrapper = mountLoginView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="current-password"]')

    await usernameInput.setValue('alice')
    await passwordInput.setValue('saved-password')
    await usernameInput.trigger('keyup.enter')

    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'saved-password' },
      '/teacher/dashboard'
    )
  })

  it('登录失败时应停留在当前页并展示错误信息', async () => {
    authMocks.login.mockRejectedValue(new Error('用户名或密码错误'))

    const wrapper = mountLoginView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="current-password"]')

    await usernameInput.setValue('alice')
    await passwordInput.setValue('wrong-password')

    await expect(wrapper.get('form').trigger('submit.prevent')).resolves.toBeUndefined()
    await flushPromises()

    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'wrong-password' },
      undefined
    )
    expect(wrapper.text()).toContain('用户名或密码错误')
    expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
  })
})
