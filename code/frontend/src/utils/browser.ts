export function redirectTo(url: string): void {
  window.location.assign(url)
}

export function getNavigationType(): string | null {
  const entry = window.performance.getEntriesByType('navigation')[0] as
    | PerformanceNavigationTiming
    | undefined
  return entry?.type ?? null
}

export function reloadPage(): void {
  window.location.reload()
}
