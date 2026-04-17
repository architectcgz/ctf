export function resolveClassManagementRouteName(
  role?: string | null
): 'AdminClassManagement' | 'ClassManagement' {
  return role === 'admin' ? 'AdminClassManagement' : 'ClassManagement'
}
