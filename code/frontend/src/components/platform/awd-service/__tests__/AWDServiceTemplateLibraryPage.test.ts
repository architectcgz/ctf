import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AWDServiceTemplateLibraryPage from '../AWDServiceTemplateLibraryPage.vue'
import awdServiceTemplateLibraryPageSource from '../AWDServiceTemplateLibraryPage.vue?raw'

describe('AWDServiceTemplateLibraryPage', () => {
  it('renders awd service template rows and emits row actions', async () => {
    const wrapper = mount(AWDServiceTemplateLibraryPage, {
      props: {
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
            warnings: ['meta.points 仅作为建议分值，不会直接写入模板。'],
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
      },
    })

    expect(wrapper.text()).toContain('AWD 服务模板库')
    expect(wrapper.text()).toContain('导入 AWD 题目包')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(wrapper.text()).toContain('Web HTTP')
    expect(wrapper.text()).toContain('Single')
    expect(wrapper.text()).toContain('bank-portal-awd')
    expect(wrapper.text()).not.toContain('@bank-portal-awd')

    let buttons = wrapper.findAll('button')
    await buttons.find((button) => button.text() === '编辑')?.trigger('click')
    await buttons.find((button) => button.text() === '删除')?.trigger('click')

    await wrapper.get('.awd-library-tabs button:last-child').trigger('click')
    expect(wrapper.text()).toContain('dynamic_team')

    buttons = wrapper.findAll('button')
    await buttons.find((button) => button.text() === '确认导入')?.trigger('click')

    expect(wrapper.emitted('openEditDialog')).toHaveLength(1)
    expect(wrapper.emitted('deleteTemplate')).toHaveLength(1)
    expect(wrapper.emitted('commitImport')).toHaveLength(1)
  })

  it('uses shared platform workspace surfaces instead of page-private premium panels', () => {
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero awd-template-library-shell"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="content-pane awd-template-library-content"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="admin-summary-grid awd-template-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="workspace-directory-section awd-import-tool-section"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="workspace-directory-section awd-import-queue-section"'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'class="workspace-directory-list awd-template-import__queue"'
    )
    expect(awdServiceTemplateLibraryPageSource).not.toContain('metric-panel-card--premium')
    expect(awdServiceTemplateLibraryPageSource).not.toContain('metric-panel-grid--premium')
  })

  it('keeps AWD template spacing on shared tokens instead of ad hoc page margins', () => {
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'gap: var(--workspace-directory-page-block-gap, var(--space-5));'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      'margin-top: var(--workspace-hero-summary-gap, var(--space-5));'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      '--admin-summary-grid-columns: repeat(4, minmax(0, 1fr));'
    )
    expect(awdServiceTemplateLibraryPageSource).toContain(
      '--metric-panel-columns: repeat(4, minmax(0, 1fr));'
    )
    expect(awdServiceTemplateLibraryPageSource).not.toContain('class="awd-library-tabs mt-10"')
    expect(awdServiceTemplateLibraryPageSource).not.toContain('class="awd-library-body mt-10"')
    expect(awdServiceTemplateLibraryPageSource).not.toContain('class="awd-library-pane space-y-10"')
    expect(awdServiceTemplateLibraryPageSource).not.toContain('class="awd-import-pane space-y-12"')
    expect(awdServiceTemplateLibraryPageSource).not.toContain('workspace-directory-pagination mt-6')
    expect(awdServiceTemplateLibraryPageSource).not.toContain('class="mt-8"')
  })
})
