import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeImportPreview from '@/pages/platform/challenges/ChallengeImportPreviewRoutePage.vue'
import challengeImportPreviewSource from '@/pages/platform/challenges/ChallengeImportPreviewRoutePage.vue?raw'
import challengeImportPreviewPageModelSource from '@/features/platform/challenge-package-import/model/useChallengeImportPreviewPage.ts?raw'

const pushMock = vi.fn()
const adminApiMocks = vi.hoisted(() => ({
  commitChallengeImport: vi.fn(),
  getChallengeImport: vi.fn(),
  listChallengeImports: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    RouterLink: { props: ['to'], template: '<a :data-route-target="JSON.stringify(to)"><slot /></a>' },
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/admin/authoring', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/authoring')>('@/api/admin/authoring')
  return {
    ...actual,
    commitChallengeImport: adminApiMocks.commitChallengeImport,
    getChallengeImport: adminApiMocks.getChallengeImport,
    listChallengeImports: adminApiMocks.listChallengeImports,
  }
})

describe('ChallengeImportPreview', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.commitChallengeImport.mockReset()
    adminApiMocks.getChallengeImport.mockReset()
    adminApiMocks.listChallengeImports.mockReset()

    adminApiMocks.getChallengeImport.mockResolvedValue({
      id: 'import-1',
      file_name: 'demo.zip',
      slug: 'demo-challenge',
      title: 'Demo Challenge',
      description: 'demo description',
      category: 'web',
      difficulty: 'easy',
      points: 100,
      attachments: [],
      hints: [{ level: 1, title: 'Hint 1', content: 'hint content' }],
      flag: { type: 'static', prefix: 'flag' },
      runtime: { type: 'container', image_ref: 'ctf/demo:latest' },
      extensions: { topology: { source: '', enabled: false } },
      warnings: [],
      created_at: '2026-04-09T08:00:00.000Z',
    })
    adminApiMocks.commitChallengeImport.mockResolvedValue({
      challenge: {
        id: 'challenge-1',
        title: 'Demo Challenge',
        description: '',
        category: 'web',
        difficulty: 'easy',
        points: 100,
        status: 'draft',
        created_at: '2026-04-09T08:00:00.000Z',
        updated_at: '2026-04-09T08:00:00.000Z',
      },
    })
    adminApiMocks.listChallengeImports.mockResolvedValue([])
  })

  it('应按路由参数加载独立导入预览并支持确认导入', async () => {
    const wrapper = mount(ChallengeImportPreview, {
      props: {
        importId: 'import-1',
      },
      global: {
        stubs: {
          ChallengeImportPreviewWorkspacePanel: {
            props: ['preview'],
            emits: ['confirm'],
            template:
              '<div><div data-testid="preview-title">{{ preview?.title }}</div><button id="import-preview-confirm-primary" type="button" @click="$emit(\'confirm\')">确认导入</button></div>',
          },
        },
      },
    })
    await flushPromises()

    expect(wrapper.get('[data-testid="preview-title"]').text()).toBe('Demo Challenge')
    expect(adminApiMocks.getChallengeImport).toHaveBeenCalledWith('import-1')

    await wrapper.get('#import-preview-confirm-primary').trigger('click')
    await flushPromises()

    expect(adminApiMocks.commitChallengeImport).toHaveBeenCalledWith('import-1')
    expect(pushMock).toHaveBeenCalledWith({ name: 'ChallengeManage' })
  })

  it('路由页应仅负责组合，不直接耦合路由参数与提交流程', () => {
    expect(challengeImportPreviewSource).toContain('useChallengeImportPreviewPage')
    expect(challengeImportPreviewSource).not.toContain('useRoute')
    expect(challengeImportPreviewSource).not.toContain('useRouter')
    expect(challengeImportPreviewPageModelSource).not.toContain("from 'vue-router'")
  })

  it('父页应把返回导航下传为 route target，并只保留确认导入 owner', async () => {
    const wrapper = mount(ChallengeImportPreview, {
      props: {
        importId: 'import-1',
      },
      global: {
        stubs: {
          ChallengeImportPreviewWorkspacePanel: {
            props: ['preview', 'backToImportRoute', 'backToQueueRoute'],
            emits: ['confirm'],
            template:
              '<div><div data-testid="preview-title">{{ preview?.title }}</div><div data-testid="back-import-route">{{ JSON.stringify(backToImportRoute) }}</div><div data-testid="back-queue-route">{{ JSON.stringify(backToQueueRoute) }}</div><button id="import-preview-confirm" type="button" @click="$emit(\'confirm\')">确认导入</button></div>',
          },
        },
      },
    })
    await flushPromises()

    expect(wrapper.get('[data-testid="preview-title"]').text()).toBe('Demo Challenge')
    expect(wrapper.get('[data-testid="back-import-route"]').text()).toBe(
      JSON.stringify({ name: 'PlatformChallengeImportManage' })
    )
    expect(wrapper.get('[data-testid="back-queue-route"]').text()).toBe(
      JSON.stringify({ name: 'PlatformChallengeImportManage', hash: '#challenge-queue-workspace' })
    )

    await wrapper.get('#import-preview-confirm').trigger('click')
    await flushPromises()

    expect(adminApiMocks.commitChallengeImport).toHaveBeenCalledWith('import-1')
    expect(pushMock).toHaveBeenLastCalledWith({ name: 'ChallengeManage' })
  })
})
