import { useAuth } from '@/features/auth'

/**
 * 桥接 app shell 与 auth feature 的登出流程。
 * 通过回调注入路由导航，避免 feature 层直接依赖 vue-router。
 */
export function useLayoutSessionActionsBridge(navigateToLogin: () => void) {
  const { logout } = useAuth()

  return {
    async logout() {
      await logout()
      navigateToLogin()
    },
  }
}
