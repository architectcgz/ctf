import axios, { AxiosError, AxiosHeaders, type AxiosInstance, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'

import { useAuthStore } from '@/stores/auth'
import { AUTH_ERROR_CODES, mapErrorCode } from '@/utils/errorMap'
import { useToast } from '@/composables/useToast'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  request_id: string
  errors?: Array<{ field: string; message: string }>
}

export class ApiError extends Error {
  readonly code?: number
  readonly requestId?: string
  readonly status?: number

  constructor(message: string, opts?: { code?: number; requestId?: string; status?: number }) {
    super(message)
    this.name = 'ApiError'
    this.code = opts?.code
    this.requestId = opts?.requestId
    this.status = opts?.status
  }
}

const toast = useToast()

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

instance.interceptors.request.use((config) => attachAuth(config as InternalAxiosRequestConfig))

async function refreshTokens(): Promise<{ access_token: string }> {
  const refreshClient = axios.create({
    baseURL,
    timeout: DEFAULT_TIMEOUT,
    withCredentials: true,
    headers: { 'Content-Type': 'application/json' },
  })

  const resp = await refreshClient.post<ApiEnvelope<{ access_token: string; token_type?: string; expires_in?: number }>>(
    '/auth/refresh',
    {}
  )
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

function toApiError(
  code: number | undefined,
  requestId: string | undefined,
  status: number | undefined,
  fallbackMessage: string
): ApiError {
  return new ApiError(mapErrorCode(code) || fallbackMessage, {
    code,
    requestId,
    status,
  })
}

instance.interceptors.response.use(
  (response) => {
    const envelope = response.data as ApiEnvelope<unknown>
    if (typeof envelope?.code === 'number') {
      if (envelope.code === 0) return response
      const apiError = toApiError(envelope.code, envelope.request_id, response.status, '请求失败')
      toast.error(apiError.message)
      return Promise.reject(apiError)
    }
    // Non-envelope response, pass through.
    return response
  },
  async (error: AxiosError<ApiEnvelope<unknown>>) => {
    const authStore = useAuthStore()

    const status = error.response?.status
    const code = error.response?.data?.code

    if (status === 401 && code === AUTH_ERROR_CODES.ACCESS_TOKEN_EXPIRED && !isRefreshRequest(error.config)) {
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
        window.location.assign('/login')
        return Promise.reject(refreshErr)
      } finally {
        isRefreshing = false
        pendingRequests = []
      }
    }

    if (status === 429) {
      toast.warning('请求过于频繁，请稍后再试')
      return Promise.reject(error)
    }

    const mapped = mapErrorCode(code)
    if (mapped) {
      const apiError = toApiError(code, error.response?.data?.request_id, status, mapped)
      toast.error(apiError.message)
      return Promise.reject(apiError)
    }

    toast.error('网络连接失败')
    return Promise.reject(error)
  }
)

export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const resp = await instance.request<ApiEnvelope<T>>(config)
  return resp.data.data
}

export function getAxiosInstance(): AxiosInstance {
  return instance
}
