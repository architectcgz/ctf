export const AUTH_ERROR_CODES = {
  ACCESS_TOKEN_EXPIRED: 11002,
} as const

const ERROR_MAP: Record<number, string> = {
  11001: '用户名或密码错误',
  11002: '登录状态已过期，请重新登录',
  11010: '登录失败次数过多，请稍后再试',
  13002: '实例数量已达上限，请先销毁已有实例',
  13003: 'Flag 错误，请检查后重试',
  13004: '提交过于频繁，请稍后再试',
  14003: '队伍人数已满',
  14008: '邀请码无效',
}

export function mapErrorCode(code: number | undefined): string | undefined {
  if (code === undefined) return undefined
  return ERROR_MAP[code]
}
