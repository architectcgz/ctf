export type UserRole = 'student' | 'teacher' | 'admin'

export const USER_ROLES: readonly UserRole[] = ['student', 'teacher', 'admin'] as const

export const APP_TITLE_PREFIX = 'CTF 靶场平台'

// WebSocket 配置
export const WS_MAX_RECONNECT_ATTEMPTS = 20
export const WS_HEARTBEAT_INTERVAL_MS = 30_000
export const WS_PONG_TIMEOUT_MS = 60_000
export const WS_MAX_RECONNECT_DELAY_MS = 30_000
export const WS_RECONNECT_BASE_DELAY_MS = 1000

// Toast 持续时间配置
export const TOAST_DURATION = {
  SUCCESS: 3000,
  INFO: 3000,
  WARNING: 4000,
  ERROR: 5000,
} as const

// 分页配置
export const DEFAULT_PAGE_SIZE = 20
