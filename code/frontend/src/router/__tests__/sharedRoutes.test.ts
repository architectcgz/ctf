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
  it('uses role-neutral paths as the canonical location for shared profile pages', () => {
    expect(findChild('profile')?.name).toBe('Profile')
    expect(findChild('settings/security')?.name).toBe('SecuritySettings')

    expect(findChild('student/profile')?.redirect).toBeTruthy()
    expect(findChild('student/settings/security')?.redirect).toBeTruthy()
  })

  it('uses academy paths as the canonical location for shared teaching pages', () => {
    expect(findChild('academy/overview')?.name).toBe('TeacherDashboard')
    expect(findChild('academy/classes')?.name).toBe('ClassManagement')
    expect(findChild('academy/students')?.name).toBe('TeacherStudentManagement')
    expect(findChild('academy/awd-reviews')?.name).toBe('TeacherAWDReviewIndex')
    expect(findChild('academy/awd-reviews/:contestId')?.name).toBe('TeacherAWDReviewDetail')
    expect(findChild('academy/classes/:className')?.name).toBe('TeacherClassStudents')
    expect(findChild('academy/classes/:className/students/:studentId')?.name).toBe(
      'TeacherStudentAnalysis'
    )
    expect(findChild('academy/classes/:className/students/:studentId/review-archive')?.name).toBe(
      'TeacherStudentReviewArchive'
    )
    expect(findChild('academy/instances')?.name).toBe('TeacherInstanceManagement')
    expect(findChild('academy/reports')).toBeUndefined()

    expect(findChild('teacher/dashboard')?.redirect).toBeTruthy()
    expect(findChild('teacher/classes')?.redirect).toBeTruthy()
    expect(findChild('teacher/students')?.redirect).toBeTruthy()
    expect(findChild('teacher/classes/:className')?.redirect).toBeTruthy()
    expect(findChild('teacher/classes/:className/students/:studentId')?.redirect).toBeTruthy()
    expect(
      findChild('teacher/classes/:className/students/:studentId/review-archive')?.redirect
    ).toBeTruthy()
    expect(findChild('teacher/instances')?.redirect).toBeTruthy()
    expect(findChild('teacher/reports')).toBeUndefined()
  })

  it('uses platform paths as the canonical location for shared governance pages', () => {
    expect(findChild('platform/students')?.name).toBe('AdminStudentManagement')
    expect(findChild('platform/classes/:className')?.name).toBe('AdminClassStudents')
    expect(findChild('platform/classes/:className/students/:studentId')?.name).toBe(
      'AdminStudentAnalysis'
    )
    expect(
      findChild('platform/classes/:className/students/:studentId/review-archive')?.name
    ).toBe('AdminStudentReviewArchive')
    expect(findChild('platform/awd-reviews')?.name).toBe('AdminAWDReviewIndex')
    expect(findChild('platform/awd-reviews/:contestId')?.name).toBe('AdminAWDReviewDetail')
    expect(findChild('platform/instances')?.name).toBe('AdminInstanceManagement')
    expect(findChild('platform/classes')?.name).toBe('AdminClassManagement')
    expect(findChild('platform/challenges')?.name).toBe('ChallengeManage')
    expect(findChild('platform/challenges/filter-patterns/mock')).toBeUndefined()
    expect(findChild('platform/challenges/package-format')?.name).toBe(
      'AdminChallengePackageFormat'
    )
    expect(findChild('platform/challenges/imports')?.name).toBe('AdminChallengeImportManage')
    expect(findChild('platform/challenges/imports/:importId')?.name).toBe(
      'AdminChallengeImportPreview'
    )
    expect(findChild('platform/challenges/:id')?.name).toBe('AdminChallengeDetail')
    expect(findChild('platform/challenges/:id/topology')?.name).toBe('AdminChallengeTopologyStudio')
    expect(findChild('platform/challenges/:id/writeup')?.name).toBe('AdminChallengeWriteup')
    expect(findChild('platform/challenges/:id/writeup/view')?.name).toBe(
      'AdminChallengeWriteupView'
    )
    expect(findChild('platform/challenges/:id/writeup')?.redirect).toBeFalsy()
    expect(findChild('platform/challenges/:id/writeup/view')?.redirect).toBeFalsy()
    expect(findChild('platform/environment-templates')?.name).toBe(
      'AdminEnvironmentTemplateLibrary'
    )
    expect(findChild('platform/images')?.name).toBe('ImageManage')

    expect(findChild('admin/classes')?.redirect).toBeTruthy()
    expect(findChild('admin/students')?.redirect).toBeTruthy()
    expect(findChild('admin/classes/:className')?.redirect).toBeTruthy()
    expect(findChild('admin/classes/:className/students/:studentId')?.redirect).toBeTruthy()
    expect(
      findChild('admin/classes/:className/students/:studentId/review-archive')?.redirect
    ).toBeTruthy()
    expect(findChild('admin/awd-reviews')?.redirect).toBeTruthy()
    expect(findChild('admin/awd-reviews/:contestId')?.redirect).toBeTruthy()
    expect(findChild('admin/instances')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/package-format')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/imports')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/imports/:importId')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/:id')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/:id/topology')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/:id/writeup')?.redirect).toBeTruthy()
    expect(findChild('admin/challenges/:id/writeup/view')?.redirect).toBeTruthy()
    expect(findChild('admin/environment-templates')?.redirect).toBeTruthy()
    expect(findChild('admin/images')?.redirect).toBeTruthy()
  })

  it('keeps contest management on admin paths and exposes a dedicated contest edit route', () => {
    expect(findChild('admin/contests')?.name).toBe('ContestManage')
    expect(findChild('admin/contests/:id/edit')?.name).toBe('ContestEdit')
    expect(findChild('platform/contests')?.redirect).toBeTruthy()
  })
})
