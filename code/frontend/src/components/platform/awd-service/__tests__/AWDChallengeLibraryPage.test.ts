import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type {
  AdminAwdChallengeData,
  AdminAwdChallengeImportPreview,
} from '@/api/contracts'
import type { PlatformAwdChallengeImportUploadResult } from '@/composables/usePlatformAwdChallenges'

import AWDChallengeLibraryPage from '../AWDChallengeLibraryPage.vue'
import awdChallengeLibraryPageSource from '../AWDChallengeLibraryPage.vue?raw'

describe('AWDChallengeLibraryPage', () => {
  interface TestPageProps {
    uploading: boolean
    queueLoading: boolean
    importQueue: AdminAwdChallengeImportPreview[]
    uploadResults: PlatformAwdChallengeImportUploadResult[]
    selectedFileName: string
    list: AdminAwdChallengeData[]
    total: number
    page: number
    pageSize: number
    loading: boolean
    keyword: string
    serviceTypeFilter: ''
    statusFilter: ''
  }

  function createProps(): TestPageProps {
    return {
      uploading: false,
      queueLoading: false,
      importQueue: [
        {
          id: 'imp-1',
          file_name: 'awd-bank-portal-01.zip',
          slug: 'awd-bank-portal-01',
          title: 'Bank Portal AWD',
          category: 'web',
          difficulty: 'hard',
          description: 'multi-step banking target',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: 'v2026.04',
          checker_type: 'http_standard',
          checker_config: {
            put_flag: {
              method: 'PUT',
              path: '/api/flag',
            },
          },
          flag_mode: 'dynamic_team',
          flag_config: {
            flag_prefix: 'awd',
          },
          defense_entry_mode: 'http',
          access_config: {
            public_base_url: 'http://{{TEAM_HOST}}:8080',
            service_port: 8080,
          },
          runtime_config: {
            image_ref: 'registry.example.edu/ctf/awd-bank-portal:v1',
            service_port: 8080,
          },
          warnings: ['meta.points 仅作为建议分值，不会直接写入 AWD 题目。'],
          created_at: '2026-04-21T08:00:00.000Z',
        },
      ],
      uploadResults: [],
      selectedFileName: '',
      list: [
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
      ],
      total: 1,
      page: 1,
      pageSize: 20,
      loading: false,
      keyword: '',
      serviceTypeFilter: '',
      statusFilter: '',
    }
  }

  it('renders awd challenge rows and emits row actions', async () => {
    const wrapper = mount(AWDChallengeLibraryPage, {
      props: createProps(),
    })

    expect(wrapper.text()).toContain('AWD 题目库')
    expect(wrapper.text()).toContain('导入题目包')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(wrapper.text()).toContain('Web HTTP')
    expect(wrapper.text()).toContain('Single')
    expect(wrapper.text()).toContain('bank-portal-awd')
    expect(wrapper.text()).not.toContain('@bank-portal-awd')
    expect(wrapper.find('.awd-library-tabs').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('创建模板')

    const buttons = wrapper.findAll('button')
    await buttons.find((button) => button.text() === '编辑')?.trigger('click')
    await buttons.find((button) => button.text() === '删除')?.trigger('click')
    await buttons.find((button) => button.text() === '导入题目包')?.trigger('click')

    expect(wrapper.emitted('openEditDialog')).toHaveLength(1)
    expect(wrapper.emitted('deleteChallenge')).toHaveLength(1)
    expect(wrapper.emitted('openImportPage')).toHaveLength(1)
    expect(wrapper.emitted('commitImport')).toBeUndefined()
  })

  it('renders the import workspace as a standalone page mode', async () => {
    const wrapper = mount(AWDChallengeLibraryPage, {
      props: {
        ...createProps(),
        mode: 'import',
      },
    })

    expect(wrapper.text()).toContain('导入 AWD 题目包')
    expect(wrapper.text()).toContain('dynamic_team')
    expect(wrapper.find('.awd-library-tabs').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('AWD 题目库')

    const buttons = wrapper.findAll('button')
    await buttons.find((button) => button.text() === '确认导入')?.trigger('click')

    expect(wrapper.emitted('commitImport')).toHaveLength(1)
  })

  it('uses shared platform workspace surfaces instead of page-private premium panels', () => {
    expect(awdChallengeLibraryPageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero awd-template-library-shell"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="content-pane awd-template-library-content"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="admin-summary-grid awd-template-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
    expect(awdChallengeLibraryPageSource).toContain('当前筛选条件下可管理的题目')
    expect(awdChallengeLibraryPageSource).toContain('已开放给 AWD 编排使用的题目')
    expect(awdChallengeLibraryPageSource).toContain('使用 HTTP 探测与 Web 服务模式的题目')
    expect(awdChallengeLibraryPageSource).toContain('仍需完成 Checker 验证的题目')
    expect(awdChallengeLibraryPageSource).toContain(
      'class="workspace-directory-section awd-import-tool-section"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="workspace-directory-section awd-import-queue-section"'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      'class="workspace-directory-list awd-template-import__queue"'
    )
    expect(awdChallengeLibraryPageSource).not.toContain('metric-panel-card--premium')
    expect(awdChallengeLibraryPageSource).not.toContain('metric-panel-grid--premium')
  })

  it('keeps AWD template spacing on shared tokens instead of ad hoc page margins', () => {
    expect(awdChallengeLibraryPageSource).toContain(
      'gap: var(--workspace-directory-page-block-gap, var(--space-5));'
    )
    expect(awdChallengeLibraryPageSource).not.toContain(
      '.awd-library-body {\n  margin-top: var(--workspace-hero-summary-gap, var(--space-5));\n}'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      '--admin-summary-grid-columns: repeat(4, minmax(0, 1fr));'
    )
    expect(awdChallengeLibraryPageSource).toContain(
      '--metric-panel-columns: repeat(4, minmax(0, 1fr));'
    )
    expect(awdChallengeLibraryPageSource).not.toContain('class="awd-library-tabs mt-10"')
    expect(awdChallengeLibraryPageSource).not.toContain('class="awd-library-body mt-10"')
    expect(awdChallengeLibraryPageSource).not.toContain('class="awd-library-pane space-y-10"')
    expect(awdChallengeLibraryPageSource).not.toContain('class="awd-import-pane space-y-12"')
    expect(awdChallengeLibraryPageSource).not.toContain('workspace-directory-pagination mt-6')
    expect(awdChallengeLibraryPageSource).not.toContain('class="mt-8"')
  })
})
