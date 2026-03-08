import type { TimelineEvent } from '@/api/contracts'

export function progressRate(total: number, solved: number): number {
  if (!total) return 0
  return Math.round((solved / total) * 100)
}

export function timelineSummary(event: TimelineEvent): string {
  if (event.type === 'solve') {
    return `成功解出题目${event.points ? `，获得 ${event.points} 分` : ''}`
  }
  if (event.type === 'submit') {
    return '提交过 Flag，当前记录未判定为成功'
  }
  if ((event.meta?.raw_type as string | undefined) === 'instance_destroy') {
    return '结束了一个练习实例'
  }
  return '启动或操作了练习实例'
}

export function timelineTypeLabel(event: TimelineEvent): string {
  if (event.type === 'solve') return '解题成功'
  if (event.type === 'submit') return '提交记录'
  if (event.type === 'hint') return '提示操作'
  if ((event.meta?.raw_type as string | undefined) === 'instance_destroy') return '实例销毁'
  return '实例操作'
}

export function timelineTypeTone(event: TimelineEvent): string {
  if (event.type === 'solve') return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200'
  if (event.type === 'submit') return 'border-amber-500/30 bg-amber-500/10 text-amber-200'
  if (event.type === 'hint') return 'border-fuchsia-500/30 bg-fuchsia-500/10 text-fuchsia-200'
  return 'border-sky-500/30 bg-sky-500/10 text-sky-200'
}
