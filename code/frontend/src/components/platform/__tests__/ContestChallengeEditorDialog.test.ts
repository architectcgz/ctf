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
      awdChallengeOptions: [
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
      loadingAwdChallengeCatalog: false,
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
  it('应该在 AWD 题目池创建时用列表选择题目且不展示快照', async () => {
    const wrapper = mountDialog()

    expect(wrapper.text()).toContain('关联 AWD 题目')
    expect(wrapper.text()).toContain('AWD 题目')
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(wrapper.text()).toContain('IoT Hub AWD')
    expect(wrapper.text()).toContain('web')
    expect(wrapper.text()).toContain('misc')
    expect(wrapper.text()).toContain('Web HTTP')
    expect(wrapper.text()).toContain('Binary TCP')
    expect(wrapper.text()).not.toContain('multi-step banking target')
    expect(wrapper.text()).not.toContain('device control target')
    expect(wrapper.find('#contest-challenge-template').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-challenge-option-11').classes()).toContain('is-selected')
    expect(wrapper.find('#contest-awd-service-points').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-order').exists()).toBe(false)
    expect(wrapper.find('#contest-awd-service-visibility').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('题目快照')
    expect(wrapper.text()).not.toContain('public_base_url')
    expect(wrapper.text()).not.toContain('flag_prefix')

    await wrapper.get('#contest-awd-challenge-name-12').trigger('click')
    expect(wrapper.find('#contest-awd-challenge-option-12').classes()).toContain('is-selected')

    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]).toEqual([
      {
        challenge_id: undefined,
        awd_challenge_id: 11,
        awd_challenge_ids: [11, 12],
        points: 100,
        order: 0,
        is_visible: true,
      },
    ])
  })

  it('AWD 题目池创建时应该支持复选多个题目', async () => {
    const wrapper = mountDialog()

    await wrapper.get('#contest-awd-challenge-option-12').trigger('click')

    expect(wrapper.find('#contest-awd-challenge-option-11').classes()).toContain('is-selected')
    expect(wrapper.find('#contest-awd-challenge-option-12').classes()).toContain('is-selected')

    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]).toEqual([
      {
        challenge_id: undefined,
        awd_challenge_id: 11,
        awd_challenge_ids: [11, 12],
        points: 100,
        order: 0,
        is_visible: true,
      },
    ])
  })

  it('编辑 AWD 题目编排时只编辑当前题目的分值顺序和可见性', async () => {
    const wrapper = mountDialog({
      mode: 'edit',
      draft: {
        id: 'service-11',
        contest_id: 'contest-1',
        challenge_id: '11',
        awd_service_id: 'service-11',
        awd_challenge_id: '11',
        title: 'Bank Portal AWD',
        category: 'web',
        difficulty: 'hard',
        points: 150,
        order: 2,
        is_visible: false,
        created_at: '2026-04-17T08:00:00.000Z',
      },
    })

    expect(wrapper.find('#contest-awd-challenge-list').exists()).toBe(false)
    expect(wrapper.text()).toContain('Bank Portal AWD')
    expect(wrapper.find('#contest-challenge-points').exists()).toBe(true)
    expect(wrapper.find('#contest-challenge-order').exists()).toBe(true)
    expect(wrapper.find('#contest-challenge-visibility').exists()).toBe(true)

    await wrapper.get('#contest-challenge-points').setValue('180')
    await wrapper.get('#contest-challenge-order').setValue('4')
    await wrapper.get('#contest-challenge-visibility').setValue('true')
    await wrapper.get('#contest-challenge-dialog-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]).toEqual([
      {
        challenge_id: undefined,
        awd_challenge_id: 11,
        awd_challenge_ids: undefined,
        points: 180,
        order: 4,
        is_visible: true,
      },
    ])
  })
})
