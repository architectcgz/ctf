import type { AdminImageListItem, ImageStatus } from '@/api/contracts'

export type ImageSortKey = 'created_at' | 'name' | 'tag'

export interface ImageSortConfig {
  key: ImageSortKey
  order: 'asc' | 'desc'
}

export type ImageStatusSummaryItem = {
  key: string
  label: string
  value: number
  tone: 'success' | 'warning' | 'danger' | 'muted'
}

const imageStatusMeta: Record<
  ImageStatus,
  { label: string; color: string; backgroundColor: string }
> = {
  pending: {
    label: '等待中',
    color: 'color-mix(in srgb, var(--journal-muted) 84%, var(--journal-ink))',
    backgroundColor: 'color-mix(in srgb, var(--journal-muted) 14%, transparent)',
  },
  building: {
    label: '构建中',
    color: 'var(--color-warning)',
    backgroundColor: 'color-mix(in srgb, var(--color-warning) 14%, transparent)',
  },
  pushed: {
    label: '已推送',
    color: 'var(--color-primary)',
    backgroundColor: 'color-mix(in srgb, var(--color-primary) 14%, transparent)',
  },
  verifying: {
    label: '校验中',
    color: 'var(--color-warning)',
    backgroundColor: 'color-mix(in srgb, var(--color-warning) 14%, transparent)',
  },
  available: {
    label: '可用',
    color: 'var(--color-success)',
    backgroundColor: 'color-mix(in srgb, var(--color-success) 14%, transparent)',
  },
  failed: {
    label: '失败',
    color: 'var(--color-danger)',
    backgroundColor: 'color-mix(in srgb, var(--color-danger) 14%, transparent)',
  },
}

export function filterAndSortImages(
  rows: AdminImageListItem[],
  keyword: string,
  statusFilter: ImageStatus | '',
  sortConfig: ImageSortConfig
): AdminImageListItem[] {
  const normalizedKeyword = keyword.trim().toLowerCase()
  const filteredRows = rows.filter((row) => {
    const matchesKeyword =
      !normalizedKeyword ||
      row.name.toLowerCase().includes(normalizedKeyword) ||
      row.tag.toLowerCase().includes(normalizedKeyword) ||
      (row.description || '').toLowerCase().includes(normalizedKeyword)
    const matchesStatus = !statusFilter || row.status === statusFilter

    return matchesKeyword && matchesStatus
  })

  const sortedRows = [...filteredRows]
  sortedRows.sort((left, right) => {
    switch (sortConfig.key) {
      case 'name': {
        const delta = left.name.localeCompare(right.name, 'zh-CN')
        return sortConfig.order === 'asc' ? delta : -delta
      }
      case 'tag': {
        const delta = left.tag.localeCompare(right.tag, 'zh-CN', { numeric: true })
        return sortConfig.order === 'asc' ? delta : -delta
      }
      case 'created_at':
      default: {
        const delta = new Date(left.created_at).getTime() - new Date(right.created_at).getTime()
        return sortConfig.order === 'asc' ? delta : -delta
      }
    }
  })

  return sortedRows
}

export function buildImageStatusSummary(rows: AdminImageListItem[]): ImageStatusSummaryItem[] {
  const counts = {
    available: 0,
    pending: 0,
    building: 0,
    pushed: 0,
    verifying: 0,
    failed: 0,
  }

  for (const row of rows) {
    counts[row.status] += 1
  }

  const summary: ImageStatusSummaryItem[] = []

  if (counts.building > 0) {
    summary.push({ key: 'building', label: '构建中', value: counts.building, tone: 'warning' })
  }
  if (counts.verifying > 0) {
    summary.push({ key: 'verifying', label: '校验中', value: counts.verifying, tone: 'warning' })
  }
  if (counts.pending > 0) {
    summary.push({ key: 'pending', label: '等待中', value: counts.pending, tone: 'muted' })
  }
  if (counts.failed > 0) {
    summary.push({ key: 'failed', label: '失败', value: counts.failed, tone: 'danger' })
  }

  return summary
}

export function getImageStatusLabel(status: ImageStatus): string {
  return imageStatusMeta[status].label
}

export function getImageStatusStyle(status: ImageStatus): Record<string, string> {
  const meta = imageStatusMeta[status]
  return {
    backgroundColor: meta.backgroundColor,
    color: meta.color,
  }
}

export function formatImageSize(bytes?: number): string {
  if (!bytes) return '未知大小'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(2)} ${units[unitIndex]}`
}

export function formatImageDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
