import type { AWDCheckerType } from '@/api/contracts'

export const AWD_HTTP_METHOD_OPTIONS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'] as const

export const AWD_CHECKER_FIELD_ERROR_KEYS = [
  'legacy_health_path',
  'http_put_path',
  'http_put_expected_status',
  'http_put_headers_text',
  'http_get_path',
  'http_get_expected_status',
  'http_get_headers_text',
  'http_havoc_expected_status',
  'http_havoc_headers_text',
  'tcp_timeout',
  'tcp_steps',
  'script_entry',
  'script_timeout',
  'script_args_text',
  'script_env_text',
] as const

export type AWDCheckerFieldErrorKey = (typeof AWD_CHECKER_FIELD_ERROR_KEYS)[number]

export interface AWDLegacyProbeDraft {
  health_path: string
}

export interface AWDHTTPActionDraft {
  method: string
  path: string
  expected_status: number
  headers_text: string
  body_template: string
  expected_substring: string
}

export interface AWDHTTPStandardDraft {
  put_flag: AWDHTTPActionDraft
  get_flag: AWDHTTPActionDraft
  havoc: AWDHTTPActionDraft
}

export interface AWDScriptCheckerDraft {
  runtime: string
  entry: string
  timeout_sec: number
  args_text: string
  env_text: string
  output: 'exit_code' | 'json'
}

export interface AWDTCPCheckerStepDraft {
  send: string
  send_template: string
  send_hex: string
  expect_contains: string
  expect_regex: string
  timeout_ms: number
}

export interface AWDTCPStandardDraft {
  timeout_ms: number
  steps: AWDTCPCheckerStepDraft[]
}

export interface AWDHTTPStandardPreset {
  id: string
  label: string
  description: string
  draft: AWDHTTPStandardDraft
}

export interface AWDCheckerBuildResult {
  config: Record<string, unknown>
  errors: Partial<Record<AWDCheckerFieldErrorKey, string>>
}

const DEFAULT_PUT_BODY_TEMPLATE = '{{FLAG}}'
const DEFAULT_GET_EXPECTED_SUBSTRING = '{{FLAG}}'

export const AWD_HTTP_STANDARD_PRESETS: AWDHTTPStandardPreset[] = [
  {
    id: 'rest-api',
    label: 'REST /api/flag',
    description: '适合用同一路径执行 PUT/GET，再补一个轻量 health check。',
    draft: {
      put_flag: {
        method: 'PUT',
        path: '/api/flag',
        expected_status: 200,
        headers_text: '',
        body_template: DEFAULT_PUT_BODY_TEMPLATE,
        expected_substring: '',
      },
      get_flag: {
        method: 'GET',
        path: '/api/flag',
        expected_status: 200,
        headers_text: '',
        body_template: '',
        expected_substring: DEFAULT_GET_EXPECTED_SUBSTRING,
      },
      havoc: {
        method: 'GET',
        path: '/healthz',
        expected_status: 200,
        headers_text: '',
        body_template: '',
        expected_substring: '',
      },
    },
  },
  {
    id: 'form-flag',
    label: 'Form /flag',
    description: '适合表单提交 flag，再通过 GET 页面或接口回读结果。',
    draft: {
      put_flag: {
        method: 'POST',
        path: '/flag',
        expected_status: 200,
        headers_text: JSON.stringify(
          {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          null,
          2
        ),
        body_template: 'flag={{FLAG}}',
        expected_substring: '',
      },
      get_flag: {
        method: 'GET',
        path: '/flag',
        expected_status: 200,
        headers_text: '',
        body_template: '',
        expected_substring: DEFAULT_GET_EXPECTED_SUBSTRING,
      },
      havoc: {
        method: 'GET',
        path: '/healthz',
        expected_status: 200,
        headers_text: '',
        body_template: '',
        expected_substring: '',
      },
    },
  },
  {
    id: 'file-flag',
    label: 'File /flag.txt',
    description: '适合直接写入和读取文本文件接口的简单服务。',
    draft: {
      put_flag: {
        method: 'PUT',
        path: '/flag.txt',
        expected_status: 200,
        headers_text: '',
        body_template: DEFAULT_PUT_BODY_TEMPLATE,
        expected_substring: '',
      },
      get_flag: {
        method: 'GET',
        path: '/flag.txt',
        expected_status: 200,
        headers_text: '',
        body_template: '',
        expected_substring: DEFAULT_GET_EXPECTED_SUBSTRING,
      },
      havoc: {
        method: 'GET',
        path: '/',
        expected_status: 200,
        headers_text: '',
        body_template: '',
        expected_substring: '',
      },
    },
  },
]

function createDefaultActionDraft(
  defaultMethod: string,
  options?: Partial<AWDHTTPActionDraft>
): AWDHTTPActionDraft {
  return {
    method: options?.method || defaultMethod,
    path: options?.path || '',
    expected_status: options?.expected_status ?? 200,
    headers_text: options?.headers_text || '',
    body_template: options?.body_template || '',
    expected_substring: options?.expected_substring || '',
  }
}

function readRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}
}

