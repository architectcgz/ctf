import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { refreshToken } from '@/api/auth'
import type { UserRole } from '@/utils/constants'

export interface AuthUser {
  id: string
  username: string
  role: UserRole
  avatar?: string
  name?: string
  class_name?: string
}

const LEGACY_LS_ACCESS_TOKEN_KEY = 'ctf_access_token'
const LEGACY_LS_REFRESH_TOKEN_KEY = 'ctf_refresh_token'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const accessToken = ref<string>('')
  const sessionRestored = ref(false)
  let restorePromise: Promise<void> | null = null

  const isLoggedIn = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTeacher = computed(() => user.value?.role === 'teacher')
  const isStudent = computed(() => user.value?.role === 'student')

  function cleanupLegacyStorage(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(LEGACY_LS_ACCESS_TOKEN_KEY)
    localStorage.removeItem(LEGACY_LS_REFRESH_TOKEN_KEY)
  }

  function setAuth(nextUser: AuthUser, nextAccessToken: string): void {
    user.value = nextUser
    accessToken.value = nextAccessToken
    sessionRestored.value = true
    cleanupLegacyStorage()
  }

  function updateTokens(nextAccessToken: string): void {
    accessToken.value = nextAccessToken
    sessionRestored.value = true
    cleanupLegacyStorage()
  }

  function logout(): void {
    user.value = null
    accessToken.value = ''
    sessionRestored.value = true
    cleanupLegacyStorage()
  }

  // Access Token stays in memory only. Refresh after reload relies on the HttpOnly refresh cookie.
  async function restore(): Promise<void> {
    cleanupLegacyStorage()

    if (accessToken.value || sessionRestored.value) {
      return
    }
    if (restorePromise) {
      return restorePromise
    }

    restorePromise = (async () => {
      try {
        const tokens = await refreshToken({ suppressErrorToast: true })
        accessToken.value = tokens.access_token
      } catch {
        accessToken.value = ''
      } finally {
        sessionRestored.value = true
        restorePromise = null
      }
    })()

    return restorePromise
  }

  return {
    user,
    accessToken,
    sessionRestored,
    isLoggedIn,
    isAdmin,
    isTeacher,
    isStudent,
    setAuth,
    updateTokens,
    logout,
    restore,
  }
})
