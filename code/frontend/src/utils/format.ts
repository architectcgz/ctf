export function formatDurationHms(totalSeconds: number): string {
  const safeSeconds = Math.max(0, Math.floor(totalSeconds))
  const hours = Math.floor(safeSeconds / 3600)
  const minutes = Math.floor((safeSeconds % 3600) / 60)
  const seconds = safeSeconds % 60
  return [hours, minutes, seconds].map((v) => String(v).padStart(2, '0')).join(':')
}

export function formatTime(time: string): string {
  return new Date(time).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}天 ${hours % 24}时`
  if (hours > 0) return `${hours}时 ${minutes % 60}分`
  if (minutes > 0) return `${minutes}分 ${seconds % 60}秒`
  return `${seconds}秒`
}

