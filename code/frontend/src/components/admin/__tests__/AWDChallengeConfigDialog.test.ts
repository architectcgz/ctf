import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const previewMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    runContestAWDCheckerPreview: previewMock,
  }
})

import AWDChallengeConfigDialog from '../contest/AWDChallengeConfigDialog.vue'

function mountDialog(props?: Record<string, unknown>) {
  return mount(AWDChallengeConfigDialog, {
    props: {
      contestId: 'awd-1',
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
      templateOptions: [
        {
          id: '1',
          name: 'Web 标准模板',
          slug: 'web-standard',
          category: 'web',
          difficulty: 'easy',
          description: '标准 Web AWD 服务模板',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          created_at: '2026-03-20T09:00:00.000Z',
          updated_at: '2026-03-20T09:00:00.000Z',
        },
      ],
      existingChallengeIds: [],
      draft: null,
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

describe('AWDChallengeConfigDialog', () => {
  beforeEach(() => {
    previewMock.mockReset()
  })

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
    expect(
      (wrapper.get('#awd-http-put-body-template').element as HTMLTextAreaElement).value
    ).toContain('{{FLAG}}')
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
      template_id: 1,
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
      template_id: 1,
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
      template_id: 1,
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

  it('应该允许试跑 checker 并展示结果摘要', async () => {
    previewMock.mockResolvedValue({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: {
        checker_type: 'http_standard',
        check_source: 'checker_preview',
        status_reason: 'healthy',
        checked_at: '2026-03-24T09:10:00.000Z',
        put_flag: {
          healthy: true,
          method: 'PUT',
          path: '/api/flag',
          status_code: 201,
        },
        get_flag: {
          healthy: true,
          method: 'GET',
          path: '/api/flag',
          status_code: 200,
        },
        havoc: {
          healthy: true,
          method: 'GET',
          path: '/healthz',
          status_code: 200,
        },
        targets: [
          {
            access_url: 'http://preview.internal',
            healthy: true,
            latency_ms: 42,
          },
        ],
      },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: '0',
        challenge_id: '101',
      },
      preview_token: 'preview-token-1',
    })

    const wrapper = mountDialog()

    await wrapper.get('#awd-challenge-config-checker-type').setValue('http_standard')
    await wrapper.get('#awd-http-preset-rest-api').trigger('click')
    await wrapper.get('#awd-http-put-path').setValue('/api/flag')
    await wrapper.get('#awd-http-get-path').setValue('/api/flag')
    await wrapper.get('#awd-http-havoc-path').setValue('/healthz')
    await wrapper.get('#awd-challenge-preview-access-url').setValue('http://preview.internal')
    await wrapper.get('#awd-challenge-preview-flag').setValue('flag{preview}')
    await wrapper.get('#awd-challenge-preview-submit').trigger('click')
    await flushPromises()

    expect(previewMock).toHaveBeenCalledWith('awd-1', {
      challenge_id: 101,
      checker_type: 'http_standard',
      checker_config: {
        put_flag: {
          method: 'PUT',
          path: '/api/flag',
          body_template: '{{FLAG}}',
          expected_status: 200,
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
      access_url: 'http://preview.internal',
      preview_flag: 'flag{preview}',
    })
    expect(wrapper.text()).toContain('试跑结果')
    expect(wrapper.text()).toContain('正常')
    expect(wrapper.text()).toContain('PUT Flag')
    expect(wrapper.text()).toContain('GET Flag')
    expect(wrapper.text()).toContain('http://preview.internal')

    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      challenge_id: 101,
      template_id: 1,
      points: 100,
      order: 0,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: {
        put_flag: {
          method: 'PUT',
          path: '/api/flag',
          body_template: '{{FLAG}}',
          expected_status: 200,
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
      awd_sla_score: 0,
      awd_defense_score: 0,
      awd_checker_preview_token: 'preview-token-1',
    })
  })

  it('应该在 checker 草稿变更后使试跑 token 失效', async () => {
    previewMock.mockResolvedValue({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: {
        checker_type: 'http_standard',
        check_source: 'checker_preview',
        status_reason: 'healthy',
      },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: '0',
        challenge_id: '101',
      },
      preview_token: 'preview-token-stale',
    })

    const wrapper = mountDialog()

    await wrapper.get('#awd-challenge-config-checker-type').setValue('http_standard')
    await wrapper.get('#awd-http-preset-rest-api').trigger('click')
    await wrapper.get('#awd-challenge-preview-access-url').setValue('http://preview.internal')
    await wrapper.get('#awd-challenge-preview-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#awd-http-get-path').setValue('/flag-v2')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      challenge_id: 101,
      template_id: 1,
      points: 100,
      order: 0,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: {
        put_flag: {
          method: 'PUT',
          path: '/api/flag',
          body_template: '{{FLAG}}',
          expected_status: 200,
        },
        get_flag: {
          method: 'GET',
          path: '/flag-v2',
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
    expect(wrapper.text()).toContain('需要重新试跑')
  })

  it('应该在编辑已有题目时展示最近一次已保存校验结果', async () => {
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
        },
        awd_sla_score: 18,
        awd_defense_score: 28,
        awd_checker_validation_state: 'passed',
        awd_checker_last_preview_at: '2026-03-24T09:12:00.000Z',
        awd_checker_last_preview_result: {
          checker_type: 'http_standard',
          service_status: 'up',
          check_result: {
            checker_type: 'http_standard',
            check_source: 'checker_preview',
            status_reason: 'healthy',
            checked_at: '2026-03-24T09:12:00.000Z',
          },
          preview_context: {
            access_url: 'http://saved.internal',
            preview_flag: 'flag{preview}',
            round_number: 0,
            team_id: '0',
            challenge_id: '101',
          },
        },
        created_at: '2026-03-24T09:00:00.000Z',
      },
    })

    expect(wrapper.text()).toContain('最近通过')
    expect(wrapper.text()).toContain('http://saved.internal')
    expect(wrapper.text()).toContain('配置试跑')
  })
})
