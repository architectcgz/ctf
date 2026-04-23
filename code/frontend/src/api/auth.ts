import { request, type RequestConfig } from './request'

import type { AuthUser } from '@/stores/auth'
import type { WsTicketData } from '@/api/contracts'

export interface LoginRequest {
  username: string
  password: string
}

export interface AuthTokens {
  access_token: string
  token_type: 'Bearer'
  expires_in: number
}

export interface LoginResponse extends AuthTokens {
  user: AuthUser
}

export async function login(data: LoginRequest): Promise<LoginResponse> {
  return request<LoginResponse>({ method: 'POST', url: '/auth/login', data })
}

export interface RegisterRequest {
  username: string
  password: string
  class_name?: string
  name?: string
}

export async function register(data: RegisterRequest): Promise<LoginResponse> {
  return request<LoginResponse>({ method: 'POST', url: '/auth/register', data })
}

export async function refreshToken(config?: Pick<RequestConfig, 'suppressErrorToast'>): Promise<AuthTokens> {
  return request<AuthTokens>({
    method: 'POST',
    url: '/auth/refresh',
    data: {},
    suppressErrorToast: config?.suppressErrorToast,
  })
}

export async function logout(): Promise<void> {
  await request<void>({ method: 'POST', url: '/auth/logout' })
}

export async function getProfile(): Promise<AuthUser> {
  const payload = await request<Omit<AuthUser, 'id'> & { id: string | number }>({
    method: 'GET',
    url: '/auth/profile',
  })
  return {
    ...payload,
    id: String(payload.id),
  }
}

export async function changePassword(data: {
  old_password: string
  new_password: string
}): Promise<void> {
  await request<void>({ method: 'PUT', url: '/auth/password', data })
}

export async function getWsTicket(): Promise<WsTicketData> {
  return request<WsTicketData>({ method: 'POST', url: '/auth/ws-ticket' })
}