function readString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function readPositiveInteger(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isInteger(value) && value > 0 ? value : fallback
}

function stringifyHeaders(value: unknown): string {
  const headers = readRecord(value)
  return Object.keys(headers).length > 0 ? JSON.stringify(headers, null, 2) : ''
}

function stringifyStringRecord(value: unknown): string {
  const record = readRecord(value)
  return Object.keys(record).length > 0 ? JSON.stringify(record, null, 2) : ''
}

function normalizeHTTPMethod(value: string, fallback: string): string {
  const method = value.trim().toUpperCase()
  return AWD_HTTP_METHOD_OPTIONS.includes(method as (typeof AWD_HTTP_METHOD_OPTIONS)[number])
    ? method
    : fallback
}

function parseHeadersText(
  value: string,
  errorKey: AWDCheckerFieldErrorKey,
  errors: Partial<Record<AWDCheckerFieldErrorKey, string>>,
  strict: boolean
): Record<string, string> {
  const trimmed = value.trim()
  if (!trimmed) {
    return {}
  }

  try {
    const parsed = JSON.parse(trimmed)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      if (strict) {
        errors[errorKey] = 'Headers 必须是 JSON 对象'
      }
      return {}
    }

    return Object.fromEntries(
      Object.entries(parsed as Record<string, unknown>).map(([key, headerValue]) => [
        key,
        String(headerValue),
      ])
    )
  } catch {
    if (strict) {
      errors[errorKey] = 'Headers 必须是合法 JSON'
    }
    return {}
  }
}

function buildActionConfig(
  action: AWDHTTPActionDraft,
  options: {
    defaultMethod: string
    pathErrorKey?: AWDCheckerFieldErrorKey
    statusErrorKey: AWDCheckerFieldErrorKey
    headersErrorKey: AWDCheckerFieldErrorKey
    defaultBodyTemplate?: string
    defaultExpectedSubstring?: string
    optionalPath?: boolean
    strict: boolean
  },
  errors: Partial<Record<AWDCheckerFieldErrorKey, string>>
): Record<string, unknown> | null {
  const path = action.path.trim()
  if (!path) {
    if (options.strict && options.pathErrorKey && !options.optionalPath) {
      errors[options.pathErrorKey] = '路径不能为空'
    }
    if (options.optionalPath) {
      return null
    }
  }

  if (!Number.isInteger(action.expected_status) || action.expected_status <= 0) {
    if (options.strict) {
      errors[options.statusErrorKey] = '期望状态码必须是大于 0 的整数'
    }
  }

  const result: Record<string, unknown> = {
    method: normalizeHTTPMethod(action.method, options.defaultMethod),
    path,
    expected_status:
      Number.isInteger(action.expected_status) && action.expected_status > 0
        ? action.expected_status
        : 200,
  }

  const headers = parseHeadersText(
    action.headers_text,
    options.headersErrorKey,
    errors,
    options.strict
  )
  if (Object.keys(headers).length > 0) {
    result.headers = headers
  }

  if (options.defaultBodyTemplate) {
    result.body_template = action.body_template.trim() || options.defaultBodyTemplate
  }

  if (options.defaultExpectedSubstring) {
    result.expected_substring = action.expected_substring.trim() || options.defaultExpectedSubstring
  }

  return result
}

