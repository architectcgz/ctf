import { request } from '../request'

import type {
  AdminUserImportData,
  AdminUserListItem,
  AdminUserUpsertData,
  PageResult,
} from '../contracts'
import type { UserRole } from '@/utils/constants'

type UserStatus = 'active' | 'inactive' | 'locked' | 'banned'

interface UserListParams {
  page?: number
  page_size?: number
  keyword?: string
  student_no?: string
  teacher_no?: string
  role?: UserRole
  status?: UserStatus
  class_name?: string
}

interface RawAdminUser {
  id: string | number
  username: string
  name?: string | null
  email?: string | null
  student_no?: string | null
  teacher_no?: string | null
  class_name?: string | null
  status: UserStatus
  roles: UserRole[]
  created_at: string
}

export interface AdminUserCreatePayload {
  username: string
  name?: string
  password: string
  email?: string
  student_no?: string
  teacher_no?: string
  class_name?: string
  role: UserRole
  status?: UserStatus
}

export interface AdminUserUpdatePayload {
  name?: string
  password?: string
  email?: string
  student_no?: string
  teacher_no?: string
  class_name?: string
  role?: UserRole
  status?: UserStatus
}

function normalizeAdminUser(item: RawAdminUser): AdminUserListItem {
  return {
    id: String(item.id),
    username: item.username,
    name: item.name || undefined,
    email: item.email || undefined,
    student_no: item.student_no || undefined,
    teacher_no: item.teacher_no || undefined,
    class_name: item.class_name || undefined,
    status: item.status,
    roles: item.roles,
    created_at: item.created_at,
  }
}

export async function getUsers(
  params?: UserListParams,
  options?: { signal?: AbortSignal }
): Promise<PageResult<AdminUserListItem>> {
  const response = await request<PageResult<RawAdminUser>>({
    method: 'GET',
    url: '/admin/users',
    params: {
      page: params?.page,
      page_size: params?.page_size,
      keyword: params?.keyword,
      student_no: params?.student_no,
      teacher_no: params?.teacher_no,
      role: params?.role,
      status: params?.status,
      class_name: params?.class_name,
    },
    signal: options?.signal,
  })

  return {
    ...response,
    list: response.list.map(normalizeAdminUser),
  }
}

export async function createUser(data: AdminUserCreatePayload): Promise<AdminUserUpsertData> {
  const response = await request<{ user: RawAdminUser }>({
    method: 'POST',
    url: '/admin/users',
    data,
  })
  return {
    user: normalizeAdminUser(response.user),
  }
}

export async function updateUser(
  id: string,
  data: AdminUserUpdatePayload
): Promise<AdminUserUpsertData> {
  const response = await request<{ user: RawAdminUser }>({
    method: 'PUT',
    url: `/admin/users/${encodeURIComponent(id)}`,
    data,
  })
  return {
    user: normalizeAdminUser(response.user),
  }
}

export async function deleteUser(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/users/${encodeURIComponent(id)}` })
}

export async function importUsers(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request<AdminUserImportData>({
    method: 'POST',
    url: '/admin/users/import',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
