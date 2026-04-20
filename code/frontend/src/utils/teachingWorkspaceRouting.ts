type TeachingWorkspaceRole = string | null | undefined

export function resolveClassManagementRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformClassManagement' | 'ClassManagement' {
  return role === 'admin' ? 'PlatformClassManagement' : 'ClassManagement'
}

export function resolveTeachingDashboardRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformOverview' | 'TeacherDashboard' {
  return role === 'admin' ? 'PlatformOverview' : 'TeacherDashboard'
}

export function resolveStudentManagementRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformStudentManagement' | 'TeacherStudentManagement' {
  return role === 'admin' ? 'PlatformStudentManagement' : 'TeacherStudentManagement'
}

export function resolveClassStudentsRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformClassStudents' | 'TeacherClassStudents' {
  return role === 'admin' ? 'PlatformClassStudents' : 'TeacherClassStudents'
}

export function resolveStudentAnalysisRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformStudentAnalysis' | 'TeacherStudentAnalysis' {
  return role === 'admin' ? 'PlatformStudentAnalysis' : 'TeacherStudentAnalysis'
}

export function resolveStudentReviewArchiveRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformStudentReviewArchive' | 'TeacherStudentReviewArchive' {
  return role === 'admin' ? 'PlatformStudentReviewArchive' : 'TeacherStudentReviewArchive'
}

export function resolveAwdReviewIndexRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformAwdReviewIndex' | 'TeacherAWDReviewIndex' {
  return role === 'admin' ? 'PlatformAwdReviewIndex' : 'TeacherAWDReviewIndex'
}

export function resolveAwdReviewDetailRouteName(
  role?: TeachingWorkspaceRole
): 'AdminAwdReplay' | 'TeacherAwdOverview' {
  return role === 'admin' ? 'AdminAwdReplay' : 'TeacherAwdOverview'
}

export function resolveInstanceManagementRouteName(
  role?: TeachingWorkspaceRole
): 'PlatformInstanceManagement' | 'TeacherInstanceManagement' {
  return role === 'admin' ? 'PlatformInstanceManagement' : 'TeacherInstanceManagement'
}
