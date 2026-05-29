export type AppRouteParamValue = string | number | null | undefined

export interface AppNamedRouteTarget {
  name: string
  params?: Record<string, AppRouteParamValue>
  query?: Record<string, AppRouteParamValue>
  hash?: string
}

export interface AppPathRouteTarget {
  path: string
  query?: Record<string, AppRouteParamValue>
  hash?: string
}

export type AppRouteTarget = string | AppNamedRouteTarget | AppPathRouteTarget