export function createLegacyProbeDraft(
  config?: Record<string, unknown> | null
): AWDLegacyProbeDraft {
  return {
    health_path: readString(config?.health_path).trim(),
  }
}

export function createHTTPStandardDraft(
  config?: Record<string, unknown> | null
): AWDHTTPStandardDraft {
  const putFlag = readRecord(config?.put_flag)
  const getFlag = readRecord(config?.get_flag)
  const havoc = readRecord(config?.havoc)

  return {
    put_flag: createDefaultActionDraft('PUT', {
      method: readString(putFlag.method) || 'PUT',
      path: readString(putFlag.path).trim(),
      expected_status: readPositiveInteger(putFlag.expected_status, 200),
      headers_text: stringifyHeaders(putFlag.headers),
      body_template: readString(putFlag.body_template) || DEFAULT_PUT_BODY_TEMPLATE,
    }),
    get_flag: createDefaultActionDraft('GET', {
      method: readString(getFlag.method) || 'GET',
      path: readString(getFlag.path).trim(),
      expected_status: readPositiveInteger(getFlag.expected_status, 200),
      headers_text: stringifyHeaders(getFlag.headers),
      expected_substring: readString(getFlag.expected_substring) || DEFAULT_GET_EXPECTED_SUBSTRING,
    }),
    havoc: createDefaultActionDraft('GET', {
      method: readString(havoc.method) || 'GET',
      path: readString(havoc.path).trim(),
      expected_status: readPositiveInteger(havoc.expected_status, 200),
      headers_text: stringifyHeaders(havoc.headers),
    }),
  }
}

export function cloneHTTPStandardDraft(draft: AWDHTTPStandardDraft): AWDHTTPStandardDraft {
  return {
    put_flag: { ...draft.put_flag },
    get_flag: { ...draft.get_flag },
    havoc: { ...draft.havoc },
  }
}

export function createScriptCheckerDraft(
  config?: Record<string, unknown> | null
): AWDScriptCheckerDraft {
  const output = readString(config?.output) === 'json' ? 'json' : 'exit_code'
  const args = Array.isArray(config?.args)
    ? (config.args as unknown[]).map((item) => String(item)).join('\n')
    : '{{TARGET_URL}}'
  return {
    runtime: readString(config?.runtime) || 'python3',
    entry: readString(config?.entry) || 'docker/check/check.py',
    timeout_sec: readPositiveInteger(config?.timeout_sec, 10),
    args_text: args,
    env_text: stringifyStringRecord(config?.env),
    output,
  }
}

function createTCPCheckerStepDraft(config?: Record<string, unknown> | null): AWDTCPCheckerStepDraft {
  return {
    send: readString(config?.send),
    send_template: readString(config?.send_template),
    send_hex: readString(config?.send_hex),
    expect_contains: readString(config?.expect_contains),
    expect_regex: readString(config?.expect_regex),
    timeout_ms: readPositiveInteger(config?.timeout_ms, 3000),
  }
}

export function createTCPStandardDraft(
  config?: Record<string, unknown> | null
): AWDTCPStandardDraft {
  const steps = Array.isArray(config?.steps)
    ? (config.steps as unknown[])
        .map((item) => createTCPCheckerStepDraft(readRecord(item)))
        .filter(
          (step) =>
            step.send ||
            step.send_template ||
            step.send_hex ||
            step.expect_contains ||
            step.expect_regex
        )
    : []

  return {
    timeout_ms: readPositiveInteger(config?.timeout_ms, 3000),
    steps:
      steps.length > 0
        ? steps
        : [
            createTCPCheckerStepDraft({
              send: 'PING\n',
              expect_contains: 'PONG',
            }),
            createTCPCheckerStepDraft({
              send_template: 'SET_FLAG {{FLAG}}\n',
              expect_contains: 'OK',
            }),
            createTCPCheckerStepDraft({
              send: 'GET_FLAG\n',
              expect_contains: '{{FLAG}}',
            }),
          ],
  }
}

