import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

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

  it('应该渲染 CAS 登录入口并支持触发跳转', async () => {
    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          ElForm: { template: '<form><slot /></form>' },
          ElFormItem: { template: '<label><slot /></label>' },
          ElInput: {
            props: ['modelValue', 'type', 'autocomplete', 'showPassword', 'size'],
            template: '<div class="el-input-stub" />',
          },
          ElButton: {
            props: ['loading', 'size', 'type', 'disabled'],
            template: '<button @click="$emit(\'click\')"><slot /></button>',
          },
        },
      },
    })

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
})
