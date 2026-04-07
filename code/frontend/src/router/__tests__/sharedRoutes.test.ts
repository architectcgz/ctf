import { describe, expect, it } from 'vitest'

import { routes } from '@/router'

function getRootChildren() {
  const root = routes.find((route) => route.path === '/')
  return root?.children ?? []
}

function findChild(path: string) {
  return getRootChildren().find((route) => route.path === path)
}

describe('shared route canonical paths', () => {
  it('uses academy paths as the canonical location for shared teaching pages', () => {
    expect(findChild('academy/overview')?.name).toBe('TeacherDashboard')
    expect(findChild('academy/classes')?.name).toBe('ClassManagement')
    expect(findChild('academy/students')?.name).toBe('TeacherStudentManagement')
    expect(findChild('academy/classes/:className')?.name).toBe('TeacherClassStudents')
    expect(findChild('academy/classes/:className/students/:studentId')?.name).toBe('TeacherStudentAnalysis')
    expect(findChild('academy/classes/:className/students/:studentId/review-archive')?.name).toBe('TeacherStudentReviewArchive')
    expect(findChild('academy/instances')?.name).toBe('TeacherInstanceManagement')
    expect(findChild('academy/reports')?.name).toBe('ReportExport')

    expect(findChild('teacher/dashboard')?.redirect).toBeTruthy()
    expect(findChild('teacher/classes')?.redirect).toBeTruthy()
    expect(findChild('teacher/students')?.redirect).toBeTruthy()
    expect(findChild('teacher/classes/:className')?.redirect).toBeTruthy()
    expect(findChild('teacher/classes/:className/students/:studentId')?.redirect).toBeTruthy()
    expect(findChild('teacher/classes/:className/students/:studentId/review-archive')?.redirect).toBeTruthy()
    expect(findChild('teacher/instances')?.redirect).toBeTruthy()
    expect(findChild('teacher/reports')?.redirect).toBeTruthy()
  })

  it('uses platform paths as the canonical location for shared governance pages', () => {
    expect(findChild('platform/challenges')?.name).toBe('ChallengeManage')
    expect(findChild('platform/challenges/package-format')?.name).toBe('AdminChallengePackageFormat')
    expect(findChild('platform/challenges/:id')?.name).toBe('AdminChallengeDetail')
    expect(findChild('platform/challenges/:id/topology')?.name).toBe('AdminChallengeTopologyStudio')
    expect(findChild('platform/challenges/:id/writeup')?.name).toBe('AdminChallengeWriteup')
    expect(findChild('platform/environment-templates')?.name).toBe('AdminEnvironmentTemplateLibrary')
    expect(findChild('platform/images')?.name).toBe('ImageManage')

    expect(findChild('admin/challenges')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/package-format')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/:id')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/:id/topology')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/:id/writeup')?.redirect).toBeTruthy()
    expect(findChild('admin/environment-templates')?.redirect).toBeTruthy()
    expect(findChild('admin/images')?.redirect).toBeTruthy()
  })
})
