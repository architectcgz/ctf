import { useRouter } from 'vue-router'

import { login as loginApi, logout as logoutApi, register as registerApi, type LoginRequest, type RegisterRequest } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

export function useAuth() {
  const router = useRouter()
  const authStore = useAuthStore()
  const toast = useToast()

  async function login(payload: LoginRequest, redirectTo: string = '/dashboard'): Promise<void> {
    const resp = await loginApi(payload)
    authStore.setAuth(resp.user, resp.access_token)
    toast.success('登录成功')
    await router.push(redirectTo)
  }

  async function register(payload: RegisterRequest, redirectTo: string = '/dashboard'): Promise<void> {
    const resp = await registerApi(payload)
    authStore.setAuth(resp.user, resp.access_token)
    toast.success('注册成功')
    await router.push(redirectTo)
  }

  async function logout(): Promise<void> {
    try {
      await logoutApi()
    } catch {
      // Ignore network failures on logout.
    } finally {
      authStore.logout()
      toast.info('已退出登录')
      await router.push('/login')
    }
  }

  return { login, register, logout }
}
