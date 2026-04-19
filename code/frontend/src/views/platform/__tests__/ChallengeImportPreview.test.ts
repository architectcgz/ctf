import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeImportPreview from '../ChallengeImportPreview.vue'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { importId: 'import-1' } as Record<string, string>,
}))
const adminApiMocks = vi.hoisted(() => ({
  commitChallengeImport: vi.fn(),
  getChallengeImport: vi.fn(),
  listChallengeImports: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
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
    routeState.params = { importId: 'import-1' }
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
    const wrapper = mount(ChallengeImportPreview)
    await flushPromises()

    expect(wrapper.text()).toContain('导入预览')
    expect(wrapper.text()).toContain('Demo Challenge')
    expect(adminApiMocks.getChallengeImport).toHaveBeenCalledWith('import-1')

    await wrapper.get('.import-review__primary').trigger('click')
    await flushPromises()

    expect(adminApiMocks.commitChallengeImport).toHaveBeenCalledWith('import-1')
    expect(pushMock).toHaveBeenCalledWith({ name: 'ChallengeManage' })
  })
})
