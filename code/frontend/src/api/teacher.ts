import { request } from './request'

import type {
  MyProgressData,
  RecommendationItem,
  ReportExportData,
  SkillProfileData,
  TeacherClassItem,
  TeacherStudentItem,
} from './contracts'
import { normalizeSkillProfile, type RawSkillProfileResponse } from '@/utils/skillProfile'

export async function getClasses(): Promise<TeacherClassItem[]> {
  return request<TeacherClassItem[]>({ method: 'GET', url: '/teacher/classes' })
}

export async function getClassStudents(name: string, params?: { student_no?: string }) {
  const payload = await request<
    Array<{
      id: string | number
      username: string
      student_no?: string
      name?: string
    }>
  >({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/students`,
    params: {
      student_no: params?.student_no,
    },
  })

  return payload.map((item) => ({
    ...item,
    id: String(item.id),
  }))
}

export async function getStudentProgress(id: string) {
  return request<MyProgressData>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/progress`,
  })
}

export async function getStudentSkillProfile(id: string): Promise<SkillProfileData> {
  const payload = await request<RawSkillProfileResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/skill-profile`,
  })
  return normalizeSkillProfile(payload)
}

export async function getStudentRecommendations(id: string): Promise<RecommendationItem[]> {
  const payload = await request<
    Array<{
      challenge_id: string | number
      title: string
      category: RecommendationItem['category']
      difficulty: RecommendationItem['difficulty']
      reason: string
    }>
  >({ method: 'GET', url: `/teacher/students/${encodeURIComponent(id)}/recommendations` })

  return payload.map((item) => ({
    ...item,
    challenge_id: String(item.challenge_id),
  }))
}

export async function exportClassReport(data: Record<string, unknown>) {
  return request<ReportExportData>({ method: 'POST', url: '/reports/class', data })
}
