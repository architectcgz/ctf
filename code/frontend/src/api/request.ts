import axios, {
  AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
} from 'axios'
import NProgress from 'nprogress'

import { mapErrorCode } from '@/utils/errorMap'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  request_id: string
  errors?: ApiValidationIssue[]
}

export interface RequestConfig extends AxiosRequestConfig {}

export interface ApiValidationIssue {
  field: string
  message: string
}

export class ApiError extends Error {
  readonly code?: number
  readonly requestId?: string
  readonly status?: number
  readonly errors?: ApiValidationIssue[]
  readonly requestUrl?: string

  constructor(
    message: string,
    opts?: {
      code?: number
      requestId?: string
      status?: number
      errors?: ApiValidationIssue[]
      requestUrl?: string
    }
  ) {
    super(message)
    this.name = 'ApiError'
    this.code = opts?.code
    this.requestId = opts?.requestId
    this.status = opts?.status
    this.errors = opts?.errors
    this.requestUrl = opts?.requestUrl
  }
}

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'
const DEFAULT_TIMEOUT = Number(import.meta.env.VITE_API_TIMEOUT) || 15000

const instance = axios.create({
  baseURL,
  timeout: DEFAULT_TIMEOUT,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})

instance.interceptors.request.use((config) => {
  NProgress.start()
  return config
})

function resolveApiMessage(
  code: number | undefined,
  message: string | undefined,
  fallbackMessage: string
): string {
  const normalizedMessage = typeof message === 'string' ? message.trim() : ''
  return normalizedMessage || mapErrorCode(code) || fallbackMessage
}

function toApiError(
  code: number | undefined,
  requestId: string | undefined,
  status: number | undefined,
  fallbackMessage: string,
  message?: string,
  errors?: ApiValidationIssue[],
  requestUrl?: string
): ApiError {
  return new ApiError(resolveApiMessage(code, message, fallbackMessage), {
    code,
    requestId,
    status,
    errors,
    requestUrl,
  })
}

instance.interceptors.response.use(
  (response) => {
    NProgress.done()
    const envelope = response.data as ApiEnvelope<unknown>
    if (typeof envelope?.code === 'number') {
      if (envelope.code === 0) return response
      const apiError = toApiError(
        envelope.code,
        envelope.request_id,
        response.status,
        '请求失败',
        envelope.message,
        envelope.errors,
        response.config?.url
      )
      return Promise.reject(apiError)
    }
    // Non-envelope response, pass through.
    return response
  },
  async (error: AxiosError<ApiEnvelope<unknown>>) => {
    NProgress.done()

    if (axios.isCancel(error) || error.code === AxiosError.ERR_CANCELED) {
      return Promise.reject(error)
    }

    const status = error.response?.status
    const code = error.response?.data?.code

    if (status === 429) {
      const retryAfter = error.response?.headers?.['retry-after']
      const retryMessage = retryAfter
        ? `请求过于频繁，请 ${retryAfter} 秒后重试`
        : '请求过于频繁，请稍后再试'
      return Promise.reject(
        toApiError(
          code,
          error.response?.data?.request_id,
          status,
          retryMessage,
          error.response?.data?.message,
          error.response?.data?.errors,
          error.config?.url
        )
      )
    }

    if (status === 401) {
      return Promise.reject(
        toApiError(
          code,
          error.response?.data?.request_id,
          status,
          '登录状态已失效，请重新登录',
          error.response?.data?.message,
          error.response?.data?.errors,
          error.config?.url
        )
      )
    }

    const mapped = mapErrorCode(code)
    if (mapped) {
      const apiError = toApiError(
        code,
        error.response?.data?.request_id,
        status,
        mapped,
        error.response?.data?.message,
        error.response?.data?.errors,
        error.config?.url
      )
      return Promise.reject(apiError)
    }

    if (!error.response) {
      return Promise.reject(
        toApiError(undefined, undefined, undefined, '网络连接失败')
      )
    }

    const fallbackMessage =
      status && status >= 500 ? '服务暂时不可用，请稍后重试' : '请求失败，请稍后重试'
    const apiError = toApiError(
      code,
      error.response?.data?.request_id,
      status,
      fallbackMessage,
      error.response?.data?.message,
      error.response?.data?.errors,
      error.config?.url
    )
    return Promise.reject(apiError)
  }
)

export async function request<T>(config: RequestConfig): Promise<T> {
  const resp = await instance.request<ApiEnvelope<T>>({
    ...config,
    signal: config.signal,
  })
  return resp.data.data
}

export function getAxiosInstance(): AxiosInstance {
  return instance
}
