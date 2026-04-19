type TeachingWorkspaceRole = string | null | undefined

export function resolveClassManagementRouteName(
  role?: TeachingWorkspaceRole
): 'AdminClassManagement' | 'ClassManagement' {
  return role === 'admin' ? 'AdminClassManagement' : 'ClassManagement'
}

export function resolveTeachingDashboardRouteName(
  role?: TeachingWorkspaceRole
): 'AdminDashboard' | 'TeacherDashboard' {
  return role === 'admin' ? 'AdminDashboard' : 'TeacherDashboard'
}

export function resolveStudentManagementRouteName(
  role?: TeachingWorkspaceRole
): 'AdminStudentManagement' | 'TeacherStudentManagement' {
  return role === 'admin' ? 'AdminStudentManagement' : 'TeacherStudentManagement'
}

export function resolveClassStudentsRouteName(
  role?: TeachingWorkspaceRole
): 'AdminClassStudents' | 'TeacherClassStudents' {
  return role === 'admin' ? 'AdminClassStudents' : 'TeacherClassStudents'
}

export function resolveStudentAnalysisRouteName(
  role?: TeachingWorkspaceRole
): 'AdminStudentAnalysis' | 'TeacherStudentAnalysis' {
  return role === 'admin' ? 'AdminStudentAnalysis' : 'TeacherStudentAnalysis'
}

export function resolveStudentReviewArchiveRouteName(
  role?: TeachingWorkspaceRole
): 'AdminStudentReviewArchive' | 'TeacherStudentReviewArchive' {
  return role === 'admin' ? 'AdminStudentReviewArchive' : 'TeacherStudentReviewArchive'
}

export function resolveAwdReviewIndexRouteName(
  role?: TeachingWorkspaceRole
): 'AdminAWDReviewIndex' | 'TeacherAWDReviewIndex' {
  return role === 'admin' ? 'AdminAWDReviewIndex' : 'TeacherAWDReviewIndex'
}

export function resolveAwdReviewDetailRouteName(
  role?: TeachingWorkspaceRole
): 'AdminAWDReviewDetail' | 'TeacherAWDReviewDetail' {
  return role === 'admin' ? 'AdminAWDReviewDetail' : 'TeacherAWDReviewDetail'
}

export function resolveInstanceManagementRouteName(
  role?: TeachingWorkspaceRole
): 'AdminInstanceManagement' | 'TeacherInstanceManagement' {
  return role === 'admin' ? 'AdminInstanceManagement' : 'TeacherInstanceManagement'
}
