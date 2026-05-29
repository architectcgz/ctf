import {
  login as loginApi,
  logout as logoutApi,
  register as registerApi,
  type LoginRequest,
  type RegisterRequest,
} from '@/api/auth'
import { useAuthStore, type AuthUser } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

export function useAuth() {
  const authStore = useAuthStore()
  const toast = useToast()

  async function login(payload: LoginRequest): Promise<AuthUser> {
    const resp = await loginApi(payload)
    authStore.setAuth(resp.user)
    toast.success('登录成功')
    return resp.user
  }

  async function register(payload: RegisterRequest): Promise<AuthUser> {
    const resp = await registerApi(payload)
    authStore.setAuth(resp.user)
    toast.success('注册成功')
    return resp.user
  }

  async function logout(): Promise<void> {
    try {
      await logoutApi()
    } catch {
      // Ignore network failures on logout.
    } finally {
      authStore.logout()
      toast.info('已退出登录')
    }
  }

  return { login, register, logout }
}
