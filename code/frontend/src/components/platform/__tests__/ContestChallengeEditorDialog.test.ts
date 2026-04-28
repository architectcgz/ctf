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
        {
          id: '12',
          name: 'IoT Hub AWD',
          slug: 'iot-hub-awd',
          category: 'misc',
          difficulty: 'medium',
          description: 'device control target',
          service_type: 'binary_tcp',
          deployment_mode: 'single_container',
          version: 'v1',
          status: 'published',
          readiness_status: 'pending',
          created_at: '2026-04-18T08:00:00.000Z',
          updated_at: '2026-04-18T09:00:00.000Z',
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
        AdminSurfaceModal: {
          props: ['open', 'title'],
          template:
            '<div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div>',
        },
      },
    },
  })
}

describe('ContestChallengeEditorDialog', () => {
  it('应该在 AWD 题目池创建时用列表选择题库模板且不展示快照', async () => {
    const wrapper = mountDialog()

    expect(wrapper.text()).toContain('关联 AWD 题库服务')
    expect(wrapper.text()).toContain('AWD 题库模板')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(wrapper.text()).toContain('IoT Hub AWD')
    expect(wrapper.text()).toContain('web')
    expect(wrapper.text()).toContain('misc')
    expect(wrapper.text()).toContain('Web HTTP')
    expect(wrapper.text()).toContain('Binary TCP')
    expect(wrapper.text()).not.toContain('multi-step banking target')
    expect(wrapper.text()).not.toContain('device control target')
    expect(wrapper.find('#contest-challenge-template').exists()).toBe(false)
    expect(wrapper.find('#contest-template-option-11').classes()).toContain('is-selected')
    expect(wrapper.text()).not.toContain('题库模板快照')
    expect(wrapper.text()).not.toContain('public_base_url')
    expect(wrapper.text()).not.toContain('flag_prefix')

    await wrapper.get('#contest-template-option-12').trigger('click')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]).toEqual([
      {
        challenge_id: undefined,
        template_id: 12,
        points: 100,
        order: 0,
        is_visible: true,
      },
    ])
  })
})
