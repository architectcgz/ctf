type TeachingWorkspaceRole = string | null | undefined
export type ClassWorkspaceSectionKey = 'trend' | 'review' | 'insights' | 'intervention'

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

export function resolveClassWorkspaceSectionRouteName(
  role: TeachingWorkspaceRole,
  section: ClassWorkspaceSectionKey
):
  | 'AdminClassTrend'
  | 'AdminClassReview'
  | 'AdminClassInsights'
  | 'AdminClassIntervention'
  | 'TeacherClassTrend'
  | 'TeacherClassReview'
  | 'TeacherClassInsights'
  | 'TeacherClassIntervention' {
  if (role === 'admin') {
    switch (section) {
      case 'trend':
        return 'AdminClassTrend'
      case 'review':
        return 'AdminClassReview'
      case 'insights':
        return 'AdminClassInsights'
      case 'intervention':
        return 'AdminClassIntervention'
    }
  }

  switch (section) {
    case 'trend':
      return 'TeacherClassTrend'
    case 'review':
      return 'TeacherClassReview'
    case 'insights':
      return 'TeacherClassInsights'
    case 'intervention':
      return 'TeacherClassIntervention'
  }
}

export function resolveClassWorkspaceSectionKeyFromRouteName(
  routeName: unknown
): ClassWorkspaceSectionKey | null {
  switch (routeName) {
    case 'TeacherClassTrend':
    case 'AdminClassTrend':
      return 'trend'
    case 'TeacherClassReview':
    case 'AdminClassReview':
      return 'review'
    case 'TeacherClassInsights':
    case 'AdminClassInsights':
      return 'insights'
    case 'TeacherClassIntervention':
    case 'AdminClassIntervention':
      return 'intervention'
    default:
      return null
  }
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