export function getHTTPStandardPresetDraft(presetId: string): AWDHTTPStandardDraft {
  const preset = AWD_HTTP_STANDARD_PRESETS.find((item) => item.id === presetId)
  return cloneHTTPStandardDraft(preset?.draft || AWD_HTTP_STANDARD_PRESETS[0].draft)
}

export function buildLegacyProbeCheckerConfig(draft: AWDLegacyProbeDraft): AWDCheckerBuildResult {
  const healthPath = draft.health_path.trim()
  return {
    config: healthPath ? { health_path: healthPath } : {},
    errors: {},
  }
}

export function buildHTTPStandardCheckerConfig(
  draft: AWDHTTPStandardDraft,
  strict = true
): AWDCheckerBuildResult {
  const errors: Partial<Record<AWDCheckerFieldErrorKey, string>> = {}
  const putFlag = buildActionConfig(
    draft.put_flag,
    {
      defaultMethod: 'PUT',
      pathErrorKey: 'http_put_path',
      statusErrorKey: 'http_put_expected_status',
      headersErrorKey: 'http_put_headers_text',
      defaultBodyTemplate: DEFAULT_PUT_BODY_TEMPLATE,
      strict,
    },
    errors
  )
  const getFlag = buildActionConfig(
    draft.get_flag,
    {
      defaultMethod: 'GET',
      pathErrorKey: 'http_get_path',
      statusErrorKey: 'http_get_expected_status',
      headersErrorKey: 'http_get_headers_text',
      defaultExpectedSubstring: DEFAULT_GET_EXPECTED_SUBSTRING,
      strict,
    },
    errors
  )
  const havoc = buildActionConfig(
    draft.havoc,
    {
      defaultMethod: 'GET',
      statusErrorKey: 'http_havoc_expected_status',
      headersErrorKey: 'http_havoc_headers_text',
      optionalPath: true,
      strict,
    },
    errors
  )

  const config: Record<string, unknown> = {}
  if (putFlag) {
    config.put_flag = putFlag
  }
  if (getFlag) {
    config.get_flag = getFlag
  }
  if (havoc) {
    config.havoc = havoc
  }

  return {
    config,
    errors,
  }
}

export function buildScriptCheckerConfig(
  draft: AWDScriptCheckerDraft,
  strict = true
): AWDCheckerBuildResult {
  const errors: Partial<Record<AWDCheckerFieldErrorKey, string>> = {}
  const entry = draft.entry.trim()
  if (!entry && strict) {
    errors.script_entry = '入口文件不能为空'
  }
  if (entry.startsWith('/') || entry.includes('..')) {
    errors.script_entry = '入口文件必须是题目包内相对路径'
  }
  if (!Number.isInteger(draft.timeout_sec) || draft.timeout_sec <= 0 || draft.timeout_sec > 60) {
    errors.script_timeout = '超时时间必须是 1-60 秒'
  }

  const args = draft.args_text
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)

  let env: Record<string, string> = {}
  const envText = draft.env_text.trim()
  if (envText) {
    try {
      const parsed = JSON.parse(envText)
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
        errors.script_env_text = '环境变量必须是 JSON 对象'
      } else {
        env = Object.fromEntries(
          Object.entries(parsed as Record<string, unknown>).map(([key, value]) => [
            key,
            String(value),
          ])
        )
      }
    } catch {
      errors.script_env_text = '环境变量必须是合法 JSON'
    }
  }

  return {
    config: {
      runtime: draft.runtime.trim() || 'python3',
      entry,
      timeout_sec:
        Number.isInteger(draft.timeout_sec) && draft.timeout_sec > 0 ? draft.timeout_sec : 10,
      args,
      ...(Object.keys(env).length > 0 ? { env } : {}),
      output: draft.output,
    },
    errors,
  }
}

