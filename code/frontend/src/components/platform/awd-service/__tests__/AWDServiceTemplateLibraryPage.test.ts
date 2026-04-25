import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AWDServiceTemplateLibraryPage from '../AWDServiceTemplateLibraryPage.vue'

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
})
