import type { RouteRecordRaw } from 'vue-router'

export function redirectWithQuery(path: string): NonNullable<RouteRecordRaw['redirect']> {
  return (to) => ({
    path,
    query: to.query,
    hash: to.hash,
  })
}
