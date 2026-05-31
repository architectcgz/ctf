export function sanitizeRedirectPath(input: unknown): string {
  if (typeof input !== 'string') return '/'
  if (/^\s*$/.test(input)) return '/'
  if (/^(?:[a-z][a-z0-9+\-.]*:)?\/\//i.test(input) || input.startsWith('/\\')) return '/'
  const normalized = '/' + input.replace(/^\/+/, '')
  if (/^\/teacher(?:\/|$)/u.test(normalized)) {
    return '/'
  }
  return normalized
}
