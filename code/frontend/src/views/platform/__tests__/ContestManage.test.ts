import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestManage from '../ContestManage.vue'
import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'
import { ApiError } from '@/api/request'

const pushMock = vi.fn()

const contestMocks = vi.hoisted(() => ({
  getContests: vi.fn(),
  getChallenges: vi.fn(),
  createContest: vi.fn(),
  updateContest: vi.fn(),
  getContestAWDReadiness: vi.fn(),
}))
const destructiveConfirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContests: contestMocks.getContests,
    getChallenges: contestMocks.getChallenges,
    createContest: contestMocks.createContest,
    updateContest: contestMocks.updateContest,
    getContestAWDReadiness: contestMocks.getContestAWDReadiness,
  }
})

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock, replace: vi.fn(), back: vi.fn() }),
  }
})

vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: destructiveConfirmMock,
}))

describe('ContestManage', () => {
  beforeEach(() => {
    pushMock.mockReset()
    contestMocks.getContests.mockReset()
    contestMocks.getChallenges.mockReset()
    contestMocks.createContest.mockReset()
    contestMocks.updateContest.mockReset()
    contestMocks.getContestAWDReadiness.mockReset()
    destructiveConfirmMock.mockReset()

    contestMocks.getChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
    contestMocks.getContestAWDReadiness.mockResolvedValue({
      contest_id: 'awd-start',
      ready: false,
      total_challenges: 1,
      passed_challenges: 0,
      pending_challenges: 0,
      failed_challenges: 1,
      stale_challenges: 0,
      missing_checker_challenges: 0,
      blocking_count: 1,
      global_blocking_reasons: [],
      blocking_actions: ['start_contest'],
      items: [
        {
          challenge_id: '101',
          title: 'Challenge 101',
          checker_type: 'http_standard',
          validation_state: 'failed',
          last_preview_at: '2026-04-12T08:00:00.000Z',
          last_access_url: 'http://checker.internal/flag',
          blocking_reason: 'last_preview_failed',
        },
      ],
    })
  })

  it('应该在 AWD 赛事启动被 gate 拦截后拉取 readiness 并允许强制继续', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-start',
          title: '2026 AWD 联赛',
          description: '开赛门禁',
          mode: 'awd',
          status: 'registering',
          starts_at: '2026-04-12T09:00:00.000Z',
          ends_at: '2026-04-12T18:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.updateContest
      .mockRejectedValueOnce(new ApiError('AWD 开赛就绪检查未通过', { status: 409, code: 14025 }))
      .mockResolvedValueOnce({
        contest: {
          id: 'awd-start',
          title: '2026 AWD 联赛',
          description: '开赛门禁',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-04-12T09:00:00.000Z',
          ends_at: '2026-04-12T18:00:00.000Z',
        },
      })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ContestOrchestrationPage: {
            props: ['list'],
            template:
              '<div><button id="open-awd-edit" type="button" @click="$emit(\'openEditDialog\', list[0])">编辑 AWD</button></div>',
          },
          PlatformContestFormDialog: {
            props: ['open', 'draft'],
            template:
              '<div><button v-if="open" id="submit-awd-contest" type="button" @click="$emit(\'save\', { ...draft, mode: \'awd\', status: \'running\' })">保存 AWD</button></div>',
          },
          AWDReadinessOverrideDialog: {
            props: ['open', 'title'],
            data() {
              return { reason: '' }
            },
            template:
              '<div v-if="open"><div>{{ title }}</div><textarea id="awd-readiness-override-reason" v-model="reason" /><button id="awd-readiness-override-submit" type="button" @click="$emit(\'confirm\', reason)">强制继续</button></div>',
          },
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#open-awd-edit').trigger('click')
    await flushPromises()
    await wrapper.get('#submit-awd-contest').trigger('click')
    await flushPromises()

    expect(contestMocks.updateContest).toHaveBeenNthCalledWith(
      1,
      'awd-start',
      expect.objectContaining({
        status: 'running',
      }),
      { suppressErrorToast: true }
    )
    expect(contestMocks.getContestAWDReadiness).toHaveBeenCalledWith('awd-start')
    expect(wrapper.text()).toContain('启动赛事')
    expect(wrapper.text()).toContain('强制继续')

    await wrapper.get('#awd-readiness-override-reason').setValue('teacher drill')
    await wrapper.get('#awd-readiness-override-submit').trigger('click')
    await flushPromises()

    expect(contestMocks.updateContest).toHaveBeenNthCalledWith(
      2,
      'awd-start',
      expect.objectContaining({
        status: 'running',
        force_override: true,
        override_reason: 'teacher drill',
      }),
      { suppressErrorToast: true }
    )
  })

  it('应该忽略普通 409，不误打开启动赛事 gate 弹层', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-conflict',
          title: '2026 AWD 联赛',
          description: '普通冲突',
          mode: 'awd',
          status: 'registering',
          starts_at: '2026-04-12T09:00:00.000Z',
          ends_at: '2026-04-12T18:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.updateContest.mockRejectedValueOnce(
      new ApiError('普通冲突', { status: 409, code: 14099 })
    )

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ContestOrchestrationPage: {
            props: ['list'],
            template:
              '<div><button id="open-awd-edit" type="button" @click="$emit(\'openEditDialog\', list[0])">编辑 AWD</button></div>',
          },
          PlatformContestFormDialog: {
            props: ['open', 'draft'],
            template:
              '<div><button v-if="open" id="submit-awd-contest" type="button" @click="$emit(\'save\', { ...draft, mode: \'awd\', status: \'running\' })">保存 AWD</button></div>',
          },
          AWDReadinessOverrideDialog: {
            props: ['open', 'title'],
            data() {
              return { reason: '' }
            },
            template:
              '<div v-if="open"><div>{{ title }}</div><textarea id="awd-readiness-override-reason" v-model="reason" /><button id="awd-readiness-override-submit" type="button" @click="$emit(\'confirm\', reason)">强制继续</button></div>',
          },
          ElDialog: {
            props: ['modelValue', 'title'],
            template:
              '<div><div v-if="modelValue"><div>{{ title }}</div><slot /><slot name="footer" /></div></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#open-awd-edit').trigger('click')
    await flushPromises()
    await wrapper.get('#submit-awd-contest').trigger('click')
    await flushPromises()

    expect(contestMocks.getContestAWDReadiness).not.toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('填写本次放行原因')
  })

  it('应该在管理弹窗中结束进行中的竞赛前先请求确认', async () => {
    destructiveConfirmMock.mockResolvedValue(false)
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'contest-running',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-04-12T09:00:00.000Z',
          ends_at: '2026-04-12T18:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ContestOrchestrationPage: {
            props: ['list'],
            template:
              '<div><button id="open-edit" type="button" @click="$emit(\'openEditDialog\', list[0])">编辑</button></div>',
          },
          PlatformContestFormDialog: {
            props: ['open', 'draft'],
            template:
              '<div><button v-if="open" id="submit-ended" type="button" @click="$emit(\'save\', { ...draft, status: \'ended\' })">结束</button></div>',
          },
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#open-edit').trigger('click')
    await flushPromises()
    await wrapper.get('#submit-ended').trigger('click')
    await flushPromises()

    expect(destructiveConfirmMock).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '确认结束赛事',
      })
    )
    expect(contestMocks.updateContest).not.toHaveBeenCalled()
  })

  it('应该在管理弹窗中冻结进行中的竞赛时省略不可修改的时间字段', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'contest-running',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'running',
          starts_at: '2026-04-12T09:00:00.000Z',
          ends_at: '2026-04-12T18:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ContestOrchestrationPage: {
            props: ['list'],
            template:
              '<div><button id="open-edit" type="button" @click="$emit(\'openEditDialog\', list[0])">编辑</button></div>',
          },
          PlatformContestFormDialog: {
            props: ['open', 'draft'],
            template:
              '<div><button v-if="open" id="submit-frozen" type="button" @click="$emit(\'save\', { ...draft, status: \'frozen\' })">冻结</button></div>',
          },
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#open-edit').trigger('click')
    await flushPromises()
    await wrapper.get('#submit-frozen').trigger('click')
    await flushPromises()

    expect(contestMocks.updateContest).toHaveBeenCalledWith(
      'contest-running',
      expect.objectContaining({
        title: '2026 春季校园 CTF',
        status: 'frozen',
      })
    )
    expect(contestMocks.updateContest).toHaveBeenCalledWith(
      'contest-running',
      expect.not.objectContaining({
        starts_at: expect.anything(),
      })
    )
    expect(contestMocks.updateContest).toHaveBeenCalledWith(
      'contest-running',
      expect.not.objectContaining({
        ends_at: expect.anything(),
      })
    )
  })
  it('应该渲染真实竞赛列表', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      attachTo: document.body,
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('竞赛目录')
    expect(wrapper.text()).toContain('2026 春季校园 CTF')
    expect(wrapper.text()).toContain('报名中')
    expect(contestMocks.getContests).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: undefined,
    })
  })

  it('竞赛目录页不应渲染说明性文案和重复统计摘要', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      attachTo: document.body,
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain(
      '上面直接查看关键赛事指标，下面围绕具体竞赛对象完成筛选、编辑、导出和进入攻防运维。'
    )
    expect(wrapper.text()).not.toContain('当前页 1 场赛事')
  })

  it('赛事目录筛选应切到共享目录工具栏', () => {
    expect(contestOrchestrationSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(contestOrchestrationSource).toContain('<WorkspaceDirectoryToolbar')
    expect(contestOrchestrationSource).toContain('filter-panel-title="赛事筛选"')
    expect(contestOrchestrationSource).toContain('total-suffix="场赛事"')
    expect(contestOrchestrationSource).not.toMatch(
      /\.contest-directory-section,\s*\.contest-create-panel\s*\{[\s\S]*gap:\s*var\(--space-4\);/s
    )
    expect(contestOrchestrationSource).not.toMatch(
      /\.contest-directory-section :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(contestOrchestrationSource).not.toContain('<nav class="top-tabs"')
    expect(contestOrchestrationSource).not.toContain('class="contest-filter-grid"')
    expect(contestOrchestrationSource).not.toContain('class="contest-filter-strip"')
  })

  it('应该在赛事目录通过共享筛选面板切换状态筛选', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    contestMocks.getContests.mockClear()

    await wrapper.get('.workspace-directory-toolbar__filter-toggle').trigger('click')
    await flushPromises()
    await wrapper.get('.contest-filter-control').setValue('running')
    await flushPromises()

    expect(contestMocks.getContests).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: 'running',
    })
  })

  it('应该在赛事目录通过更多操作菜单点击编辑后跳转到独立编辑页', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'contest-1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#contest-row-more-contest-1').trigger('click')
    await flushPromises()

    const editButton = document.body.querySelector<HTMLButtonElement>(
      '#contest-row-menu-edit-contest-1'
    )

    expect(editButton).toBeTruthy()

    editButton!.click()
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith({ name: 'ContestEdit', params: { id: 'contest-1' } })
    wrapper.unmount()
  })

  it('管理页工作台入口应跳转到具体竞赛工作台，且不再保留顶层并行运维标签', async () => {
    window.sessionStorage.setItem('ctf_admin_awd_ops_panel:awd-running', 'challenges')
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: 'awd-registering',
          title: '2026 AWD 联赛预备场',
          description: '赛前检查',
          mode: 'awd',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
        {
          id: 'awd-running',
          title: '2026 AWD 联赛正式场',
          description: '正在进行',
          mode: 'awd',
          status: 'running',
          starts_at: '2026-03-16T09:00:00.000Z',
          ends_at: '2026-03-16T13:00:00.000Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('#contest-panel-operations').exists()).toBe(false)
    expect(wrapper.text()).toContain('进入 AWD 赛区')

    await wrapper.get('#contest-open-workbench-awd-running').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'ContestEdit',
      params: { id: 'awd-running' },
      query: { panel: 'operations', opsPanel: 'inspector' },
    })
  })

  it('应该在创建竞赛成功后切回赛事工作台', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    contestMocks.createContest.mockResolvedValue({
      id: '2',
      title: '2026 新生赛',
      description: '迎新赛',
      mode: 'jeopardy',
      status: 'draft',
      starts_at: '2026-03-20T09:00:00.000Z',
      ends_at: '2026-03-20T12:00:00.000Z',
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()
    await wrapper.get('#contest-open-create').trigger('click')
    await wrapper.get('#contest-title').setValue('2026 新生赛')
    await wrapper.get('#contest-description').setValue('迎新赛')
    await wrapper.get('#contest-starts-at').setValue('2026-03-20T09:00')
    await wrapper.get('#contest-ends-at').setValue('2026-03-20T12:00')
    await wrapper.get('#contest-panel-create .contest-form-button--primary').trigger('click')
    await flushPromises()

    expect(contestMocks.createContest).toHaveBeenCalledWith({
      title: '2026 新生赛',
      description: '迎新赛',
      mode: 'jeopardy',
      starts_at: new Date('2026-03-20T09:00').toISOString(),
      ends_at: new Date('2026-03-20T12:00').toISOString(),
    })
    expect(wrapper.get('#contest-panel-overview').attributes('aria-hidden')).toBe('false')
    expect(wrapper.text()).toContain('竞赛列表')
  })

  it('应该在空列表时展示显式空态', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('暂无竞赛')
  })
})
