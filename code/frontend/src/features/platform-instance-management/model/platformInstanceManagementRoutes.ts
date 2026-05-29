export const platformOverviewRoute = {
  name: 'PlatformOverview',
} as const

export function platformInstanceStudentAnalysisRoute(studentId: string, className: string) {
  return {
    name: 'PlatformStudentAnalysis',
    params: {
      className,
      studentId,
    },
  } as const
}
