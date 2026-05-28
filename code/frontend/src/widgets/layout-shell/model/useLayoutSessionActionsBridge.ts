import { useAuth } from '@/features/auth'

export function useLayoutSessionActionsBridge() {
  const { logout } = useAuth()

  return {
    logout,
  }
}
