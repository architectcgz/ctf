import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeDetail from '../ChallengeDetail.vue'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: '11' } as Record<string, string>,
  query: {} as Record<string, string>,
}))

const adminApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn(),
  configureChallengeFlag: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock, back: vi.fn() }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('Admin ChallengeDetail', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    routeState.params = { id: '11' }
    routeState.query = {}
    adminApiMocks.getChallengeDetail.mockReset()
    adminApiMocks.configureChallengeFlag.mockReset()
    adminApiMocks.getChallengeDetail.mockResolvedValue({
      id: '11',
      title: '双节点演练',
      category: 'web',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      image_id: 'img-1',
      attachment_url: 'https://example.com/demo.zip',
      description: 'desc',
      hints: [{ id: 'hint-1', level: 1, title: '入口', content: '观察回显' }],
      flag_config: {
        configured: true,
        flag_type: 'static',
      },
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T00:00:00.000Z',
    })
  })

  it('应该默认显示题目管理 tab，并保留独立的拓扑编排入口', async () => {
    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div data-testid="challenge-writeup-manage-panel">题解目录</div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('题目管理')
    expect(wrapper.text()).toContain('题解管理')
    expect(wrapper.find('#admin-challenge-tab-detail').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#admin-challenge-panel-detail').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#admin-challenge-panel-writeup').attributes('aria-hidden')).toBe('true')
    expect(wrapper.text()).toContain('双节点演练')

    const topologyButton = wrapper.findAll('button').find((button) => button.text().includes('拓扑编排'))
    expect(topologyButton).toBeTruthy()

    await topologyButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith('/platform/challenges/11/topology')
  })

  it('应该根据 query 切到题解管理 tab', async () => {
    routeState.query = { panel: 'writeup' }

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div data-testid="challenge-writeup-tab">题解目录</div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('#admin-challenge-tab-writeup').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#admin-challenge-panel-writeup').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('[data-testid="challenge-writeup-tab"]').exists()).toBe(true)
  })

  it('切换题解管理 tab 时应同步更新 panel query', async () => {
    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#admin-challenge-tab-writeup').trigger('click')

    expect(replaceMock).toHaveBeenCalledWith({
      name: 'AdminChallengeDetail',
      params: { id: '11' },
      query: { panel: 'writeup' },
    })
  })

  it('共享实例题应明确提示答案不做用户隔离', async () => {
    adminApiMocks.getChallengeDetail.mockResolvedValue({
      id: '11',
      title: '共享密码题',
      category: 'crypto',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      image_id: 'img-1',
      description: 'desc',
      instance_sharing: 'shared',
      flag_config: {
        configured: true,
        flag_type: 'static',
      },
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T00:00:00.000Z',
    })

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('共享实例只适用于无状态题')
    expect(wrapper.text()).toContain('不提供用户级答案隔离')
    expect(wrapper.text()).toContain('若需隔离答案，请使用 per_user 或 per_team')
  })

  it('共享实例题不应允许保存动态 Flag', async () => {
    adminApiMocks.getChallengeDetail.mockResolvedValue({
      id: '11',
      title: '共享密码题',
      category: 'crypto',
      difficulty: 'easy',
      status: 'draft',
      points: 100,
      image_id: 'img-1',
      description: 'desc',
      instance_sharing: 'shared',
      flag_config: {
        configured: true,
        flag_type: 'static',
      },
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T00:00:00.000Z',
    })

    const wrapper = mount(ChallengeDetail, {
      global: {
        stubs: {
          ChallengeDescriptionPanel: { template: '<div>描述面板</div>' },
          ChallengeWriteupManagePanel: { template: '<div>题解目录</div>' },
        },
      },
    })

    await flushPromises()

    await wrapper.get('select.flag-field-input').setValue('dynamic')
    const saveButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('保存配置'))
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')

    expect(adminApiMocks.configureChallengeFlag).not.toHaveBeenCalled()
    expect(toastMocks.error).toHaveBeenCalledWith(
      '共享实例只适用于无状态题，不支持动态 Flag；若需隔离答案，请使用 per_user 或 per_team'
    )
  })
})
