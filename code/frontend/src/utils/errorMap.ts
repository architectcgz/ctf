export const AUTH_ERROR_CODES = {
  SESSION_EXPIRED: 11002,
} as const

const ERROR_MAP: Record<number, string> = {
  // 通用错误码 (10xxx)
  10001: '参数校验失败，请检查输入',
  10002: '资源不存在',
  10003: '操作被拒绝',

  // 认证错误码 (11xxx)
  11001: '用户名或密码错误',
  11002: '登录状态已过期，请重新登录',
  11003: '登录状态已过期，请重新登录',
  11010: '登录失败次数过多，请稍后再试',

  // 实例错误码 (13xxx)
  13001: '靶场不存在或已下线',
  13002: '实例数量已达上限，请先销毁已有实例',
  13003: 'Flag 错误，请检查后重试',
  13004: '提交过于频繁，请稍后再试',

  // 竞赛错误码 (14xxx)
  14001: '竞赛未开始或已结束',
  14002: '未报名该竞赛',
  14003: '队伍人数已满',
  14008: '邀请码无效',
}

export function mapErrorCode(code: number | undefined): string | undefined {
  if (code === undefined) return undefined
  return ERROR_MAP[code]
}
