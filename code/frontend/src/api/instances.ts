import type {
  InstanceDirectoryItem,
  InstanceDirectoryPageData,
} from '@/api/contracts'
import {
  destroyPlatformInstance,
  getPlatformInstances,
} from '@/api/admin'
import {
  destroyTeacherInstance,
  getTeacherInstances,
  type InstanceDirectoryStatusFilter,
} from '@/api/teacher'
import type { UserRole } from '@/utils/constants'

type InstanceAccessRole = UserRole | null | undefined

interface InstanceDirectoryQueryParams {
  class_name?: string
  keyword?: string
  student_no?: string
  status?: InstanceDirectoryStatusFilter
  page?: number
  page_size?: number
}

export async function getInstanceDirectoryByRole(
  role: InstanceAccessRole,
  params?: InstanceDirectoryQueryParams,
  options?: { signal?: AbortSignal }
): Promise<InstanceDirectoryPageData<InstanceDirectoryItem>> {
  return role === 'admin'
    ? getPlatformInstances(params, options)
    : getTeacherInstances(params, options)
}

export async function destroyManagedInstanceByRole(
  role: InstanceAccessRole,
  id: string
): Promise<void> {
  return role === 'admin'
    ? destroyPlatformInstance(id)
    : destroyTeacherInstance(id)
}
