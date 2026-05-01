import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const previewMock = vi.hoisted(() => vi.fn())
const awdPreviewRealtimeMocks = vi.hoisted(() => {
  let progressHandler: ((payload: Record<string, unknown>) => void) | null = null
  const start = vi.fn(async () => undefined)
  const stop = vi.fn()
  const useContestAwdPreviewRealtime = vi.fn(
    (_contestId: string, onProgress: (payload: Record<string, unknown>) => void) => {
      progressHandler = onProgress
      return {
        status: { value: 'open' },
        start,
        stop,
      }
    }
  )
  return {
    start,
    stop,
    useContestAwdPreviewRealtime,
    emitProgress(payload: Record<string, unknown>) {
      progressHandler?.(payload)
    },
    reset() {
      progressHandler = null
      start.mockClear()
      stop.mockClear()
      useContestAwdPreviewRealtime.mockClear()
    },
  }
})

vi.mock('@/api/admin/contests', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contests')>('@/api/admin/contests')
  return {
    ...actual,
    runContestAWDCheckerPreview: previewMock,
  }
})

vi.mock('@/features/awd-inspector', async () => {
  const actual =
    await vi.importActual<typeof import('@/features/awd-inspector')>('@/features/awd-inspector')
  return {
    ...actual,
    useContestAwdPreviewRealtime: awdPreviewRealtimeMocks.useContestAwdPreviewRealtime,
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
      awdChallengeOptions: [
        {
          id: '501',
          name: 'Bank Portal',
          slug: 'bank-portal',
          category: 'web',
          difficulty: 'easy',
          description: 'Bank portal template',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: '1.0.0',
          status: 'published',
          readiness_status: 'passed',
          checker_type: 'http_standard',
          checker_config: {
            put_flag: { path: '/api/flag' },
            get_flag: { path: '/api/flag' },
          },
          flag_mode: 'dynamic_team',
          flag_config: {
            flag_prefix: 'awd',
            rotate_interval_sec: 120,
          },
          defense_entry_mode: 'http',
          access_config: {
            public_base_url: 'http://bank.internal',
            service_port: 8080,
          },
          runtime_config: {
            image_id: 9901,
            service_port: 8080,
          },
          created_by: '9',
          last_verified_at: '2026-03-20T09:00:00.000Z',
          created_at: '2026-03-20T09:00:00.000Z',
          updated_at: '2026-03-20T09:00:00.000Z',
        },
      ],
      existingChallengeIds: [],
      draft: null,
      loadingChallengeCatalog: false,
      loadingAwdChallengeCatalog: false,
      saving: false,
      ...props,
    },
    global: {
      stubs: {
        AdminSurfaceModal: {
          props: ['open', 'title'],
          template: '<div v-if="open"><div>{{ title }}</div><slot /><slot name="footer" /></div>',
        },
      },
    },
  })
}

async function enableCheckerOverride(wrapper: ReturnType<typeof mountDialog>) {
  const toggle = wrapper.get('#awd-checker-override-enabled')
  if (!(toggle.element as HTMLInputElement).checked) {
    await toggle.setValue(true)
    await flushPromises()
  }
}

