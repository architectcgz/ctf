import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import ContestChallengeOrchestrationPanel from '../contest/ContestChallengeOrchestrationPanel.vue'
import type { AdminContestChallengeData } from '@/api/contracts'

const contestApiMocks = vi.hoisted(() => ({
  listAdminContestChallenges: vi.fn(),
  getChallenges: vi.fn(),
  createAdminContestChallenge: vi.fn(),
  updateAdminContestChallenge: vi.fn(),
  deleteAdminContestChallenge: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    listAdminContestChallenges: contestApiMocks.listAdminContestChallenges,
    getChallenges: contestApiMocks.getChallenges,
    createAdminContestChallenge: contestApiMocks.createAdminContestChallenge,
    updateAdminContestChallenge: contestApiMocks.updateAdminContestChallenge,
    deleteAdminContestChallenge: contestApiMocks.deleteAdminContestChallenge,
  }
})

vi.mock('@/composables/useToast', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: vi.fn(),
}))

function buildChallenge(overrides: Partial<AdminContestChallengeData> = {}): AdminContestChallengeData {
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

const DialogStub = defineComponent({
  name: 'ContestChallengeEditorDialog',
  props: {
    open: { type: Boolean, default: false },
  },
  setup(props) {
    return () => h('div', { 'data-testid': 'contest-challenge-editor-dialog' }, props.open ? 'dialog-open' : '')
  },
})

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
        ContestChallengeEditorDialog: DialogStub,
      },
    },
  })
}

describe('ContestChallengeOrchestrationPanel', () => {
  beforeEach(() => {
    contestApiMocks.listAdminContestChallenges.mockReset()
    contestApiMocks.getChallenges.mockReset()
    contestApiMocks.createAdminContestChallenge.mockReset()
    contestApiMocks.updateAdminContestChallenge.mockReset()
    contestApiMocks.deleteAdminContestChallenge.mockReset()
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

  it('应该在 AWD 模式下显示 checker / SLA / 防守分 / 验证状态摘要列', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge({
        awd_checker_type: 'http_standard',
        awd_sla_score: 18,
        awd_defense_score: 28,
        awd_checker_validation_state: 'stale',
        awd_checker_last_preview_at: '2026-04-12T08:00:00.000Z',
      }),
    ])

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()

    expect(wrapper.text()).toContain('Checker')
    expect(wrapper.text()).toContain('验证状态')
    expect(wrapper.text()).toContain('SLA')
    expect(wrapper.text()).toContain('防守分')
    expect(wrapper.text()).toContain('最近试跑')
    expect(wrapper.text()).toContain('HTTP 标准 Checker')
    expect(wrapper.text()).toContain('SLA 18 / 防守 28')
    expect(wrapper.text()).toContain('待重新验证')
  })

  it('应该支持按未配置 AWD 和预检失败筛选', async () => {
    contestApiMocks.listAdminContestChallenges.mockResolvedValue([
      buildChallenge({
        id: 'link-1',
        challenge_id: '101',
        title: '未配 Checker 题目',
        awd_checker_type: undefined,
        awd_checker_validation_state: 'pending',
      }),
      buildChallenge({
        id: 'link-2',
        challenge_id: '102',
        title: '最近失败题目',
        order: 2,
        awd_checker_type: 'legacy_probe',
        awd_checker_validation_state: 'failed',
      }),
      buildChallenge({
        id: 'link-3',
        challenge_id: '103',
        title: '最近通过题目',
        order: 3,
        awd_checker_type: 'http_standard',
        awd_checker_validation_state: 'passed',
      }),
    ])

    const wrapper = mountPanel({
      contestMode: 'awd',
    })

    await flushPromises()

    await wrapper.get('#contest-challenge-filter-unconfigured').trigger('click')
    expect(wrapper.text()).toContain('未配 Checker 题目')
    expect(wrapper.text()).not.toContain('最近失败题目')
    expect(wrapper.text()).not.toContain('最近通过题目')

    await wrapper.get('#contest-challenge-filter-validation-failed').trigger('click')
    expect(wrapper.text()).not.toContain('未配 Checker 题目')
    expect(wrapper.text()).toContain('最近失败题目')
    expect(wrapper.text()).not.toContain('最近通过题目')
  })

  it('外部题目数据加载失败且尚无成功数据时应展示失败态而不是空态', async () => {
    const wrapper = mountPanel({
      contestMode: 'awd',
      challengeLinks: [],
      loadingExternal: false,
      loadErrorExternal: '同步失败',
    })

    await flushPromises()

    expect(wrapper.text()).toContain('赛事题目暂时不可用')
    expect(wrapper.text()).toContain('同步失败')
    expect(wrapper.text()).not.toContain('当前竞赛还没有关联题目')
    expect(wrapper.text()).not.toContain('共 0 道题目')
  })
})
