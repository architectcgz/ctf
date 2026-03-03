import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { UserRole } from '@/utils/constants'

export interface AuthUser {
  id: string
  username: string
  role: UserRole
  avatar?: string
  name?: string
  class_name?: string
}

const LS_ACCESS_TOKEN_KEY = 'ctf_access_token'
const LEGACY_LS_REFRESH_TOKEN_KEY = 'ctf_refresh_token'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const accessToken = ref<string>('')

  const isLoggedIn = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTeacher = computed(() => user.value?.role === 'teacher')
  const isStudent = computed(() => user.value?.role === 'student')

  function setAuth(nextUser: AuthUser, nextAccessToken: string): void {
    user.value = nextUser
    accessToken.value = nextAccessToken
    localStorage.setItem(LS_ACCESS_TOKEN_KEY, nextAccessToken)
  }

  function updateTokens(nextAccessToken: string): void {
    accessToken.value = nextAccessToken
    localStorage.setItem(LS_ACCESS_TOKEN_KEY, nextAccessToken)
  }

  function logout(): void {
    user.value = null
    accessToken.value = ''
    localStorage.removeItem(LS_ACCESS_TOKEN_KEY)
    // Cleanup legacy key from older dual-mode designs (Refresh Token must not be stored client-side).
    localStorage.removeItem(LEGACY_LS_REFRESH_TOKEN_KEY)
  }

  // Only restore tokens. User profile is fetched via /auth/profile on-demand (router guards).
  function restore(): void {
    accessToken.value = localStorage.getItem(LS_ACCESS_TOKEN_KEY) || ''
    localStorage.removeItem(LEGACY_LS_REFRESH_TOKEN_KEY)
  }

  return {
    user,
    accessToken,
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
