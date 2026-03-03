import { request } from './request'

import type { AuthUser } from '@/stores/auth'

export interface LoginRequest {
  username: string
  password: string
}

export interface AuthTokens {
  access_token: string
  refresh_token?: string
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

export async function refreshToken(refreshToken?: string): Promise<AuthTokens> {
  return request<AuthTokens>({ method: 'POST', url: '/auth/refresh', data: refreshToken ? { refresh_token: refreshToken } : {} })
}

export async function logout(): Promise<void> {
  await request<void>({ method: 'POST', url: '/auth/logout' })
}

export async function getProfile(): Promise<AuthUser> {
  return request<AuthUser>({ method: 'GET', url: '/auth/profile' })
}

export async function changePassword(data: { old_password: string; new_password: string }): Promise<void> {
  await request<void>({ method: 'PUT', url: '/auth/password', data })
}

export async function getWsTicket(): Promise<{ ticket: string; expires_at?: string }> {
  return request<{ ticket: string; expires_at?: string }>({ method: 'POST', url: '/auth/ws-ticket' })
}

