export function platformStudentAnalysisRoute(studentId: string, className: string) {
  return {
    name: 'PlatformStudentAnalysis',
    params: {
      className,
      studentId,
    },
  } as const
}
