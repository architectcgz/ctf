import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AWDChallengeConfigDialog from '../contest/AWDChallengeConfigDialog.vue'

function mountDialog(props?: Record<string, unknown>) {
  return mount(AWDChallengeConfigDialog, {
    props: {
      open: true,
      mode: 'create',
      challengeOptions: [
        {
          id: '101',
          title: 'Web Checker',
          description: '已有题目',
          category: 'web',
          difficulty: 'easy',
          points: 120,
          instance_sharing: 'per_user',
          created_by: '9',
          image_id: undefined,
          attachment_url: undefined,
          hints: undefined,
          status: 'published',
          created_at: '2026-03-20T09:00:00.000Z',
          updated_at: '2026-03-20T09:00:00.000Z',
          flag_config: undefined,
        },
      ],
      existingChallengeIds: [],
      draft: null,
      loadingChallengeCatalog: false,
      saving: false,
      ...props,
    },
    global: {
      stubs: {
        ElDialog: {
          props: ['modelValue', 'title'],
          template:
            '<div><div v-if="modelValue">{{ title }}</div><slot /><slot name="footer" /></div>',
        },
      },
    },
  })
}

describe('AWDChallengeConfigDialog', () => {
  it('应该在编辑 http_standard 配置时回填结构化字段和 JSON 预览', async () => {
    const wrapper = mountDialog({
      mode: 'edit',
      draft: {
        id: 'link-1',
        contest_id: 'awd-1',
        challenge_id: '101',
        title: 'Web Checker',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 1,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          put_flag: {
            method: 'PUT',
            path: '/api/flag',
            body_template: '{"flag":"{{FLAG}}"}',
            expected_status: 201,
          },
          get_flag: {
            method: 'GET',
            path: '/api/flag',
            expected_status: 200,
            expected_substring: '{{FLAG}}',
          },
          havoc: {
            method: 'GET',
            path: '/healthz',
            expected_status: 200,
          },
        },
        awd_sla_score: 18,
        awd_defense_score: 28,
        created_at: '2026-03-24T09:00:00.000Z',
      },
    })

    expect((wrapper.get('#awd-http-put-path').element as HTMLInputElement).value).toBe('/api/flag')
    expect((wrapper.get('#awd-http-put-body-template').element as HTMLTextAreaElement).value).toContain(
      '{{FLAG}}'
    )
    expect((wrapper.get('#awd-http-get-path').element as HTMLInputElement).value).toBe('/api/flag')
    expect(
      (wrapper.get('#awd-http-get-expected-substring').element as HTMLInputElement).value
    ).toBe('{{FLAG}}')
    expect((wrapper.get('#awd-http-havoc-path').element as HTMLInputElement).value).toBe('/healthz')
    expect(wrapper.get('#awd-challenge-config-preview').text()).toContain('"put_flag"')
    expect(wrapper.get('#awd-challenge-config-preview').text()).toContain('/api/flag')

    await wrapper.find('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      challenge_id: 101,
      points: 120,
      order: 1,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: {
        put_flag: {
          method: 'PUT',
          path: '/api/flag',
          body_template: '{"flag":"{{FLAG}}"}',
          expected_status: 201,
        },
        get_flag: {
          method: 'GET',
          path: '/api/flag',
          expected_status: 200,
          expected_substring: '{{FLAG}}',
        },
        havoc: {
          method: 'GET',
          path: '/healthz',
          expected_status: 200,
        },
      },
      awd_sla_score: 18,
      awd_defense_score: 28,
    })
  })

  it('应该允许通过预置模板创建 http_standard 配置', async () => {
    const wrapper = mountDialog()

    await wrapper.get('#awd-challenge-config-checker-type').setValue('http_standard')
    await wrapper.get('#awd-http-preset-rest-api').trigger('click')
    await wrapper.get('#awd-http-put-path').setValue('/flag')
    await wrapper.get('#awd-http-get-path').setValue('/flag')
    await wrapper.get('#awd-http-havoc-path').setValue('/healthz')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      challenge_id: 101,
      points: 100,
      order: 0,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: {
        put_flag: {
          method: 'PUT',
          path: '/flag',
          body_template: '{{FLAG}}',
          expected_status: 200,
        },
        get_flag: {
          method: 'GET',
          path: '/flag',
          expected_status: 200,
          expected_substring: '{{FLAG}}',
        },
        havoc: {
          method: 'GET',
          path: '/healthz',
          expected_status: 200,
        },
      },
      awd_sla_score: 0,
      awd_defense_score: 0,
    })
  })

  it('应该在 legacy_probe 模式下使用健康检查路径字段保存', async () => {
    const wrapper = mountDialog()

    await wrapper.get('#awd-challenge-config-checker-type').setValue('legacy_probe')
    await wrapper.get('#awd-challenge-config-legacy-health-path').setValue('/healthz')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      challenge_id: 101,
      points: 100,
      order: 0,
      is_visible: true,
      awd_checker_type: 'legacy_probe',
      awd_checker_config: {
        health_path: '/healthz',
      },
      awd_sla_score: 0,
      awd_defense_score: 0,
    })
  })
})
