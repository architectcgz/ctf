import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'

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

vi.mock('@/features/auth', () => ({
  useLoginPage: () => {
    // eslint-disable-next-line no-console
    console.log(
      '%c[CTF COMMAND CENTER] %cSystem online. Initializing monitoring...',
      'font-weight: bold; font-size: 14px;',
      'font-style: italic;'
    )
    // eslint-disable-next-line no-console
    console.log(
      '%cAudit note: %ccuriosity detected. Keep it academic.',
      'font-weight: bold;',
      ''
    )
    // eslint-disable-next-line no-console
    console.log(
      '%cMemo: %cIf this page were the weak point, we would all be having a worse day.',
      'font-weight: bold;',
      ''
    )

    const form = reactive({ username: '', password: '' })
    const loading = ref(false)
    const submitError = ref('')
    const probeMessage = ref('')

    return {
      form,
      loading,
      submitError,
      probeMessage,
      clearSubmitError: () => {
        submitError.value = ''
      },
      handleHeroProbe: () => {
        probeMessage.value = '隐藏入口排查完毕，结果让你失望了。'
      },
      onSubmit: async (fallbackValues?: { username?: string; password?: string }) => {
        const username = form.username.trim() || fallbackValues?.username?.trim() || ''
        const password = form.password || fallbackValues?.password || ''
        if (loading.value || !username || !password) return
        loading.value = true
        submitError.value = ''
        try {
          await authMocks.login(
            { username, password },
            routeState.query.redirect ? routeState.query.redirect : undefined
          )
        } catch (err) {
          submitError.value = err instanceof Error ? err.message : '身份验证失败，请核对信息'
        } finally {
          loading.value = false
        }
      },
      submitWithFallback: async (fallbackValues?: { username?: string; password?: string }) => {
        const username = form.username.trim() || fallbackValues?.username?.trim() || ''
        const password = form.password || fallbackValues?.password || ''
        if (loading.value || !username || !password) return
        loading.value = true
        submitError.value = ''
        try {
          await authMocks.login(
            { username, password },
            routeState.query.redirect ? routeState.query.redirect : undefined
          )
        } catch (err) {
          submitError.value = err instanceof Error ? err.message : '身份验证失败，请核对信息'
        } finally {
          loading.value = false
        }
      },
    }
  },
}))
vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
}))

describe('LoginView', () => {
  let consoleLogSpy: ReturnType<typeof vi.spyOn>

  afterEach(() => {
    vi.restoreAllMocks()
  })

  beforeEach(() => {
    authMocks.login.mockReset()
    routeState.query.redirect = undefined
    consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
  })

  function mountLoginView() {
    return mount(LoginView)
  }

  it('不应渲染 CAS 登录入口', async () => {
    const wrapper = mountLoginView()

    await flushPromises()

    expect(wrapper.text()).toContain('CTF Platform Infrastructure')
    expect(wrapper.text()).toContain('登录工作台')
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

    expect(wrapper.get('button[type="submit"]').attributes('type')).toBe('submit')
  })

  it('登录表单标签应与输入框建立明确关联', async () => {
    const wrapper = mountLoginView()

    await flushPromises()

    expect(wrapper.get('label[for="login-username"]').text()).toContain('用户名 / 学号')
    expect(wrapper.get('input#login-username').attributes('autocomplete')).toBe('username')
    expect(wrapper.get('label[for="login-password"]').text()).toContain('安全密码')
    expect(wrapper.get('input#login-password').attributes('autocomplete')).toBe(
      'current-password'
    )
  })

  it('登录表单应切到共享控件原语而不是继续使用 Element Plus 表单', () => {
    expect(loginViewSource).toContain('useLoginPage')
    expect(loginViewSource).toContain('submitWithFallback')
    expect(loginViewSource).not.toContain('useLoginViewPage')
    expect(loginViewSource).not.toContain('function submitWithFallback(')
    expect(loginViewSource).toContain('class="ui-control-wrap"')
    expect(loginViewSource).toContain('class="ui-control"')
    expect(loginViewSource).toContain('class="ui-btn ui-btn--primary ui-btn--block auth-submit-btn"')
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

  it('回车触发表单提交重叠时只应登录一次', async () => {
    authMocks.login.mockImplementation(() => new Promise(() => {}))

    const wrapper = mountLoginView()
    await flushPromises()

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="current-password"]')

    await usernameInput.setValue('alice')
    await passwordInput.setValue('saved-password')

    await usernameInput.trigger('keyup.enter')
    await wrapper.get('form').trigger('submit.prevent')

    expect(authMocks.login).toHaveBeenCalledTimes(1)
    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'saved-password' },
      undefined
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
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('应继续输出基础控制台提示并追加审计口吻提示', async () => {
    mountLoginView()

    await flushPromises()

    expect(consoleLogSpy).toHaveBeenCalledWith(
      expect.stringContaining('[CTF COMMAND CENTER]'),
      expect.any(String),
      expect.any(String)
    )
    expect(consoleLogSpy).toHaveBeenCalledWith(
      '%cAudit note: %ccuriosity detected. Keep it academic.',
      expect.any(String),
      expect.any(String)
    )
    expect(consoleLogSpy).toHaveBeenCalledWith(
      '%cMemo: %cIf this page were the weak point, we would all be having a worse day.',
      expect.any(String),
      expect.any(String)
    )
  })

  it('连续点击品牌区后应出现轻提示且不影响表单提交', async () => {
    authMocks.login.mockResolvedValue(undefined)

    const wrapper = mountLoginView()
    await flushPromises()

    expect(wrapper.text()).not.toContain('隐藏入口排查完毕，结果让你失望了。')

    const hero = wrapper.get('.auth-entry-shell__hero')
    await hero.trigger('click')
    await hero.trigger('click')
    await hero.trigger('click')
    await hero.trigger('click')

    expect(wrapper.text()).toContain('隐藏入口排查完毕，结果让你失望了。')

    const usernameInput = wrapper.find('input[autocomplete="username"]')
    const passwordInput = wrapper.find('input[autocomplete="current-password"]')
    await usernameInput.setValue('alice')
    await passwordInput.setValue('saved-password')
    await wrapper.get('form').trigger('submit.prevent')

    expect(authMocks.login).toHaveBeenCalledTimes(1)
    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'saved-password' },
      undefined
    )
  })
})
