import axios, {
  AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
} from 'axios'
import NProgress from 'nprogress'

import { useAuthStore } from '@/stores/auth'
import { mapErrorCode } from '@/utils/errorMap'
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
      if (shouldToast(error.config)) {
        toast.warning(retryMessage)
      }
      redirectToErrorStatusPage(429, error.config?.url)
      return Promise.reject(error)
    }

    if (status === 401) {
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
