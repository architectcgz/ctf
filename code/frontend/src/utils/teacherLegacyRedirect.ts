function encodeRouteParam(value: unknown): string {
  return encodeURIComponent(String(value || ''))
}

interface TeacherLegacyRedirectDefinition {
  legacyPath: string
  matchPath: RegExp
  buildCanonicalPathFromCaptures: (...captures: string[]) => string
  buildCanonicalPathFromParams: (params: Record<string, unknown>) => string
}

function createStaticTeacherLegacyRedirectDefinition(
  legacyPath: string,
  canonicalPath: string
): TeacherLegacyRedirectDefinition {
  const pathWithoutLeadingSlash = legacyPath.replace(/^\/+/, '')
  return {
    legacyPath: pathWithoutLeadingSlash,
    matchPath: new RegExp(`^/${pathWithoutLeadingSlash}/?$`, 'u'),
    buildCanonicalPathFromCaptures: () => canonicalPath,
    buildCanonicalPathFromParams: () => canonicalPath,
  }
}

function createDynamicTeacherLegacyRedirectDefinition(
  legacyPath: string,
  matchPath: RegExp,
  buildCanonicalPathFromCaptures: (...captures: string[]) => string,
  buildCanonicalPathFromParams: (params: Record<string, unknown>) => string
): TeacherLegacyRedirectDefinition {
  return {
    legacyPath: legacyPath.replace(/^\/+/, ''),
    matchPath,
    buildCanonicalPathFromCaptures,
    buildCanonicalPathFromParams,
  }
}

export const teacherLegacyRedirectDefinitions: TeacherLegacyRedirectDefinition[] = [
  createStaticTeacherLegacyRedirectDefinition('teacher/dashboard', '/academy/overview'),
  createStaticTeacherLegacyRedirectDefinition('teacher/classes', '/academy/classes'),
  createStaticTeacherLegacyRedirectDefinition('teacher/students', '/academy/students'),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/classes/:className',
    /^\/teacher\/classes\/([^/?#]+)\/?$/u,
    (className) => `/academy/classes/${className}`,
    (params) => `/academy/classes/${encodeRouteParam(params.className)}`
  ),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/classes/:className/trend',
    /^\/teacher\/classes\/([^/?#]+)\/trend\/?$/u,
    (className) => `/academy/classes/${className}/trend`,
    (params) => `/academy/classes/${encodeRouteParam(params.className)}/trend`
  ),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/classes/:className/review',
    /^\/teacher\/classes\/([^/?#]+)\/review\/?$/u,
    (className) => `/academy/classes/${className}/review`,
    (params) => `/academy/classes/${encodeRouteParam(params.className)}/review`
  ),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/classes/:className/insights',
    /^\/teacher\/classes\/([^/?#]+)\/insights\/?$/u,
    (className) => `/academy/classes/${className}/insights`,
    (params) => `/academy/classes/${encodeRouteParam(params.className)}/insights`
  ),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/classes/:className/intervention',
    /^\/teacher\/classes\/([^/?#]+)\/intervention\/?$/u,
    (className) => `/academy/classes/${className}/intervention`,
    (params) => `/academy/classes/${encodeRouteParam(params.className)}/intervention`
  ),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/classes/:className/students/:studentId',
    /^\/teacher\/classes\/([^/?#]+)\/students\/([^/?#]+)\/?$/u,
    (className, studentId) => `/academy/classes/${className}/students/${studentId}`,
    (params) =>
      `/academy/classes/${encodeRouteParam(params.className)}/students/${encodeRouteParam(params.studentId)}`
  ),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/classes/:className/students/:studentId/review-archive',
    /^\/teacher\/classes\/([^/?#]+)\/students\/([^/?#]+)\/review-archive\/?$/u,
    (className, studentId) =>
      `/academy/classes/${className}/students/${studentId}/review-archive`,
    (params) =>
      `/academy/classes/${encodeRouteParam(params.className)}/students/${encodeRouteParam(params.studentId)}/review-archive`
  ),
  createStaticTeacherLegacyRedirectDefinition('teacher/awd-reviews', '/academy/awd-reviews'),
  createDynamicTeacherLegacyRedirectDefinition(
    'teacher/awd-reviews/:contestId',
    /^\/teacher\/awd-reviews\/([^/?#]+)\/?$/u,
    (contestId) => `/academy/awd-reviews/${contestId}`,
    (params) => `/academy/awd-reviews/${encodeRouteParam(params.contestId)}`
  ),
  createStaticTeacherLegacyRedirectDefinition('teacher/instances', '/academy/instances'),
]

export const teacherLegacyRedirectAllowlist = teacherLegacyRedirectDefinitions.map(
  ({ legacyPath }) => legacyPath
)

export function canonicalizeTeacherLegacyRedirectPath(path: string): string {
  const match = path.match(/^([^?#]*)(.*)$/u)
  const pathname = match?.[1] ?? path
  const suffix = match?.[2] ?? ''

  for (const definition of teacherLegacyRedirectDefinitions) {
    const pathMatch = pathname.match(definition.matchPath)
    if (!pathMatch) {
      continue
    }
    return definition.buildCanonicalPathFromCaptures(...pathMatch.slice(1)) + suffix
  }

  return path
}
