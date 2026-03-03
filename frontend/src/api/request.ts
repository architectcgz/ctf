import axios, { AxiosError, AxiosHeaders, type AxiosInstance, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'

import { useAuthStore } from '@/stores/auth'
import { mapErrorCode } from '@/utils/errorMap'
import { useToast } from '@/composables/useToast'

export interface ApiEnvelope<T> {
  code: number
  data: T
  message?: string
  request_id?: string
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

const instance = axios.create({
  baseURL,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

let isRefreshing = false
let pendingRequests: Array<{
  resolve: (value: unknown) => void
  reject: (reason?: unknown) => void
  config: AxiosRequestConfig
}> = []

function attachAuth(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
  const authStore = useAuthStore()
  if (!config.headers) {
    config.headers = new AxiosHeaders()
  }
  if (authStore.accessToken) {
    ;(config.headers as AxiosHeaders).set('Authorization', `Bearer ${authStore.accessToken}`)
  }
  return config
}

instance.interceptors.request.use(attachAuth)

async function refreshTokens(): Promise<{ access_token: string; refresh_token?: string }> {
  const authStore = useAuthStore()
  const refreshClient = axios.create({
    baseURL,
    timeout: 15000,
    headers: { 'Content-Type': 'application/json' },
  })

  const payload = authStore.refreshToken ? { refresh_token: authStore.refreshToken } : {}
  const resp = await refreshClient.post<ApiEnvelope<{ access_token: string; refresh_token?: string }>>('/auth/refresh', payload)
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

instance.interceptors.response.use(
  (response) => {
    const envelope = response.data as ApiEnvelope<unknown>
    if (typeof envelope?.code === 'number') {
      if (envelope.code === 0) return response
      const msg = mapErrorCode(envelope.code) || '请求失败'
      throw new ApiError(msg, { code: envelope.code, requestId: envelope.request_id, status: response.status })
    }
    // Non-envelope response, pass through.
    return response
  },
  async (error: AxiosError<ApiEnvelope<unknown>>) => {
    if (error instanceof ApiError) {
      toast.error(error.message)
      return Promise.reject(error)
    }

    const authStore = useAuthStore()

    const status = error.response?.status
    const code = error.response?.data?.code

    if (status === 401 && code === 11002 && !isRefreshRequest(error.config)) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingRequests.push({ resolve, reject, config: error.config || {} })
        })
      }

      isRefreshing = true
      const originalConfig = error.config || {}
      try {
        const tokens = await refreshTokens()
        authStore.updateTokens(tokens.access_token, tokens.refresh_token)

        pendingRequests.forEach(({ resolve, config }) => {
          const replayConfig = attachAuth(config as InternalAxiosRequestConfig)
          resolve(instance(replayConfig))
        })

        const replayOriginal = attachAuth(originalConfig as InternalAxiosRequestConfig)
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
      toast.error(mapped)
      return Promise.reject(new ApiError(mapped, { code, requestId: error.response?.data?.request_id, status }))
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
