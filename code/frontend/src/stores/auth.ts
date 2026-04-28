import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { getProfile } from '@/api/auth'
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
  const sessionRestored = ref(false)
  let restorePromise: Promise<void> | null = null

  const isLoggedIn = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTeacher = computed(() => user.value?.role === 'teacher')
  const isStudent = computed(() => user.value?.role === 'student')

  function cleanupLegacyStorage(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(LEGACY_LS_ACCESS_TOKEN_KEY)
    localStorage.removeItem(LEGACY_LS_REFRESH_TOKEN_KEY)
  }

  function setAuth(nextUser: AuthUser): void {
    user.value = nextUser
    sessionRestored.value = true
    cleanupLegacyStorage()
  }

  function logout(): void {
    user.value = null
    sessionRestored.value = true
    cleanupLegacyStorage()
  }

  // Session restoration relies on the HttpOnly session cookie; the frontend only keeps the user snapshot in memory.
  async function restore(): Promise<void> {
    cleanupLegacyStorage()

    if (user.value || sessionRestored.value) {
      return
    }
    if (restorePromise) {
      return restorePromise
    }

    restorePromise = (async () => {
      try {
        user.value = await getProfile({ suppressErrorToast: true })
      } catch {
        user.value = null
      } finally {
        sessionRestored.value = true
        restorePromise = null
      }
    })()

    return restorePromise
  }

  return {
    user,
    sessionRestored,
    isLoggedIn,
    isAdmin,
    isTeacher,
    isStudent,
    setAuth,
    logout,
    restore,
  }
})