export function buildTCPStandardCheckerConfig(
  draft: AWDTCPStandardDraft,
  strict = true
): AWDCheckerBuildResult {
  const errors: Partial<Record<AWDCheckerFieldErrorKey, string>> = {}
  if (!Number.isInteger(draft.timeout_ms) || draft.timeout_ms <= 0 || draft.timeout_ms > 60000) {
    errors.tcp_timeout = '超时时间必须是 1-60000 毫秒'
  }

  const steps = draft.steps
    .map((step) => ({
      send: step.send,
      send_template: step.send_template,
      send_hex: step.send_hex.trim(),
      expect_contains: step.expect_contains,
      expect_regex: step.expect_regex,
      timeout_ms: step.timeout_ms,
    }))
    .filter(
      (step) =>
        step.send ||
        step.send_template ||
        step.send_hex ||
        step.expect_contains ||
        step.expect_regex
    )

  if (steps.length === 0 && strict) {
    errors.tcp_steps = '至少需要一个 TCP 步骤'
  }

  for (const step of steps) {
    const sendFieldCount = [step.send, step.send_template, step.send_hex].filter(Boolean).length
    const hasExpectation = Boolean(step.expect_contains || step.expect_regex)
    if (sendFieldCount === 0 && !hasExpectation && strict) {
      errors.tcp_steps = '每个步骤至少需要发送内容或期望结果'
    }
    if (step.send_hex && (step.send || step.send_template) && strict) {
      errors.tcp_steps = 'send_hex 不能与 send 或 send_template 同时填写'
    }
    if (step.send && step.send_template && strict) {
      errors.tcp_steps = 'send 不能与 send_template 同时填写'
    }
    if (
      !Number.isInteger(step.timeout_ms) ||
      step.timeout_ms < 0 ||
      step.timeout_ms > 60000
    ) {
      errors.tcp_steps = '步骤超时时间必须是 0-60000 毫秒'
    }
    if (step.expect_regex) {
      try {
        new RegExp(step.expect_regex)
      } catch {
        errors.tcp_steps = 'expect_regex 必须是合法正则'
      }
    }
  }

  return {
    config: {
      timeout_ms:
        Number.isInteger(draft.timeout_ms) && draft.timeout_ms > 0 ? draft.timeout_ms : 3000,
      steps: steps.map((step) => ({
        ...(step.send ? { send: step.send } : {}),
        ...(step.send_template ? { send_template: step.send_template } : {}),
        ...(step.send_hex ? { send_hex: step.send_hex } : {}),
        ...(step.expect_contains ? { expect_contains: step.expect_contains } : {}),
        ...(step.expect_regex ? { expect_regex: step.expect_regex } : {}),
        ...(Number.isInteger(step.timeout_ms) && step.timeout_ms > 0
          ? { timeout_ms: step.timeout_ms }
          : {}),
      })),
    },
    errors,
  }
}

export function buildCheckerConfigPreview(
  checkerType: AWDCheckerType,
  drafts: {
    legacyProbeDraft: AWDLegacyProbeDraft
    httpStandardDraft: AWDHTTPStandardDraft
    tcpStandardDraft?: AWDTCPStandardDraft
    scriptCheckerDraft?: AWDScriptCheckerDraft
  }
): Record<string, unknown> {
  switch (checkerType) {
    case 'http_standard':
      return buildHTTPStandardCheckerConfig(drafts.httpStandardDraft, false).config
    case 'tcp_standard':
      return buildTCPStandardCheckerConfig(
        drafts.tcpStandardDraft || createTCPStandardDraft(),
        false
      ).config
    case 'script_checker':
      return buildScriptCheckerConfig(
        drafts.scriptCheckerDraft || createScriptCheckerDraft(),
        false
      ).config
    default:
      return buildLegacyProbeCheckerConfig(drafts.legacyProbeDraft).config
  }
}
