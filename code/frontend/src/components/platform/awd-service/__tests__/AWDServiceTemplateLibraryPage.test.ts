import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AWDServiceTemplateLibraryPage from '../AWDServiceTemplateLibraryPage.vue'

describe('AWDServiceTemplateLibraryPage', () => {
  it('renders awd service template rows and emits row actions', async () => {
    const wrapper = mount(AWDServiceTemplateLibraryPage, {
      props: {
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
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(wrapper.text()).toContain('Web HTTP')
    expect(wrapper.text()).toContain('Single Container')

    const buttons = wrapper.findAll('button')
    await buttons.find((button) => button.text() === '编辑')?.trigger('click')
    await buttons.find((button) => button.text() === '删除')?.trigger('click')

    expect(wrapper.emitted('openEditDialog')).toHaveLength(1)
    expect(wrapper.emitted('deleteTemplate')).toHaveLength(1)
  })
})
