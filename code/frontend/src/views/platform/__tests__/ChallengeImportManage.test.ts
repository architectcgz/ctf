import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeImportManage from '../ChallengeImportManage.vue'

const pushMock = vi.fn()
const adminApiMocks = vi.hoisted(() => ({
  commitChallengeImport: vi.fn(),
  getChallengeImport: vi.fn(),
  listChallengeImports: vi.fn(),
  previewChallengeImport: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
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
    previewChallengeImport: adminApiMocks.previewChallengeImport,
  }
})

describe('ChallengeImportManage', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.commitChallengeImport.mockReset()
    adminApiMocks.getChallengeImport.mockReset()
    adminApiMocks.listChallengeImports.mockReset()
    adminApiMocks.previewChallengeImport.mockReset()

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

  it('应将题目包规范、上传入口和待确认导入统一放进独立导入页', async () => {
    const wrapper = mount(ChallengeImportManage)
    await flushPromises()

    expect(wrapper.text()).toContain('导入资源包')
    expect(wrapper.text()).toContain('题目包规范')
    expect(wrapper.text()).toContain('最近上传结果')
    expect(wrapper.text()).toContain('待确认导入')
    expect(wrapper.text()).toContain('Web Demo')
  })

  it('支持多选上传并在导入页展示最近上传结果', async () => {
    adminApiMocks.previewChallengeImport
      .mockResolvedValueOnce({
        id: 'import-ok',
        file_name: 'ok.zip',
        slug: 'ok-challenge',
        title: 'OK Challenge',
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
        new Error('格式错误')
      )

    const wrapper = mount(ChallengeImportManage)
    await flushPromises()

    const fileInput = wrapper.get('input[type="file"]')
    Object.defineProperty(fileInput.element, 'files', {
      value: [new File([''], 'ok.zip'), new File([''], 'bad.zip')],
    })
    await fileInput.trigger('change')
    await flushPromises()

    expect(wrapper.text()).toContain('ok.zip')
    expect(wrapper.text()).toContain('bad.zip')
    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformChallengeImportPreview',
      params: { importId: 'import-ok' },
    })
  })
})
