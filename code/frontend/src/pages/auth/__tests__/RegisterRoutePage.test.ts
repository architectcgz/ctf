import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'

import RegisterRoutePage from '@/pages/auth/RegisterRoutePage.vue'
import registerRoutePageSource from '@/pages/auth/RegisterRoutePage.vue?raw'
import registerPageModelSource from '@/features/auth/model/useRegisterPage.ts?raw'

const authMocks = vi.hoisted(() => ({
  register: vi.fn(),
}))
const pushMock = vi.hoisted(() => vi.fn())

vi.mock('@/features/auth', () => ({
  AuthEntryShell: {
    props: ['panelEyebrow', 'panelTitle', 'panelDescription'],
    template: `
      <div class="auth-entry-shell">
        <section class="auth-entry-shell__hero">
          <div>CTF Platform Infrastructure</div>
          <div>训练空间</div>
          <div>教学协同</div>
          <div>系统值守</div>
        </section>
        <main>
          <div>{{ panelEyebrow }}</div>
          <h2>{{ panelTitle }}</h2>
          <p>{{ panelDescription }}</p>
          <slot />
          <slot name="footer" />
        </main>
      </div>
    `,
  },
  useRegisterPage: () => {
    const form = reactive({ username: '', password: '', class_name: '' })
    const loading = ref(false)
    const submitError = ref('')
    const successRedirectTo = ref<string | null>(null)

    return {
      form,
      loading,
      submitError,
      successRedirectTo,
      clearSubmitError: () => {
        submitError.value = ''
      },
      onSubmit: async () => {
        if (loading.value || !form.username || !form.password) return
        loading.value = true
        submitError.value = ''
        try {
          await authMocks.register({
            username: form.username,
            password: form.password,
            class_name: form.class_name.trim() || undefined,
          })
          successRedirectTo.value = '/student/dashboard'
        } catch (err) {
          submitError.value = err instanceof Error ? err.message : '注册失败，请稍后重试'
        } finally {
          loading.value = false
        }
      },
    }
  },
}))

vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
  useRouter: () => ({ push: pushMock, replace: vi.fn() }),
}))

describe('RegisterRoutePage', () => {
  beforeEach(() => {
    authMocks.register.mockReset()
    pushMock.mockReset()
  })

  function mountRegisterRoutePage() {
    return mount(RegisterRoutePage)
  }

  it('应该渲染统一认证壳层和注册表单', async () => {
    const wrapper = mountRegisterRoutePage()

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

    const wrapper = mountRegisterRoutePage()
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
    await flushPromises()

    expect(authMocks.register).toHaveBeenCalledWith({
      username: 'alice',
      password: 'secure-pass',
      class_name: 'CTF-1',
    })
    expect(pushMock).toHaveBeenCalledWith('/student/dashboard')
  })

  it('注册按钮应使用原生 submit 类型以支持表单回车提交', async () => {
    const wrapper = mountRegisterRoutePage()

    await flushPromises()

    expect(wrapper.get('button[type="submit"]').attributes('type')).toBe('submit')
  })

  it('注册表单标签应与输入框建立明确关联', async () => {
    const wrapper = mountRegisterRoutePage()

    await flushPromises()

    expect(wrapper.get('label[for="register-username"]').text()).toContain('用户名')
    expect(wrapper.get('input#register-username').attributes('autocomplete')).toBe('username')
    expect(wrapper.get('label[for="register-password"]').text()).toContain('设置密码')
    expect(wrapper.get('input#register-password').attributes('autocomplete')).toBe('new-password')
    expect(wrapper.get('label[for="register-class-name"]').text()).toContain('班级邀请码')
    expect(wrapper.find('input#register-class-name').exists()).toBe(true)
  })

  it('注册失败时应停留在当前页并展示错误信息', async () => {
    authMocks.register.mockRejectedValue(new Error('用户名已存在'))

    const wrapper = mountRegisterRoutePage()
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
    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('用户名已存在')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('注册进行中重复提交时只应发起一次请求', async () => {
    authMocks.register.mockImplementation(() => new Promise(() => {}))

    const wrapper = mountRegisterRoutePage()
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
    expect(registerRoutePageSource).toContain(
      "import AppRouteRedirect from '@/components/navigation/AppRouteRedirect.vue'"
    )
    expect(registerRoutePageSource).toContain('useRegisterPage')
    expect(registerRoutePageSource).toContain('class="ui-control-wrap"')
    expect(registerRoutePageSource).toContain('class="ui-control"')
    expect(registerRoutePageSource).toContain(
      'class="ui-btn ui-btn--primary ui-btn--block auth-register-submit"'
    )
    expect(registerRoutePageSource).not.toContain('<ElForm')
    expect(registerRoutePageSource).not.toContain('<ElFormItem')
    expect(registerRoutePageSource).not.toContain('<ElInput')
    expect(registerRoutePageSource).not.toContain('<ElButton')
    expect(registerPageModelSource).not.toContain("from 'vue-router'")
  })
})
