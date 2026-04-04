import { redirectTo } from '@/utils/browser'

const ERROR_STATUS_ROUTE_MAP: Record<number, string> = {
  401: '/401',
  429: '/429',
  500: '/500',
  502: '/502',
  503: '/503',
  504: '/504',
}

const AUTH_FLOW_PREFIXES = ['/auth/login', '/auth/register']

export function resolveErrorStatusPage(status: number | undefined): string | null {
  if (!status) return null
  return ERROR_STATUS_ROUTE_MAP[status] ?? null
}

export function isAuthFlowRequest(url: string | undefined): boolean {
  if (!url) return false
  return AUTH_FLOW_PREFIXES.some((prefix) => url.includes(prefix))
}

export function shouldRedirectToErrorStatusPage(
  status: number | undefined,
  requestUrl?: string
): boolean {
  const route = resolveErrorStatusPage(status)
  if (!route) return false
  if (status === 401 && isAuthFlowRequest(requestUrl)) return false
  return true
}

export function redirectToErrorStatusPage(
  status: number | undefined,
  requestUrl?: string
): boolean {
  if (!shouldRedirectToErrorStatusPage(status, requestUrl)) return false
  const route = resolveErrorStatusPage(status)
  if (!route) return false
  if (window.location.pathname === route) return false
  redirectTo(route)
  return true
}
