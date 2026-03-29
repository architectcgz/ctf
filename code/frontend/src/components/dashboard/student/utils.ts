import type { TimelineEvent } from '@/api/contracts'

export function progressRate(total: number, solved: number): number {
  if (!total) return 0
  return Math.round((solved / total) * 100)
}

export function timelineSummary(event: TimelineEvent): string {
  if (event.detail) {
    return event.detail
  }
  if (event.type === 'challenge_detail_view' || (event.meta?.raw_type as string | undefined) === 'challenge_detail_view') {
    return '查看题目详情，进入读题与线索分析阶段'
  }
  if (event.type === 'instance_access' || (event.meta?.raw_type as string | undefined) === 'instance_access') {
    return '访问攻击目标，开始与靶机进行实际交互'
  }
  if (
    event.type === 'instance_proxy_request' ||
    (event.meta?.raw_type as string | undefined) === 'instance_proxy_request'
  ) {
    return '经平台代理向靶机发起请求，系统已记录本次利用轨迹'
  }
  if (event.type === 'solve') {
    return `成功解出题目${event.points ? `，获得 ${event.points} 分` : ''}`
  }
  if (event.type === 'submit') {
    return '提交过 Flag，当前记录未判定为成功'
  }
  if (event.type === 'hint' || (event.meta?.raw_type as string | undefined) === 'hint_unlock') {
    return '解锁了一条提示，说明训练进入了更具体的利用定位阶段'
  }
  if (event.type === 'instance_extend' || (event.meta?.raw_type as string | undefined) === 'instance_extend') {
    return '延长实例有效期，继续当前利用过程'
  }
  if ((event.meta?.raw_type as string | undefined) === 'instance_destroy') {
    return '结束了一个练习实例'
  }
  return '启动或操作了练习实例'
}

export function timelineTypeLabel(event: TimelineEvent): string {
  if (event.type === 'challenge_detail_view') return '读题侦察'
  if (event.type === 'instance_access') return '访问目标'
  if (event.type === 'instance_proxy_request') return '利用轨迹'
  if (event.type === 'solve') return '解题成功'
  if (event.type === 'submit') return '提交记录'
  if (event.type === 'hint') return '提示解锁'
  if (event.type === 'instance_extend') return '实例续期'
  if ((event.meta?.raw_type as string | undefined) === 'instance_destroy') return '实例销毁'
  return '实例操作'
}

export function timelineTypeTone(event: TimelineEvent): string {
  if (event.type === 'challenge_detail_view') return 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
  if (event.type === 'instance_access') return 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
  if (event.type === 'instance_proxy_request') return 'border-[var(--color-danger)]/30 bg-[var(--color-danger)]/10 text-[var(--color-danger)]'
  if (event.type === 'solve') return 'border-[var(--color-success)]/30 bg-[var(--color-success)]/10 text-[var(--color-success)]'
  if (event.type === 'submit') return 'border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 text-[var(--color-warning)]'
  if (event.type === 'hint') return 'border-[var(--color-cat-reverse)]/30 bg-[var(--color-cat-reverse)]/10 text-[var(--color-cat-reverse)]'
  return 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
}
