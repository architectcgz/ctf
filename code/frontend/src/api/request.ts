import axios, {
  AxiosError,
  AxiosHeaders,
  type AxiosInstance,
  type AxiosRequestConfig,
  type InternalAxiosRequestConfig,
} from 'axios'
import NProgress from 'nprogress'

import { useAuthStore } from '@/stores/auth'
import { AUTH_ERROR_CODES, mapErrorCode } from '@/utils/errorMap'
import { useToast } from '@/composables/useToast'
import { redirectToErrorStatusPage, shouldRedirectToErrorStatusPage } from '@/utils/errorStatusPage'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  request_id: string
  errors?: ApiValidationIssue[]
}

export interface RequestConfig extends AxiosRequestConfig {
  suppressErrorToast?: boolean
}

export interface ApiValidationIssue {
  field: string
  message: string
}

export class ApiError extends Error {
  readonly code?: number
  readonly requestId?: string
  readonly status?: number
  readonly errors?: ApiValidationIssue[]

  constructor(
    message: string,
    opts?: { code?: number; requestId?: string; status?: number; errors?: ApiValidationIssue[] }
  ) {
    super(message)
    this.name = 'ApiError'
    this.code = opts?.code
    this.requestId = opts?.requestId
    this.status = opts?.status
    this.errors = opts?.errors
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

let isRefreshing = false
let pendingRequests: Array<{
  resolve: (value: unknown) => void
  reject: (reason?: unknown) => void
  config: AxiosRequestConfig
}> = []

function normalizeHeaders(headers?: AxiosRequestConfig['headers']): AxiosHeaders {
  return AxiosHeaders.from((headers ?? {}) as any)
}

function attachAuth<T extends AxiosRequestConfig>(config: T): T {
  const authStore = useAuthStore()
  config.headers = normalizeHeaders(config.headers)
  if (authStore.accessToken) {
    ;(config.headers as AxiosHeaders).set('Authorization', `Bearer ${authStore.accessToken}`)
  }
  return config
}

instance.interceptors.request.use((config) => {
  NProgress.start()
  return attachAuth(config as InternalAxiosRequestConfig)
})

async function refreshTokens(): Promise<{ access_token: string }> {
  const refreshClient = axios.create({
    baseURL,
    timeout: DEFAULT_TIMEOUT,
    withCredentials: true,
    headers: { 'Content-Type': 'application/json' },
  })

  const resp = await refreshClient.post<
    ApiEnvelope<{ access_token: string; token_type?: string; expires_in?: number }>
  >('/auth/refresh', {})
  if (resp.data?.code !== 0) {
    throw new ApiError(mapErrorCode(resp.data?.code) || '登录已过期，请重新登录', {
      code: resp.data?.code,
      requestId: resp.data?.request_id,
      status: resp.status,
    })
  }
  return resp.data.data
}

function isRefreshRequest(config: AxiosRequestConfig | undefined): boolean {
  const url = config?.url || ''
  return typeof url === 'string' && url.includes('/auth/refresh')
}

function humanizeError(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (axios.isAxiosError(err)) {
    return err.message || '网络连接失败'
  }
  if (err instanceof Error) return err.message || '请求失败'
  return '请求失败'
}

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
  errors?: ApiValidationIssue[]
): ApiError {
  return new ApiError(resolveApiMessage(code, message, fallbackMessage), {
    code,
    requestId,
    status,
    errors,
  })
}

function getToast() {
  return useToast()
}

function shouldToast(config?: AxiosRequestConfig): boolean {
  return !(config as RequestConfig | undefined)?.suppressErrorToast
}

instance.interceptors.response.use(
  (response) => {
    const toast = getToast()
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
        envelope.errors
      )
      if (shouldToast(response.config)) {
        toast.error(apiError.message)
      }
      return Promise.reject(apiError)
    }
    // Non-envelope response, pass through.
    return response
  },
  async (error: AxiosError<ApiEnvelope<unknown>>) => {
    const toast = getToast()
    NProgress.done()
    const authStore = useAuthStore()

    const status = error.response?.status
    const code = error.response?.data?.code

    if (
      status === 401 &&
      code === AUTH_ERROR_CODES.ACCESS_TOKEN_EXPIRED &&
      !isRefreshRequest(error.config)
    ) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingRequests.push({ resolve, reject, config: error.config || {} })
        })
      }

      isRefreshing = true
      const originalConfig = error.config || {}
      try {
        const tokens = await refreshTokens()
        authStore.updateTokens(tokens.access_token)

        pendingRequests.forEach(({ resolve, config }) => {
          const replayConfig = attachAuth(config)
          resolve(instance(replayConfig))
        })

        const replayOriginal = attachAuth(originalConfig)
        return instance(replayOriginal)
      } catch (refreshErr) {
        pendingRequests.forEach(({ reject }) => reject(refreshErr))
        authStore.logout()
        toast.error(humanizeError(refreshErr))
        redirectToErrorStatusPage(401, error.config?.url)
        return Promise.reject(refreshErr)
      } finally {
        isRefreshing = false
        pendingRequests = []
      }
    }

    if (status === 429) {
      const retryAfter = error.response?.headers?.['retry-after']
      const retryMessage = retryAfter
        ? `请求过于频繁，请 ${retryAfter} 秒后重试`
        : '请求过于频繁，请稍后再试'
      if (shouldToast(error.config)) {
        toast.warning(retryMessage)
      }
      redirectToErrorStatusPage(429, error.config?.url)
      return Promise.reject(error)
    }

    if (status === 401 && !isRefreshRequest(error.config)) {
      const unauthorizedError = toApiError(
        code,
        error.response?.data?.request_id,
        status,
        '登录状态已失效，请重新登录',
        error.response?.data?.message,
        error.response?.data?.errors
      )
      if (shouldToast(error.config)) {
        toast.error(unauthorizedError.message)
      }
      if (shouldRedirectToErrorStatusPage(status, error.config?.url)) {
        authStore.logout()
        redirectToErrorStatusPage(status, error.config?.url)
      }
      return Promise.reject(unauthorizedError)
    }

    const mapped = mapErrorCode(code)
    if (mapped) {
      const apiError = toApiError(
        code,
        error.response?.data?.request_id,
        status,
        mapped,
        error.response?.data?.message,
        error.response?.data?.errors
      )
      if (shouldToast(error.config)) {
        toast.error(apiError.message)
      }
      return Promise.reject(apiError)
    }

    if (!error.response) {
      if (shouldToast(error.config)) {
        toast.error('网络连接失败')
      }
      return Promise.reject(error)
    }

    const fallbackMessage =
      status && status >= 500 ? '服务暂时不可用，请稍后重试' : '请求失败，请稍后重试'
    const apiError = toApiError(
      code,
      error.response?.data?.request_id,
      status,
      fallbackMessage,
      error.response?.data?.message,
      error.response?.data?.errors
    )
    if (shouldToast(error.config)) {
      toast.error(apiError.message)
    }
    redirectToErrorStatusPage(status, error.config?.url)
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
