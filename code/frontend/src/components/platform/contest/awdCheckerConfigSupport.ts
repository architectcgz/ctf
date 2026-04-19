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

export function buildCheckerConfigPreview(
  checkerType: AWDCheckerType,
  drafts: {
    legacyProbeDraft: AWDLegacyProbeDraft
    httpStandardDraft: AWDHTTPStandardDraft
  }
): Record<string, unknown> {
  return checkerType === 'http_standard'
    ? buildHTTPStandardCheckerConfig(drafts.httpStandardDraft, false).config
    : buildLegacyProbeCheckerConfig(drafts.legacyProbeDraft).config
}
