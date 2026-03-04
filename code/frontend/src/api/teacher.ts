import { request } from './request'

import type { MyProgressData, RecommendationItem, ReportExportData, SkillProfileData, TeacherClassItem, TeacherStudentItem } from './contracts'

export async function getClasses(): Promise<TeacherClassItem[]> {
  return request<TeacherClassItem[]>({ method: 'GET', url: '/teacher/classes' })
}

export async function getClassStudents(name: string) {
  return request<TeacherStudentItem[]>({ method: 'GET', url: `/teacher/classes/${encodeURIComponent(name)}/students` })
}

export async function getStudentProgress(id: string) {
  return request<MyProgressData>({ method: 'GET', url: `/teacher/students/${encodeURIComponent(id)}/progress` })
}

export async function getStudentSkillProfile(id: string): Promise<SkillProfileData> {
  return request<SkillProfileData>({ method: 'GET', url: `/teacher/students/${encodeURIComponent(id)}/skill-profile` })
}

export async function getStudentRecommendations(id: string): Promise<RecommendationItem[]> {
  return request<RecommendationItem[]>({ method: 'GET', url: `/teacher/students/${encodeURIComponent(id)}/recommendations` })
}

export async function exportClassReport(data: Record<string, unknown>) {
  return request<ReportExportData>({ method: 'POST', url: '/reports/class', data })
}
