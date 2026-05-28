import type { App, ComponentPublicInstance } from 'vue'
import type { Pinia } from 'pinia'
import type { Router } from 'vue-router'

import { ApiError, getAxiosInstance } from '@/api/request'
import { useAuthStore } from '@/stores/auth'
import { redirectToErrorStatusPage, shouldRedirectToErrorStatusPage } from '@/utils/errorStatusPage'

let httpErrorHandlingInstalled = false

function resolveAuthStore(pinia?: Pinia) {
  return pinia ? useAuthStore(pinia) : useAuthStore()
}

export function handleGlobalSessionExpired(
  options: {
    pinia?: Pinia
    requestUrl?: string
  } = {}
): boolean {
  if (!shouldRedirectToErrorStatusPage(401, options.requestUrl)) {
    return false
  }

  const authStore = resolveAuthStore(options.pinia)
  authStore.logout()
  redirectToErrorStatusPage(401, options.requestUrl)
  return true
}

export function installGlobalHttpErrorHandling(pinia: Pinia): void {
  if (httpErrorHandlingInstalled) {
    return
  }

  getAxiosInstance().interceptors.response.use(
    (response) => response,
    (error: unknown) => {
      if (error instanceof ApiError && error.status === 401) {
        handleGlobalSessionExpired({ pinia, requestUrl: error.requestUrl })
      }
      return Promise.reject(error)
    }
  )

  httpErrorHandlingInstalled = true
}

export function createGlobalVueErrorHandler() {
  return (
    err: unknown,
    _instance: ComponentPublicInstance | null,
    info: string
  ): void => {
    console.error('Vue error:', err, info)
    if (err instanceof ApiError) {
      return
    }
    redirectToErrorStatusPage(500)
  }
}

export function createGlobalRouterErrorHandler() {
  return (error: unknown): void => {
    console.error('Router error:', error)
    redirectToErrorStatusPage(500)
  }
}

export function setupGlobalErrorRuntime(app: App, router: Router, pinia: Pinia): void {
  installGlobalHttpErrorHandling(pinia)
  app.config.errorHandler = createGlobalVueErrorHandler()
  router.onError(createGlobalRouterErrorHandler())
}
