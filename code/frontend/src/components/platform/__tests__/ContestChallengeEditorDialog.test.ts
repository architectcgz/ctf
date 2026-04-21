import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import ContestChallengeEditorDialog from '../contest/ContestChallengeEditorDialog.vue'

function mountDialog(props?: Record<string, unknown>) {
  return mount(ContestChallengeEditorDialog, {
    props: {
      open: true,
      mode: 'create',
      contestMode: 'awd',
      challengeOptions: [],
      templateOptions: [
        {
          id: '11',
          name: 'Bank Portal AWD',
          slug: 'bank-portal-awd',
          category: 'web',
          difficulty: 'hard',
          description: 'multi-step banking target',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: 'v1',
          status: 'published',
          readiness_status: 'passed',
          flag_mode: 'dynamic_team',
          flag_config: {
            flag_prefix: 'awd',
          },
          defense_entry_mode: 'http',
          access_config: {
            public_base_url: 'http://bank.internal',
            service_port: 8080,
          },
          runtime_config: {
            image_id: 9901,
          },
          created_at: '2026-04-17T08:00:00.000Z',
          updated_at: '2026-04-17T09:00:00.000Z',
        },
      ],
      existingChallengeIds: [],
      loadingChallengeCatalog: false,
      loadingTemplateCatalog: false,
      saving: false,
      ...props,
    },
    global: {
      stubs: {
        SlideOverDrawer: {
          props: ['open', 'title'],
          template:
            '<div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div>',
        },
      },
    },
  })
}

describe('ContestChallengeEditorDialog', () => {
  it('应该在 AWD 题目池创建时展示题库模板快照', () => {
    const wrapper = mountDialog()

    expect(wrapper.text()).toContain('关联 AWD 题库题目')
    expect(wrapper.text()).toContain('AWD 题库模板')
    expect(wrapper.text()).toContain('题库模板快照')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(wrapper.text()).toContain('public_base_url')
    expect(wrapper.text()).toContain('service_port')
    expect(wrapper.text()).toContain('flag_prefix')
  })
})
