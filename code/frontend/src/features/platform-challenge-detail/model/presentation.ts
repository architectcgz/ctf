import type { AdminChallengeListItem, FlagType } from '@/api/contracts'

export function summarizeChallengeFlagConfig(config?: AdminChallengeListItem['flag_config']): string {
  if (!config?.configured) return '未配置'

  switch (config.flag_type) {
    case 'static':
      return '静态 Flag'
    case 'dynamic':
      return `动态 Flag / 前缀 ${config.flag_prefix || 'flag'}`
    case 'regex':
      return `正则匹配 / ${config.flag_regex || '未填写'}`
    case 'manual_review':
      return '人工审核'
    default:
      return '未配置'
  }
}

export function buildChallengeFlagDraftSummary(draft: {
  flagPrefix: string
  flagRegex: string
  flagType: FlagType
}): string {
  return summarizeChallengeFlagConfig({
    configured: true,
    flag_type: draft.flagType,
    flag_regex: draft.flagRegex.trim() || undefined,
    flag_prefix: draft.flagPrefix.trim() || undefined,
  })
}
