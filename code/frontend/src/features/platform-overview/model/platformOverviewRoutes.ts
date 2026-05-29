export function buildPlatformAuditLogRoute(query?: Record<string, string>) {
  return query
    ? ({ name: 'AuditLog', query } as const)
    : ({ name: 'AuditLog' } as const)
}

export const platformCheatDetectionRoute = {
  name: 'CheatDetection',
} as const
