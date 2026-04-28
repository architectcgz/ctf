import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AWDServiceTemplateLibrary from '../AWDServiceTemplateLibrary.vue'
import AWDServiceTemplateImport from '../AWDServiceTemplateImport.vue'
import awdServiceTemplateLibrarySource from '../AWDServiceTemplateLibrary.vue?raw'
import awdServiceTemplateImportSource from '../AWDServiceTemplateImport.vue?raw'

const pushMock = vi.fn()

const actionMocks = vi.hoisted(() => ({
  refresh: vi.fn(),
  changePage: vi.fn(),
  openCreateDialog: vi.fn(),
  openEditDialog: vi.fn(),
  closeDialog: vi.fn(),
  saveTemplate: vi.fn(),
  removeTemplate: vi.fn(),
  refreshImportQueue: vi.fn(),
  selectImportPackages: vi.fn(),
  commitImportPreview: vi.fn(),
}))

vi.mock('@/composables/usePlatformAwdServiceTemplates', () => ({
  usePlatformAwdServiceTemplates: () => ({
    list: ref([
      {
        id: '1',
        name: 'Bank Portal AWD',
        slug: 'bank-portal-awd',
        category: 'web',
        difficulty: 'hard',
        description: 'desc',
        service_type: 'web_http',
        deployment_mode: 'single_container',
        version: 'v1',
        status: 'draft',
        readiness_status: 'pending',
        created_at: '2026-04-17T08:00:00.000Z',
        updated_at: '2026-04-17T09:00:00.000Z',
      },
    ]),
    total: ref(1),
    page: ref(1),
    pageSize: ref(20),
    loading: ref(false),
    keyword: ref(''),
    serviceTypeFilter: ref(''),
    statusFilter: ref(''),
    dialogOpen: ref(false),
    dialogMode: ref<'create' | 'edit'>('create'),
    saving: ref(false),
    uploading: ref(false),
    queueLoading: ref(false),
    importQueue: ref([]),
    uploadResults: ref([]),
    selectedImportFileName: ref(''),
    formDraft: ref({
      name: '',
      slug: '',
      category: 'web',
      difficulty: 'medium',
      description: '',
      service_type: 'web_http',
      deployment_mode: 'single_container',
      status: 'draft',
    }),
    ...actionMocks,
  }),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

beforeEach(() => {
  pushMock.mockReset()
  Object.values(actionMocks).forEach((mock) => mock.mockClear())
})

describe('AWDServiceTemplateLibrary', () => {
  it('wires the awd service template workspace and editor dialog', async () => {
    const wrapper = mount(AWDServiceTemplateLibrary)
    await flushPromises()

    expect(wrapper.text()).toContain('AWD 服务模板库')
    expect(wrapper.text()).toContain('导入题目包')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(actionMocks.refresh).toHaveBeenCalledTimes(1)
    expect(actionMocks.refreshImportQueue).not.toHaveBeenCalled()

    await wrapper.findAll('button').find((button) => button.text() === '导入题目包')?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformAwdServiceTemplateImport' })
  })

  it('does not add an extra route-level spacing wrapper around the shared workspace shell', () => {
    expect(awdServiceTemplateLibrarySource).toContain('<template>\n  <div>')
    expect(awdServiceTemplateLibrarySource).not.toContain('<div class="space-y-6">')
  })
})

describe('AWDServiceTemplateImport', () => {
  it('wires the standalone awd import workspace', async () => {
    const wrapper = mount(AWDServiceTemplateImport)
    await flushPromises()

    expect(wrapper.text()).toContain('导入 AWD 题目包')
    expect(actionMocks.refreshImportQueue).toHaveBeenCalledTimes(1)
  })

  it('renders the import page mode without a route-level spacing wrapper', () => {
    expect(awdServiceTemplateImportSource).toContain('mode="import"')
    expect(awdServiceTemplateImportSource).not.toContain('<div class="space-y-6">')
  })
})
