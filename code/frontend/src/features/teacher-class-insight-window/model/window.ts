import type { TeacherClassInsightQueryData } from '@/api/contracts'

export interface TeacherClassInsightWindowDraft {
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

export function createTeacherClassInsightWindowDraft(input?: {
  fromDate?: string
  toDate?: string
}): TeacherClassInsightWindowDraft {
  return {
    fromDate: normalizeDateValue(input?.fromDate),
    toDate: normalizeDateValue(input?.toDate),
  }
}

export function parseTeacherClassInsightWindowQuery(
  query: TeacherClassInsightQueryInput
): TeacherClassInsightWindowDraft {
  return createTeacherClassInsightWindowDraft({
    fromDate: normalizeQueryValue(query.from_date),
    toDate: normalizeQueryValue(query.to_date),
  })
}

export function hasTeacherClassInsightWindow(window: TeacherClassInsightWindowDraft): boolean {
  return window.fromDate.length > 0 || window.toDate.length > 0
}

export function getTeacherClassInsightWindowError(
  window: TeacherClassInsightWindowDraft
): string | null {
  const hasFromDate = window.fromDate.length > 0
  const hasToDate = window.toDate.length > 0
  if (hasFromDate !== hasToDate) {
    return '开始日期和结束日期需要同时填写'
  }
  return null
}

export function buildTeacherClassInsightWindowQuery(
  window: TeacherClassInsightWindowDraft
): TeacherClassInsightQueryData | undefined {
  if (!hasTeacherClassInsightWindow(window)) {
    return undefined
  }

  return {
    from_date: window.fromDate,
    to_date: window.toDate,
  }
}

export function describeTeacherClassInsightWindow(window: TeacherClassInsightWindowDraft): string {
  if (!hasTeacherClassInsightWindow(window)) {
    return '默认最近 7 天'
  }
  return `${window.fromDate} 至 ${window.toDate}`
}

export function isSameTeacherClassInsightWindow(
  left: TeacherClassInsightWindowDraft,
  right: TeacherClassInsightWindowDraft
): boolean {
  return left.fromDate === right.fromDate && left.toDate === right.toDate
}
