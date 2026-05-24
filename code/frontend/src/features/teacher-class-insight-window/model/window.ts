import type { TeacherClassInsightQueryData } from '@/api/contracts'

export interface ClassInsightWindowDraft {
  fromDate: string
  toDate: string
}

type TeacherClassInsightQueryValue =
  | string
  | null
  | undefined
  | Array<string | null | undefined>

interface TeacherClassInsightQueryInput {
  from_date?: TeacherClassInsightQueryValue
  to_date?: TeacherClassInsightQueryValue
}

function normalizeQueryValue(value: TeacherClassInsightQueryValue): string {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0].trim() : ''
  }
  return typeof value === 'string' ? value.trim() : ''
}

function normalizeDateValue(value?: string): string {
  return value?.trim() || ''
}

export function createClassInsightWindowDraft(input?: {
  fromDate?: string
  toDate?: string
}): ClassInsightWindowDraft {
  return {
    fromDate: normalizeDateValue(input?.fromDate),
    toDate: normalizeDateValue(input?.toDate),
  }
}

export function parseClassInsightWindowQuery(
  query: TeacherClassInsightQueryInput
): ClassInsightWindowDraft {
  return createClassInsightWindowDraft({
    fromDate: normalizeQueryValue(query.from_date),
    toDate: normalizeQueryValue(query.to_date),
  })
}

export function hasClassInsightWindow(window: ClassInsightWindowDraft): boolean {
  return window.fromDate.length > 0 || window.toDate.length > 0
}

export function getClassInsightWindowError(
  window: ClassInsightWindowDraft
): string | null {
  const hasFromDate = window.fromDate.length > 0
  const hasToDate = window.toDate.length > 0
  if (hasFromDate !== hasToDate) {
    return '开始日期和结束日期需要同时填写'
  }
  return null
}

export function buildClassInsightWindowQuery(
  window: ClassInsightWindowDraft
): TeacherClassInsightQueryData | undefined {
  if (!hasClassInsightWindow(window)) {
    return undefined
  }

  return {
    from_date: window.fromDate,
    to_date: window.toDate,
  }
}

export function describeClassInsightWindow(window: ClassInsightWindowDraft): string {
  if (!hasClassInsightWindow(window)) {
    return '默认最近 7 天'
  }
  return `${window.fromDate} 至 ${window.toDate}`
}

export function isSameClassInsightWindow(
  left: ClassInsightWindowDraft,
  right: ClassInsightWindowDraft
): boolean {
  return left.fromDate === right.fromDate && left.toDate === right.toDate
}
