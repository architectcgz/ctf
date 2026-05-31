import { useRouter } from 'vue-router'

import { useAuth } from '@/features/auth'

export function useLayoutSessionActionsBridge() {
  const router = useRouter()
  const { logout } = useAuth()

  return {
    async logout() {
      await logout()
      await router.push('/login')
    },
  }
}
