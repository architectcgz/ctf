import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AWDChallengeLibrary from '../AWDChallengeLibrary.vue'
import AWDChallengeImport from '../AWDChallengeImport.vue'
import awdChallengeLibrarySource from '../AWDChallengeLibrary.vue?raw'
import awdChallengeImportSource from '../AWDChallengeImport.vue?raw'
import awdChallengeLibraryPageModelSource from '@/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts?raw'

const actionMocks = vi.hoisted(() => ({
  refresh: vi.fn(),
  changePage: vi.fn(),
  openCreateDialog: vi.fn(),
  openEditDialog: vi.fn(),
  closeDialog: vi.fn(),
  saveChallenge: vi.fn(),
  removeChallenge: vi.fn(),
  refreshImportQueue: vi.fn(),
  selectImportPackages: vi.fn(),
  commitImportPreview: vi.fn(),
}))

vi.mock('@/features/platform-awd-challenges', () => ({
  AWDChallengeEditorDialog: {
    name: 'AWDChallengeEditorDialog',
    template: '<div data-testid="awd-challenge-editor-dialog" />',
  },
  AWDChallengeLibraryPage: {
    name: 'AWDChallengeLibraryPage',
    props: [
      'mode',
      'list',
      'total',
      'page',
      'pageSize',
      'loading',
      'keyword',
      'serviceTypeFilter',
      'statusFilter',
      'uploading',
      'queueLoading',
      'importQueue',
      'uploadResults',
      'selectedFileName',
      'importRoute',
    ],
    template: `
      <div>
        <h1>{{ mode === 'import' ? '导入 AWD 题目包' : 'AWD 题目库' }}</h1>
        <div>导入题目包</div>
        <div data-import-route-name>{{ importRoute?.name }}</div>
        <div v-for="item in list" :key="item.id">{{ item.name }}</div>
      </div>
    `,
  },
  useAwdChallengeLibraryPage: () => {
    actionMocks.refresh()
    return {
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
      updateKeyword: vi.fn(),
      updateServiceTypeFilter: vi.fn(),
      updateStatusFilter: vi.fn(),
      handleDialogOpenChange: vi.fn(),
      importRoute: { name: 'PlatformAwdChallengeImport' as const },
      ...actionMocks,
    }
  },
  useAwdChallengeImportPage: () => {
    actionMocks.refreshImportQueue()
    return {
      uploading: ref(false),
      queueLoading: ref(false),
      importQueue: ref([]),
      uploadResults: ref([]),
      selectedImportFileName: ref(''),
      refreshImportQueue: actionMocks.refreshImportQueue,
      selectImportPackages: actionMocks.selectImportPackages,
      commitImportPreview: actionMocks.commitImportPreview,
    }
  },
}))

beforeEach(() => {
  Object.values(actionMocks).forEach((mock) => mock.mockClear())
})

describe('AWDChallengeLibrary', () => {
  it('wires the awd challenge workspace and editor dialog', async () => {
    const wrapper = mount(AWDChallengeLibrary)
    await flushPromises()

    expect(wrapper.text()).toContain('AWD 题目库')
    expect(wrapper.text()).toContain('导入题目包')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(actionMocks.refresh).toHaveBeenCalledTimes(1)
    expect(actionMocks.refreshImportQueue).not.toHaveBeenCalled()
    expect(wrapper.get('[data-import-route-name]').text()).toBe('PlatformAwdChallengeImport')
  })

  it('does not add an extra route-level spacing wrapper around the shared workspace shell', () => {
    expect(awdChallengeLibrarySource).toContain('<template>\n  <div>')
    expect(awdChallengeLibrarySource).not.toContain('<div class="space-y-6">')
    expect(awdChallengeLibrarySource).toContain("from '@/features/platform-awd-challenges'")
    expect(awdChallengeLibrarySource).toContain('AWDChallengeLibraryPage')
    expect(awdChallengeLibrarySource).toContain('useAwdChallengeLibraryPage')
    expect(awdChallengeLibrarySource).not.toContain(
      "from '@/components/platform/awd-service/AWDChallengeLibraryPage.vue'"
    )
    expect(awdChallengeLibrarySource).not.toContain('useRouter')
    expect(awdChallengeLibrarySource).not.toContain('usePlatformAwdChallenges')
    expect(awdChallengeLibraryPageModelSource).not.toContain("from 'vue-router'")
  })
})

describe('AWDChallengeImport', () => {
  it('wires the standalone awd import workspace', async () => {
    const wrapper = mount(AWDChallengeImport)
    await flushPromises()

    expect(wrapper.text()).toContain('导入 AWD 题目包')
    expect(actionMocks.refreshImportQueue).toHaveBeenCalledTimes(1)
  })

  it('renders the import page mode without a route-level spacing wrapper', () => {
    expect(awdChallengeImportSource).toContain('mode="import"')
    expect(awdChallengeImportSource).not.toContain('<div class="space-y-6">')
    expect(awdChallengeImportSource).toContain("from '@/features/platform-awd-challenges'")
    expect(awdChallengeImportSource).toContain('AWDChallengeLibraryPage')
    expect(awdChallengeImportSource).toContain('useAwdChallengeImportPage')
    expect(awdChallengeImportSource).not.toContain('onMounted(')
  })
})
