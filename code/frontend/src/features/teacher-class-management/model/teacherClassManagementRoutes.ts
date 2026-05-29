export const teacherDashboardRoute = {
  name: 'TeacherDashboard',
} as const

export function teacherClassStudentsRoute(className: string) {
  return {
    name: 'TeacherClassStudents',
    params: { className },
  } as const
}