describe('AWDChallengeConfigDialog', () => {
  beforeEach(() => {
    vi.useRealTimers()
    previewMock.mockReset()
    awdPreviewRealtimeMocks.reset()
  })

  it('应该在创建时只展示 AWD 题库选题，并隐藏独立题目选择和题目快照', async () => {
    const wrapper = mountDialog()

    expect(wrapper.find('#awd-challenge-config-challenge').exists()).toBe(false)
    expect(wrapper.text()).toContain('AWD 题库')
    expect(wrapper.text()).toContain('Bank Portal')
    expect(wrapper.text()).not.toContain('题库模板快照')
    expect(wrapper.text()).not.toContain('public_base_url')
    expect(wrapper.text()).not.toContain('service_port')
    expect(wrapper.text()).not.toContain('flag_prefix')
    expect(wrapper.text()).not.toContain('rotate_interval_sec')
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
        awd_challenge_id: '501',
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
        awd_sla_score: 1,
        awd_defense_score: 2,
        created_at: '2026-03-24T09:00:00.000Z',
      },
    })

    expect(wrapper.text()).toContain('默认使用题目包中的 Checker')
    expect(wrapper.find('#awd-challenge-config-checker-type').exists()).toBe(false)
    expect(wrapper.find('#awd-challenge-config-preview').exists()).toBe(false)
    await enableCheckerOverride(wrapper)

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
      awd_challenge_id: 501,
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
      awd_sla_score: 1,
      awd_defense_score: 2,
    })
  })

  it('应该在新增赛事题目时继承 AWD 题目包里的 checker 配置', async () => {
    const wrapper = mountDialog()

    expect(wrapper.text()).toContain('题目包配置')
    expect(wrapper.text()).toContain('默认使用题目包中的 Checker')
    expect(wrapper.find('#awd-challenge-config-checker-type').exists()).toBe(false)
    expect(wrapper.find('#awd-checker-override-enabled').exists()).toBe(true)

    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      awd_challenge_id: 501,
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
      },
      awd_sla_score: 1,
      awd_defense_score: 2,
    })
  })

  it('应该允许通过预置模板创建 http_standard 配置', async () => {
    const wrapper = mountDialog()

    await enableCheckerOverride(wrapper)
    await wrapper.get('#awd-challenge-config-checker-type').setValue('http_standard')
    await wrapper.get('#awd-http-preset-rest-api').trigger('click')
    await wrapper.get('#awd-http-put-path').setValue('/flag')
    await wrapper.get('#awd-http-get-path').setValue('/flag')
    await wrapper.get('#awd-http-havoc-path').setValue('/healthz')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      awd_challenge_id: 501,
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
      awd_sla_score: 1,
      awd_defense_score: 2,
    })
  })

  it('应该允许编辑 tcp_standard 配置并保存结构化步骤', async () => {
    const wrapper = mountDialog()

    await enableCheckerOverride(wrapper)
    await wrapper.get('#awd-challenge-config-checker-type').setValue('tcp_standard')
    await wrapper.get('#awd-tcp-timeout-ms').setValue(5000)
    await wrapper.get('#awd-tcp-step-0-send').setValue('HELLO\n')
    await wrapper.get('#awd-tcp-step-0-expect-contains').setValue('WORLD')
    await wrapper.get('#awd-tcp-step-1-send-template').setValue('SET_FLAG {{FLAG}}\n')
    await wrapper.get('#awd-tcp-step-1-expect-contains').setValue('OK')
    await wrapper.get('#awd-tcp-step-2-send').setValue('GET_FLAG\n')
    await wrapper.get('#awd-tcp-step-2-expect-contains').setValue('{{FLAG}}')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      awd_challenge_id: 501,
      points: 100,
      order: 0,
      is_visible: true,
      awd_checker_type: 'tcp_standard',
      awd_checker_config: {
        timeout_ms: 5000,
        steps: [
          {
            send: 'HELLO\n',
            expect_contains: 'WORLD',
            timeout_ms: 3000,
          },
          {
            send_template: 'SET_FLAG {{FLAG}}\n',
            expect_contains: 'OK',
            timeout_ms: 3000,
          },
          {
            send: 'GET_FLAG\n',
            expect_contains: '{{FLAG}}',
            timeout_ms: 3000,
          },
        ],
      },
      awd_sla_score: 1,
      awd_defense_score: 2,
    })
  })

  it('应该在 legacy_probe 模式下使用健康检查路径字段保存', async () => {
    const wrapper = mountDialog()

    await enableCheckerOverride(wrapper)
    await wrapper.get('#awd-challenge-config-checker-type').setValue('legacy_probe')
    await wrapper.get('#awd-challenge-config-legacy-health-path').setValue('/healthz')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      awd_challenge_id: 501,
      points: 100,
      order: 0,
      is_visible: true,
      awd_checker_type: 'legacy_probe',
      awd_checker_config: {
        health_path: '/healthz',
      },
      awd_sla_score: 1,
      awd_defense_score: 2,
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
        awd_challenge_id: '501',
      },
      preview_token: 'preview-token-1',
    })

    const wrapper = mountDialog()

    await enableCheckerOverride(wrapper)
    await wrapper.get('#awd-challenge-config-checker-type').setValue('http_standard')
    await wrapper.get('#awd-http-preset-rest-api').trigger('click')
    await wrapper.get('#awd-http-put-path').setValue('/api/flag')
    await wrapper.get('#awd-http-get-path').setValue('/api/flag')
    await wrapper.get('#awd-http-havoc-path').setValue('/healthz')
    await wrapper.get('#awd-challenge-preview-flag').setValue('flag{preview}')
    await wrapper.get('#awd-challenge-preview-submit').trigger('click')
    await flushPromises()

    expect(previewMock).toHaveBeenCalledWith('awd-1', {
      awd_challenge_id: 501,
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
      preview_flag: 'flag{preview}',
      preview_request_id: expect.any(String),
    })
    expect(wrapper.text()).toContain('试跑结果')
    expect(wrapper.text()).toContain('正常')
    expect(wrapper.text()).toContain('PUT Flag')
    expect(wrapper.text()).toContain('GET Flag')
    expect(wrapper.text()).toContain('http://preview.internal')
    expect(wrapper.text()).toContain(
      '试跑已完成，这还是临时结果。点击下方保存按钮后，才会写入“最近一次已保存校验”。'
    )
    expect(wrapper.get('#awd-challenge-config-submit').text()).toContain('新增题目并写入试跑结果')

    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      awd_challenge_id: 501,
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
      awd_sla_score: 1,
      awd_defense_score: 2,
      awd_checker_preview_token: 'preview-token-1',
    })
  })

  it('应该在试跑期间展示三轮进度并禁止保存', async () => {
    let resolvePreview: ((value: Record<string, unknown>) => void) | null = null
    previewMock.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolvePreview = resolve as (value: Record<string, unknown>) => void
        })
    )

    const wrapper = mountDialog()

    await wrapper.get('#awd-challenge-preview-submit').trigger('click')
    await flushPromises()

    expect(wrapper.get('#awd-challenge-preview-submit').text()).toContain('试跑中')
    expect(wrapper.get('#awd-challenge-config-submit').attributes('disabled')).toBeDefined()
    expect(wrapper.get('#awd-challenge-preview-progress').text()).toContain('正在试跑 Checker')
    expect(wrapper.get('#awd-challenge-preview-progress').text()).toContain('共 3 轮试跑')
    expect(wrapper.get('#awd-challenge-preview-progress').text()).toContain('已接入实时进度事件')

    const previewRequestId = (
      previewMock.mock.calls[0]?.[1] as { preview_request_id?: string } | undefined
    )?.preview_request_id
    expect(previewRequestId).toBeTruthy()

    awdPreviewRealtimeMocks.emitProgress({
      preview_request_id: 'ignored-request',
      phase_key: 'attempt-3',
      phase_label: '第 3 轮试跑',
      detail: '这条事件不应污染当前请求。',
      attempt: 3,
      total_attempts: 3,
      status: 'running',
    })
    await flushPromises()
    expect(wrapper.get('#awd-challenge-preview-progress-status').text()).not.toContain(
      '第 3 轮试跑'
    )

    awdPreviewRealtimeMocks.emitProgress({
      preview_request_id: previewRequestId,
      phase_key: 'attempt-1',
      phase_label: '第 1 轮试跑',
      detail: '正在执行第 1 / 3 轮请求校验。',
      attempt: 1,
      total_attempts: 3,
      status: 'running',
    })
    await flushPromises()
    expect(wrapper.get('#awd-challenge-preview-progress-status').text()).toContain('第 1 轮试跑')
    expect(wrapper.get('#awd-challenge-preview-progress').text()).toContain(
      '正在执行第 1 / 3 轮请求校验。'
    )
    expect(wrapper.get('#awd-challenge-preview-progress').text()).toContain('第 1 / 3 轮')

    awdPreviewRealtimeMocks.emitProgress({
      preview_request_id: previewRequestId,
      phase_key: 'attempt-2',
      phase_label: '第 2 轮试跑',
      detail: '正在执行第 2 / 3 轮请求校验。',
      attempt: 2,
      total_attempts: 3,
      status: 'running',
    })
    await flushPromises()
    expect(wrapper.get('#awd-challenge-preview-progress-status').text()).toContain('第 2 轮试跑')
    expect(wrapper.get('#awd-challenge-preview-progress').text()).toContain('第 2 / 3 轮')

    if (!resolvePreview) {
      throw new Error('preview promise resolver was not captured')
    }

    const finishPreview = resolvePreview as (value: Record<string, unknown>) => void

    finishPreview({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: {
        checker_type: 'http_standard',
        check_source: 'checker_preview',
        status_reason: 'preview_quorum_passed',
        preview_pass_count: 2,
        preview_total_count: 3,
      },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: '0',
        challenge_id: '101',
      },
      preview_token: 'preview-token-progress',
    })
    await flushPromises()

    expect(wrapper.find('#awd-challenge-preview-progress').exists()).toBe(false)
    expect(wrapper.get('#awd-challenge-config-submit').attributes('disabled')).toBeUndefined()
    expect(wrapper.text()).toContain('2/3 通过')
  })

  it('应该在编辑模式下明确区分临时试跑结果和已保存校验', async () => {
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
      preview_token: 'preview-token-2',
    })

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
        awd_service_id: '2',
        awd_challenge_id: '501',
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          put_flag: {
            method: 'PUT',
            path: '/api/flag',
            expected_status: 200,
            body_template: '{{FLAG}}',
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
        awd_sla_score: 1,
        awd_defense_score: 2,
        awd_checker_validation_state: 'pending',
        awd_checker_last_preview_result: undefined,
        created_at: '2026-03-24T09:00:00.000Z',
      },
    })

    await wrapper.get('#awd-challenge-preview-submit').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain(
      '当前试跑结果尚未保存。点击下方保存按钮后，这里会显示最新一次已保存校验。'
    )
    expect(wrapper.text()).toContain(
      '试跑已完成，这还是临时结果。点击下方保存按钮后，才会写入“最近一次已保存校验”。'
    )
    expect(wrapper.get('#awd-challenge-config-submit').text()).toContain('保存配置并写入试跑结果')
    expect(previewMock).toHaveBeenCalledWith(
      'awd-1',
      expect.objectContaining({
        service_id: 2,
      })
    )
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

    await enableCheckerOverride(wrapper)
    await wrapper.get('#awd-challenge-config-checker-type').setValue('http_standard')
    await wrapper.get('#awd-http-preset-rest-api').trigger('click')
    await wrapper.get('#awd-challenge-preview-submit').trigger('click')
    await flushPromises()

    await wrapper.get('#awd-http-get-path').setValue('/flag-v2')
    await wrapper.get('#awd-challenge-config-submit').trigger('click')

    expect(wrapper.emitted('save')?.[0]?.[0]).toEqual({
      awd_challenge_id: 501,
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
      awd_sla_score: 1,
      awd_defense_score: 2,
    })
    expect(wrapper.text()).toContain('需要重新试跑')
  })

  it('应该把预览实例镜像拉取错误翻译成可执行提示', async () => {
    previewMock.mockRejectedValue(
      new Error(
        'Error response from daemon: failed to resolve reference "registry.example.edu/ctf/awd-bank-portal:v1": failed to do request: Head "https://registry.example.edu/v2/ctf/awd-bank-portal/manifests/v1": EOF'
      )
    )

    const wrapper = mountDialog()

    await wrapper.get('#awd-challenge-preview-submit').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain(
      '自动拉起预览实例失败：当前 AWD 题目引用的运行镜像暂时无法拉取。'
    )
    expect(wrapper.text()).toContain(
      '如果这是示例占位地址，请先在当前环境构建同名镜像，或把题目镜像改成可直接拉取的真实地址。'
    )
    expect(wrapper.text()).toContain('registry.example.edu/ctf/awd-bank-portal:v1')
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
        awd_challenge_id: '501',
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
        awd_sla_score: 1,
        awd_defense_score: 2,
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
