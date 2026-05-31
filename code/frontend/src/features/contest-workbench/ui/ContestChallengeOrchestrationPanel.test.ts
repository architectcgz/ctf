import { defineComponent, h } from 'vue'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AdminContestChallengeViewData } from '@/api/contracts'
import ContestChallengeOrchestrationPanel from './ContestChallengeOrchestrationPanel.vue'

const contestApiMocks = vi.hoisted(() => ({
  listAdminContestChallenges: vi.fn(),
  listContestAWDServices: vi.fn(),
  getChallenges: vi.fn(),
  listAdminAwdChallenges: vi.fn(),
  createContestAWDService: vi.fn(),
  createAdminContestChallenge: vi.fn(),
  updateContestAWDService: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
  deleteContestAWDService: vi.fn(),
  deleteAdminContestChallenge: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
}))

vi.mock('@/api/admin/contests', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contests')>('@/api/admin/contests')
  return {
    ...actual,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    listContestAWDServices: contestApiMocks.listContestAWDServices,
    createContestAWDService: contestApiMocks.createContestAWDService,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateContestAWDService: contestApiMocks.updateContestAWDService,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteContestAWDService: contestApiMocks.deleteContestAWDService,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})
vi.mock('@/api/admin/authoring', () => ({
  getChallenges: contestApiMocks.getChallenges,
}))
vi.mock('@/api/admin/awd-authoring', () => ({
  listAdminAwdChallenges: contestApiMocks.listAdminAwdChallenges,
}))

vi.mock('@/shared/model/common/useToast', () => ({
  useToast: () => toastMocks,
}))

vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: vi.fn(),
}))

function buildChallenge(
  overrides: Partial<AdminContestChallengeViewData> = {}
): AdminContestChallengeViewData {
  return {
    id: 'link-1',
    contest_id: 'contest-1',
    challenge_id: '101',
    title: 'Web 入门',
    category: 'web',
    difficulty: 'easy',
    points: 120,
    order: 1,
    is_visible: true,
    awd_checker_type: undefined,
    awd_checker_config: {},
    awd_sla_score: 0,
    awd_defense_score: 0,
    awd_checker_validation_state: 'pending',
    awd_checker_last_preview_at: undefined,
    awd_checker_last_preview_result: undefined,
    created_at: '2026-03-10T00:00:00.000Z',
    ...overrides,
  }
}

function buildAwdService(overrides: Record<string, unknown> = {}) {
  return {
    id: 'service-1',
    contest_id: 'contest-1',
    awd_challenge_id: '11',
    display_name: 'Web 入门',
    order: 1,
    is_visible: true,
    score_config: {
      points: 120,
      awd_sla_score: 1,
      awd_defense_score: 2,
    },
    runtime_config: {
      checker_type: 'http_standard',
      checker_config: {},
    },
    checker_type: 'http_standard',
    checker_config: {},
    sla_score: 18,
    defense_score: 28,
    validation_state: 'stale',
    last_preview_at: '2026-04-12T08:00:00.000Z',
    last_preview_result: undefined,
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T00:00:00.000Z',
    ...overrides,
  }
}

function mountPanel(props?: Record<string, unknown>) {
  return mount(ContestChallengeOrchestrationPanel, {
    props: {
      contestId: 'contest-1',
      contestMode: 'jeopardy',
      ...props,
    },
    global: {
      stubs: {
        AppEmpty: {
          props: ['title', 'description'],
          template: '<div class="app-empty-stub">{{ title }}|{{ description }}</div>',
        },
        AppLoading: {
          template: '<div class="app-loading-stub"><slot /></div>',
        },
        AdminSurfaceModal: defineComponent({
          name: 'AdminSurfaceModal',
          props: {
            open: { type: Boolean, default: false },
          },
          setup(props, { slots }) {
            return () =>
              props.open
                ? h('div', { class: 'admin-surface-modal-stub' }, [
                    h('div', { class: 'admin-surface-modal-stub__body' }, slots.default?.()),
                    h('div', { class: 'admin-surface-modal-stub__footer' }, slots.footer?.()),
                  ])
                : null
          },
        }),
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('ContestChallengeOrchestrationPanel', () => {
  beforeEach(() => {
    contestApiMocks.listAdminContestChallenges.mockReset()
    contestApiMocks.listContestAWDServices.mockReset()
    contestApiMocks.getChallenges.mockReset()
    contestApiMocks.listAdminAwdChallenges.mockReset()
    contestApiMocks.createContestAWDService.mockReset()
    contestApiMocks.createAdminContestChallenge.mockReset()
    contestApiMocks.updateContestAWDService.mockReset()
    contestApiMocks.updateAdminContestChallenge.mockReset()
    contestApiMocks.deleteContestAWDService.mockReset()
    contestApiMocks.deleteAdminContestChallenge.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    contestApiMocks.listContestAWDServices.mockResolvedValue([])
  })

  it('应该显示基础编排字段', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge({
        id: 'link-2',
        challenge_id: '102',
        title: 'Crypto 进阶',
        order: 2,
        is_visible: false,
        points: 150,
      }),
      buildChallenge(),
    ])

    const wrapper = mountPanel()

    await flushPromises()

    expect(wrapper.text()).toContain('题目池')
    expect(wrapper.text()).toContain('题目')
    expect(wrapper.text()).toContain('可见性')
    expect(wrapper.text()).toContain('分值')
    expect(wrapper.text()).toContain('顺序')
    expect(wrapper.text()).toContain('Web 入门')
    expect(wrapper.text()).toContain('第 1 位')
  })

  it('应该把题目编排表的题目标题链接到管理员题目详情页', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([buildChallenge()])

    const wrapper = mountPanel()

    await flushPromises()

    const titleLink = wrapper
      .findAllComponents(RouterLinkStub)
      .find((link) => link.text() === 'Web 入门')

    expect(titleLink?.props('to')).toEqual({
      name: 'PlatformChallengeDetail',
      params: { id: '101' },
    })
  })
})
