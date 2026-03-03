import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { UserRole } from '@/utils/constants'

export interface AuthUser {
  id: string
  username: string
  role: UserRole
  avatar?: string
  class_name?: string
}

const LS_ACCESS_TOKEN_KEY = 'ctf_access_token'
const LS_REFRESH_TOKEN_KEY = 'ctf_refresh_token'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const accessToken = ref<string>('')
  const refreshToken = ref<string>('') // Optional; prefer HttpOnly Cookie in production

  const isLoggedIn = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTeacher = computed(() => user.value?.role === 'teacher')
  const isStudent = computed(() => user.value?.role === 'student')

  function setAuth(nextUser: AuthUser, nextAccessToken: string, nextRefreshToken?: string): void {
    user.value = nextUser
    accessToken.value = nextAccessToken
    localStorage.setItem(LS_ACCESS_TOKEN_KEY, nextAccessToken)

    if (typeof nextRefreshToken === 'string' && nextRefreshToken.length > 0) {
      refreshToken.value = nextRefreshToken
      localStorage.setItem(LS_REFRESH_TOKEN_KEY, nextRefreshToken)
    }
  }

  function updateTokens(nextAccessToken: string, nextRefreshToken?: string): void {
    accessToken.value = nextAccessToken
    localStorage.setItem(LS_ACCESS_TOKEN_KEY, nextAccessToken)

    if (typeof nextRefreshToken === 'string') {
      refreshToken.value = nextRefreshToken
      localStorage.setItem(LS_REFRESH_TOKEN_KEY, nextRefreshToken)
    }
  }

  function logout(): void {
    user.value = null
    accessToken.value = ''
    refreshToken.value = ''
    localStorage.removeItem(LS_ACCESS_TOKEN_KEY)
    localStorage.removeItem(LS_REFRESH_TOKEN_KEY)
  }

  // Only restore tokens. User profile is fetched via /auth/profile on-demand (router guards).
  function restore(): void {
    accessToken.value = localStorage.getItem(LS_ACCESS_TOKEN_KEY) || ''
    refreshToken.value = localStorage.getItem(LS_REFRESH_TOKEN_KEY) || ''
  }

  return {
    user,
    accessToken,
    refreshToken,
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

