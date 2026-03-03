import { request } from './request'

import type { MyProgressData, ReportExportData, TeacherClassItem, TeacherStudentItem } from './contracts'

export async function getClasses(): Promise<TeacherClassItem[]> {
  return request<TeacherClassItem[]>({ method: 'GET', url: '/teacher/classes' })
}

export async function getClassStudents(name: string) {
  return request<TeacherStudentItem[]>({ method: 'GET', url: `/teacher/classes/${encodeURIComponent(name)}/students` })
}

export async function getStudentProgress(id: string) {
  return request<MyProgressData>({ method: 'GET', url: `/teacher/students/${encodeURIComponent(id)}/progress` })
}

export async function exportClassReport(data: Record<string, unknown>) {
  return request<ReportExportData>({ method: 'POST', url: '/reports/class', data })
}
