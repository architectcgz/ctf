import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeManage from '../ChallengeManage.vue'
import challengeManageSource from '../ChallengeManage.vue?raw'
import { ApiError } from '@/api/request'

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

    expect(wrapper.text()).toContain('题库管理')
    expect(wrapper.text()).toContain('题目管理')
    expect(wrapper.text()).toContain('导入题目包')
    expect(wrapper.text()).toContain('待确认导入')
    expect(wrapper.find('#challenge-tab-manage').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#challenge-panel-manage').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#challenge-panel-import').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#challenge-panel-queue').attributes('aria-hidden')).toBe('true')
    expect(wrapper.text()).toContain('Test Challenge')
    expect(wrapper.text()).not.toContain('CH-')
    expect(wrapper.text()).not.toContain('Publish Check')

    const headers = wrapper.findAll('.manage-directory-head span').map((item) => item.text())
    expect(headers).toEqual(['题目', '分类', '难度', '分值', '发布状态', '发布检查', '操作'])

    const row = wrapper.find('.challenge-row')
    expect(row.find('.challenge-row__identity').exists()).toBe(true)
    expect(row.find('.challenge-row__title').classes()).toContain('challenge-row__title')
    expect(row.find('.challenge-row__failure').classes()).toContain('challenge-row__failure')
    expect(row.find('.challenge-row__category').text()).toContain('Web')
    expect(row.find('.challenge-row__difficulty').text()).toContain('简单')
    expect(row.find('.challenge-row__points').text()).toContain('100 pts')
    expect(row.find('.challenge-row__status').text()).toContain('草稿')
    expect(row.findAll('.challenge-row__actions > button')).toHaveLength(2)
    expect(row.find('.challenge-row__actions').text()).toContain('查看')
    expect(row.find('.challenge-row__actions').text()).toContain('更多')
    expect(row.find('.challenge-row__actions').text()).not.toContain('编排')
    expect(row.find('.challenge-row__actions').text()).not.toContain('题解')
    expect(row.find('.challenge-row__actions').text()).not.toContain('提交发布检查')
    expect(row.find('.challenge-row__actions').text()).not.toContain('删除')

    await row.get('[data-testid="challenge-more-actions"]').trigger('click')
    await flushPromises()

    expect(row.find('.challenge-row__actions-menu').exists()).toBe(true)
    expect(row.find('.challenge-row__actions-menu').text()).toContain('编排')
    expect(row.find('.challenge-row__actions-menu').text()).toContain('题解')
    expect(row.find('.challenge-row__actions-menu').text()).toContain('提交发布检查')
    expect(row.find('.challenge-row__actions-menu').text()).toContain('删除')
  })

  it('不应该在 1200px 断点把题目列表强制折成单列', () => {
    expect(challengeManageSource).not.toMatch(
      /\.challenge-row,\s*\.queue-row\s*\{\s*grid-template-columns: minmax\(0, 1fr\);/s
    )
    expect(challengeManageSource).not.toMatch(/\.challenge-list\s*\{[^}]*--challenge-list-columns:[^;]*\bauto;/s)
    expect(challengeManageSource).toContain('.challenge-row__title')
    expect(challengeManageSource).toMatch(/\.challenge-row__title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
    expect(challengeManageSource).toMatch(/\.challenge-row__failure\s*\{[^}]*display:\s*-webkit-box;[^}]*-webkit-line-clamp:\s*2;[^}]*overflow:\s*hidden;/s)
    expect(challengeManageSource).toMatch(/class="queue-row__title"[\s\S]*:title="item\.title"/s)
    expect(challengeManageSource).toMatch(/class="queue-row__meta-text"[\s\S]*:title="item\.file_name"/s)
    expect(challengeManageSource).toMatch(/\.queue-row__title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
    expect(challengeManageSource).toMatch(/\.queue-row__meta-text\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
  })

  it('应该根据 query 切到待确认导入，并在导入标签中只保留独立示例页入口', async () => {
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

    expect(importWrapper.text()).toContain('查看题目包示例')
    expect(importWrapper.text()).toContain('下载示例题目包')
    expect(importWrapper.text()).not.toContain('challenge-package.zip')
    expect(importWrapper.text()).not.toContain('api_version: v1')
    expect(wrapper.find('.queue-row__title').attributes('title')).toBe('Web Demo')
    expect(wrapper.find('.queue-row__meta-text').attributes('title')).toBe('demo-import.zip')
    expect(importWrapper.get('[data-testid="challenge-package-download-link"]').attributes('href')).toBe(
      '/downloads/challenge-package-sample-v1.zip'
    )
    expect(importWrapper.get('[data-testid="challenge-package-download-link"]').attributes('download')).toBe(
      'challenge-package-sample-v1.zip'
    )

    await importWrapper.get('[data-testid="challenge-package-format-link"]').trigger('click')

    expect(pushMock).toHaveBeenLastCalledWith({ name: 'AdminChallengePackageFormat' })
  })

  it('支持多选上传，并在上传区域下方显示每个文件的结果', async () => {
    adminApiMocks.previewChallengeImport
      .mockResolvedValueOnce({
        id: 'import-ok',
        file_name: 'ok.zip',
        slug: 'ok-challenge',
        title: 'OK Challenge',
        description: 'ok',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        attachments: [],
        hints: [],
        flag: { type: 'static', prefix: 'flag' },
        runtime: { type: 'container', image_ref: 'ctf/ok:latest' },
        extensions: { topology: { source: '', enabled: false } },
        warnings: [],
        created_at: '2026-04-06T09:10:00.000Z',
      })
      .mockRejectedValueOnce(
        new ApiError('请求参数错误', {
          code: 10001,
          requestId: 'req_18056986c1123ac6',
        })
      )

    const wrapper = mount(ChallengeManage)
    await flushPromises()
    await wrapper.get('#challenge-tab-import').trigger('click')
    await flushPromises()

    const fileInput = wrapper.get('input[type="file"]')
    expect(fileInput.attributes('multiple')).toBeDefined()

    Object.defineProperty(fileInput.element, 'files', {
      configurable: true,
      value: [
        new File(['ok'], 'ok.zip', { type: 'application/zip' }),
        new File(['bad'], 'bad.zip', { type: 'application/zip' }),
      ],
    })
    await fileInput.trigger('change')
    await flushPromises()
    expect(adminApiMocks.previewChallengeImport).toHaveBeenCalledTimes(2)

    const text = wrapper.text()
    expect(text).toContain('最近上传结果')
    expect(text).toContain('ok.zip')
    expect(text).toContain('bad.zip')
    expect(text).toContain('成功')
    expect(text).toContain('失败')
    expect(text).toContain('题目包格式校验失败')
    expect(text).toContain('错误码 10001')
    expect(text).toContain('请求ID req_18056986c1123ac6')
  })
})
