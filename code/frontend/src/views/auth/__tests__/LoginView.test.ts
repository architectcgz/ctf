import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'

import LoginView from '@/views/auth/LoginView.vue'

const authMocks = vi.hoisted(() => ({
  login: vi.fn(),
}))

const casMocks = vi.hoisted(() => ({
  casStatusValue: {
    provider: 'cas' as const,
    enabled: true,
    configured: true,
    auto_provision: false,
    login_path: '/api/v1/auth/cas/login',
    callback_path: '/api/v1/auth/cas/callback',
  },
  casRedirectingValue: false,
  fetchCASStatus: vi.fn().mockResolvedValue(undefined),
  beginCASLogin: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('@/composables/useAuth', () => ({
  useAuth: () => authMocks,
}))
vi.mock('@/composables/useCASAuth', async () => {
  const { computed, ref } = await import('vue')
  return {
    useCASAuth: () => ({
      casStatus: ref(casMocks.casStatusValue),
      casLoading: ref(false),
      casReady: computed(() => Boolean(casMocks.casStatusValue.enabled && casMocks.casStatusValue.configured)),
      casRedirecting: ref(casMocks.casRedirectingValue),
      fetchCASStatus: casMocks.fetchCASStatus,
      beginCASLogin: casMocks.beginCASLogin,
    }),
  }
})
vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
  useRoute: () => ({
    query: {
      redirect: '/dashboard',
    },
  }),
}))

describe('LoginView', () => {
  beforeEach(() => {
    authMocks.login.mockReset()
    casMocks.fetchCASStatus.mockClear()
    casMocks.beginCASLogin.mockClear()
  })

  function mountLoginView() {
    const ElInputStub = defineComponent({
      props: ['modelValue', 'type', 'autocomplete', 'showPassword', 'size'],
      emits: ['update:modelValue', 'keyup.enter'],
      setup(props, { emit, expose }) {
        const inputRef = ref<HTMLInputElement | null>(null)
        expose({ input: inputRef })
        return {
          inputRef,
          emitInput: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).value),
          emitEnter: () => emit('keyup.enter'),
        }
      },
      template:
        '<input ref="inputRef" :value="modelValue" :type="type || \'text\'" :data-autocomplete="autocomplete" @input="emitInput" @keyup.enter="emitEnter" />',
    })

    return mount(LoginView, {
      global: {
        stubs: {
          ElForm: { template: '<form @submit.prevent="$emit(\'submit\')"><slot /></form>' },
          ElFormItem: { template: '<label><slot /></label>' },
          ElInput: ElInputStub,
          ElButton: {
            props: ['loading', 'size', 'type', 'disabled'],
            template: '<button @click="$emit(\'click\')"><slot /></button>',
          },
        },
      },
    })
  }

  it('应该渲染 CAS 登录入口并支持触发跳转', async () => {
    const wrapper = mountLoginView()

    await flushPromises()

    expect(casMocks.fetchCASStatus).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('CAS 统一认证')
    expect(wrapper.text()).toContain('使用 CAS 统一认证登录')

    const casButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('使用 CAS 统一认证登录'))
    expect(casButton).toBeTruthy()

    await casButton!.trigger('click')
    expect(casMocks.beginCASLogin).toHaveBeenCalledWith('/dashboard')
  })

  it('用户名输入框按回车时应触发登录', async () => {
    authMocks.login.mockResolvedValue(undefined)

    const wrapper = mountLoginView()
    await flushPromises()

    const usernameInput = wrapper.find('input[data-autocomplete="username"]')
    const passwordInput = wrapper.find('input[data-autocomplete="current-password"]')

    expect(usernameInput.exists()).toBe(true)
    expect(passwordInput.exists()).toBe(true)

    await usernameInput.setValue('alice')
    await passwordInput.setValue('saved-password')
    await usernameInput.trigger('keyup.enter')

    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'saved-password' },
      '/dashboard'
    )
  })

  it('密码由浏览器自动填充时，用户名输入框按回车也应触发登录', async () => {
    authMocks.login.mockResolvedValue(undefined)

    const wrapper = mountLoginView()
    await flushPromises()

    const usernameInput = wrapper.find('input[data-autocomplete="username"]')
    const passwordInput = wrapper.find('input[data-autocomplete="current-password"]')

    expect(usernameInput.exists()).toBe(true)
    expect(passwordInput.exists()).toBe(true)

    await usernameInput.setValue('alice')
    ;(passwordInput.element as HTMLInputElement).value = 'browser-saved-password'
    await usernameInput.trigger('keyup.enter')

    expect(authMocks.login).toHaveBeenCalledWith(
      { username: 'alice', password: 'browser-saved-password' },
      '/dashboard'
    )
  })
})
