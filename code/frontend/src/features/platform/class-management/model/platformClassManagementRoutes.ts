export function platformClassStudentsRoute(className: string) {
  return {
    name: 'PlatformClassStudents',
    params: { className },
  } as const
}
