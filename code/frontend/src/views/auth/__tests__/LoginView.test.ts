import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'

import LoginView from '@/views/auth/LoginView.vue'

const authMocks = vi.hoisted(() => ({
  login: vi.fn(),
}))

vi.mock('@/composables/useAuth', () => ({
  useAuth: () => authMocks,
}))
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
