export function reportFrontendError(scope: string, error: unknown): void {
  if (!import.meta.env.DEV) {
    return
  }

  console.error(scope, error)
}
