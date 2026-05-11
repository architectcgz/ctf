import type {
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
  InstanceSharing,
} from '@/api/contracts'

export function getChallengeCategoryLabel(category: ChallengeCategory): string {
  const labels: Record<ChallengeCategory, string> = {
    web: 'Web',
    pwn: 'Pwn',
    reverse: '逆向',
    crypto: '密码',
    misc: '杂项',
    forensics: '取证',
  }
  return labels[category]
}

export function getChallengeCategoryColor(
  category: ChallengeCategory,
  colorMap: Partial<Record<ChallengeCategory, string>> = {}
): string {
  const colors: Record<ChallengeCategory, string> = {
    web: 'var(--challenge-category-pill-web)',
    pwn: 'var(--challenge-category-pill-pwn)',
    reverse: 'var(--challenge-category-pill-reverse)',
    crypto: 'var(--challenge-category-pill-crypto)',
    misc: 'var(--challenge-category-pill-misc)',
    forensics: 'var(--challenge-category-pill-forensics)',
  }
  return colorMap[category] ?? colors[category]
}

export function getChallengeDifficultyLabel(difficulty: ChallengeDifficulty): string {
  const labels: Record<ChallengeDifficulty, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    insane: '地狱',
  }
  return labels[difficulty]
}

export function getChallengeDifficultyColor(
  difficulty: ChallengeDifficulty,
  colorMap: Partial<Record<ChallengeDifficulty, string>> = {}
): string {
  const defaultColors: Record<ChallengeDifficulty, string> = {
    beginner: 'var(--challenge-difficulty-pill-beginner)',
    easy: 'var(--challenge-difficulty-pill-easy)',
    medium: 'var(--challenge-difficulty-pill-medium)',
    hard: 'var(--challenge-difficulty-pill-hard)',
    insane: 'var(--challenge-difficulty-pill-insane)',
  }

  return colorMap[difficulty] ?? defaultColors[difficulty]
}

export function getChallengeStatusLabel(status?: ChallengeStatus): string {
  switch (status) {
    case 'published':
      return '已发布'
    case 'draft':
      return '草稿'
    case 'archived':
      return '已归档'
    default:
      return status || '未设置'
  }
}

export function getChallengeInstanceSharingLabel(mode?: InstanceSharing): string {
  switch (mode) {
    case 'shared':
      return '共享实例'
    case 'per_team':
      return '队伍隔离'
    case 'per_user':
      return '用户隔离'
    default:
      return '未设置'
  }
}

export function formatChallengeDateTime(value?: string): string {
  if (!value) return '未记录'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN')
}
