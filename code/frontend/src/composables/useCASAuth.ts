import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import {
  completeCASLogin as completeCASLoginApi,
  getCASLogin,
  getCASStatus,
  type CASStatusResponse,
} from '@/api/auth'
import { sanitizeRedirectPath } from '@/router/guards'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { redirectTo } from '@/utils/browser'

const CAS_REDIRECT_STORAGE_KEY = 'ctf_cas_post_login_redirect'

export function useCASAuth() {
  const router = useRouter()
  const authStore = useAuthStore()
  const toast = useToast()

  const casStatus = ref<CASStatusResponse | null>(null)
  const casLoading = ref(false)
  const casRedirecting = ref(false)

  const casReady = computed(
    () => Boolean(casStatus.value?.enabled) && Boolean(casStatus.value?.configured)
  )

  async function fetchCASStatus(): Promise<CASStatusResponse | null> {
    casLoading.value = true
    try {
      casStatus.value = await getCASStatus()
      return casStatus.value
    } catch {
      casStatus.value = null
      return null
    } finally {
      casLoading.value = false
    }
  }

  function persistRedirectTarget(redirectTo: string): void {
    if (typeof window === 'undefined') {
      return
    }
    window.sessionStorage.setItem(
      CAS_REDIRECT_STORAGE_KEY,
      sanitizeRedirectPath(redirectTo || '/dashboard')
    )
  }

  function consumeRedirectTarget(fallback: string = '/dashboard'): string {
    if (typeof window === 'undefined') {
      return fallback
    }
    const raw = window.sessionStorage.getItem(CAS_REDIRECT_STORAGE_KEY)
    window.sessionStorage.removeItem(CAS_REDIRECT_STORAGE_KEY)
    const sanitized = sanitizeRedirectPath(raw)
    return sanitized === '/' ? fallback : sanitized
  }

  async function beginCASLogin(targetPath: string = '/dashboard'): Promise<void> {
    casRedirecting.value = true
    try {
      persistRedirectTarget(targetPath)
      const loginResp = await getCASLogin()
      redirectTo(loginResp.redirect_url)
    } finally {
      casRedirecting.value = false
    }
  }

  async function finishCASLogin(ticket: string, redirectTo?: string): Promise<void> {
    const loginResp = await completeCASLoginApi(ticket, { suppressErrorToast: true })
    authStore.setAuth(loginResp.user, loginResp.access_token)
    toast.success('CAS 登录成功')
    const nextPath = redirectTo
      ? sanitizeRedirectPath(redirectTo)
      : consumeRedirectTarget('/dashboard')
    await router.replace(nextPath === '/' ? '/dashboard' : nextPath)
  }

  return {
    casStatus,
    casLoading,
    casReady,
    casRedirecting,
    fetchCASStatus,
    beginCASLogin,
    finishCASLogin,
  }
}
