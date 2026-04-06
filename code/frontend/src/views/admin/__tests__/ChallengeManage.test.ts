import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeManage from '../ChallengeManage.vue'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeState = vi.hoisted(() => ({
  query: {} as Record<string, string>,
}))
const adminApiMocks = vi.hoisted(() => ({
  commitChallengeImport: vi.fn(),
  createChallenge: vi.fn(),
  createChallengePublishRequest: vi.fn(),
  deleteChallenge: vi.fn(),
  getChallengeDetail: vi.fn(),
  getChallengeFlagConfig: vi.fn(),
  getChallengeImport: vi.fn(),
  getChallenges: vi.fn(),
  getImages: vi.fn(),
  getLatestChallengePublishRequest: vi.fn(),
  listChallengeImports: vi.fn(),
  previewChallengeImport: vi.fn(),
  updateChallenge: vi.fn(),
  configureChallengeFlag: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)

describe('ChallengeManage', () => {
  beforeEach(() => {
    pushMock.mockReset()
    replaceMock.mockReset()
    routeState.query = {}
    adminApiMocks.commitChallengeImport.mockReset()
    adminApiMocks.createChallenge.mockReset()
    adminApiMocks.createChallengePublishRequest.mockReset()
    adminApiMocks.deleteChallenge.mockReset()
    adminApiMocks.getChallengeDetail.mockReset()
    adminApiMocks.getChallengeFlagConfig.mockReset()
    adminApiMocks.getChallengeImport.mockReset()
    adminApiMocks.getChallenges.mockReset()
    adminApiMocks.getImages.mockReset()
    adminApiMocks.getLatestChallengePublishRequest.mockReset()
    adminApiMocks.listChallengeImports.mockReset()
    adminApiMocks.previewChallengeImport.mockReset()
    adminApiMocks.updateChallenge.mockReset()
    adminApiMocks.configureChallengeFlag.mockReset()

    adminApiMocks.getChallenges.mockResolvedValue({
      list: [
        {
          id: '1',
          title: 'Test Challenge',
          category: 'web',
          difficulty: 'easy',
          status: 'draft',
          points: 100,
          created_at: '2026-04-01T08:00:00.000Z',
          updated_at: '2026-04-01T08:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getLatestChallengePublishRequest.mockResolvedValue({
      id: 'req-1',
      challenge_id: '1',
      status: 'failed',
      active: false,
      failure_summary: 'Flag 未配置',
      created_at: '2026-04-01T08:00:00.000Z',
      updated_at: '2026-04-01T08:01:00.000Z',
    })
    adminApiMocks.listChallengeImports.mockResolvedValue([
      {
        id: 'import-1',
        file_name: 'demo-import.zip',
        slug: 'web-demo',
        title: 'Web Demo',
        description: 'Demo import preview',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        attachments: [{ name: 'demo.zip', path: 'attachments/demo.zip' }],
        hints: [{ level: 1, title: 'Hint 1', content: 'Check login flow' }],
        flag: { type: 'static', prefix: 'flag' },
        runtime: { type: 'container', image_ref: 'ctf/web-demo:latest' },
        extensions: { topology: { source: 'docker/topology.yml', enabled: false } },
        warnings: [],
        created_at: '2026-04-06T09:00:00.000Z',
      },
    ])
  })

  it('应该默认显示题目管理 tab，并提供导入和队列标签', async () => {
    const wrapper = mount(ChallengeManage)
    await flushPromises()

    expect(wrapper.text()).toContain('靶场管理')
    expect(wrapper.text()).toContain('题目管理')
    expect(wrapper.text()).toContain('导入题目包')
    expect(wrapper.text()).toContain('上传队列')
    expect(wrapper.find('#challenge-tab-manage').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#challenge-panel-manage').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#challenge-panel-import').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#challenge-panel-queue').attributes('aria-hidden')).toBe('true')
    expect(wrapper.text()).toContain('Test Challenge')
    expect(wrapper.text()).toContain('提交发布检查')
  })

  it('应该根据 query 切到上传队列，并支持切换到导入标签和示例页', async () => {
    routeState.query = { panel: 'queue' }

    const wrapper = mount(ChallengeManage)
    await flushPromises()

    expect(wrapper.find('#challenge-tab-queue').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#challenge-panel-queue').attributes('aria-hidden')).toBe('false')
    expect(wrapper.text()).toContain('demo-import.zip')
    expect(wrapper.text()).toContain('Web Demo')

    await wrapper.get('#challenge-tab-import').trigger('click')
    await flushPromises()

    expect(replaceMock).toHaveBeenLastCalledWith({
      name: 'ChallengeManage',
      query: { panel: 'import' },
    })

    routeState.query = { panel: 'import' }
    const importWrapper = mount(ChallengeManage)
    await flushPromises()

    const sampleButton = importWrapper
      .findAll('button')
      .find((button) => button.text().includes('查看题目包示例'))

    expect(sampleButton).toBeTruthy()
    await sampleButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'AdminChallengePackageFormat',
    })
  })
})
