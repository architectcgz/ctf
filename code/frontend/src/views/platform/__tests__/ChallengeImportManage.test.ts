import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChallengeImportManage from '../ChallengeImportManage.vue'
import challengeImportManageSource from '../ChallengeImportManage.vue?raw'
import challengeImportHeroPanelSource from '@/components/platform/challenge/ChallengeImportHeroPanel.vue?raw'
import challengeImportQueuePanelSource from '@/components/platform/challenge/ChallengeImportQueuePanel.vue?raw'
import challengePackageImportEntrySource from '@/components/platform/challenge/ChallengePackageImportEntry.vue?raw'

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

vi.mock('@/api/admin/authoring', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/authoring')>('@/api/admin/authoring')
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
    expect(challengeImportHeroPanelSource).toContain('Challenge Import')
    expect(challengeImportHeroPanelSource).toContain('返回题目目录')
    expect(challengeImportHeroPanelSource).toContain('题目包规范')
    expect(challengeImportHeroPanelSource).toContain('下载示例题目包')
    expect(challengePackageImportEntrySource).toContain(
      'class="ui-btn ui-btn--primary challenge-import-action challenge-import-action--primary import-entry__upload-action"'
    )
    expect(challengePackageImportEntrySource).toContain('导入题目包')
    expect(challengePackageImportEntrySource).not.toContain('class="import-entry__dropzone"')

    const wrapper = mount(ChallengeImportManage)
    await flushPromises()

    expect(wrapper.text()).toContain('导入题目')
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
      .mockRejectedValueOnce(new Error('格式错误'))

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

  it('导入页应复用 challenge entity 的分类与难度文案规则', () => {
    expect(challengeImportQueuePanelSource).toContain("from '@/entities/challenge'")
    expect(challengeImportManageSource).not.toContain('const categoryLabels = {')
    expect(challengeImportManageSource).not.toContain('const difficultyLabels = {')
  })
})
