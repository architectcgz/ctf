export type UserRole = 'student' | 'teacher' | 'admin'

export const USER_ROLES: readonly UserRole[] = ['student', 'teacher', 'admin'] as const

export const APP_TITLE_PREFIX = 'CTF 靶场平台'

