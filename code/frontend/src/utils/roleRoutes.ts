import type { UserRole } from '@/utils/constants'

export const STUDENT_DASHBOARD_PATH = '/student/dashboard'
export const TEACHER_DASHBOARD_PATH = '/academy/overview'
export const ADMIN_DASHBOARD_PATH = '/admin/dashboard'

export function getRoleDashboardPath(role: UserRole | null | undefined): string {
  if (role === 'admin') return ADMIN_DASHBOARD_PATH
  if (role === 'teacher') return TEACHER_DASHBOARD_PATH
  return STUDENT_DASHBOARD_PATH
}
