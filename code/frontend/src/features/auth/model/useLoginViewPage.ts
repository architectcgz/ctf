export const DEFAULT_LOGIN_REDIRECT = '/'

export function resolveLoginRedirectTarget(redirectTo: string, defaultTarget: string): string {
  return redirectTo === DEFAULT_LOGIN_REDIRECT ? defaultTarget : redirectTo
}
