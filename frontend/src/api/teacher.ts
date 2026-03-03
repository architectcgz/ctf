import { request } from './request'

export async function getClasses() {
  return request<unknown[]>({ method: 'GET', url: '/teacher/classes' })
}

export async function getClassStudents(name: string) {
  return request<unknown[]>({ method: 'GET', url: `/teacher/classes/${encodeURIComponent(name)}/students` })
}

export async function getStudentProgress(id: string) {
  return request<unknown>({ method: 'GET', url: `/teacher/students/${encodeURIComponent(id)}/progress` })
}

export async function exportClassReport(data: Record<string, unknown>) {
  return request<unknown>({ method: 'POST', url: '/reports/class', data })
}

