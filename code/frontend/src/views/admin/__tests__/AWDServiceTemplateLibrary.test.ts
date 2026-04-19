import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { describe, expect, it, vi } from 'vitest'

import AWDServiceTemplateLibrary from '../AWDServiceTemplateLibrary.vue'

const actionMocks = vi.hoisted(() => ({
  refresh: vi.fn(),
  changePage: vi.fn(),
  openCreateDialog: vi.fn(),
  openEditDialog: vi.fn(),
  closeDialog: vi.fn(),
  saveTemplate: vi.fn(),
  removeTemplate: vi.fn(),
}))

vi.mock('@/composables/useAdminAwdServiceTemplates', () => ({
  useAdminAwdServiceTemplates: () => ({
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

describe('AWDServiceTemplateLibrary', () => {
  it('wires the awd service template workspace and editor dialog', async () => {
    const wrapper = mount(AWDServiceTemplateLibrary)
    await flushPromises()

    expect(wrapper.text()).toContain('AWD 服务模板库')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(actionMocks.refresh).toHaveBeenCalledTimes(1)
  })
})
