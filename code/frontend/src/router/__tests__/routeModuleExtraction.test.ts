import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const routerIndexSource = readFileSync(`${process.cwd()}/src/router/index.ts`, 'utf-8')

describe('router route module extraction', () => {
  it('router/index.ts 应只负责装配分组后的路由模块', () => {
    expect(routerIndexSource).toContain("import { authRoutes } from './routes/authRoutes'")
    expect(routerIndexSource).toContain("import { appShellRoute } from './routes/appShellRoute'")
    expect(routerIndexSource).toContain("import { errorRoutes } from './routes/errorRoutes'")
    expect(routerIndexSource).toContain("import { utilityRoutes } from './routes/utilityRoutes'")

    expect(routerIndexSource).toContain('...authRoutes')
    expect(routerIndexSource).toContain('appShellRoute')
    expect(routerIndexSource).toContain('...errorRoutes')
    expect(routerIndexSource).toContain('...utilityRoutes')

    expect(routerIndexSource).not.toContain("path: 'academy/overview'")
    expect(routerIndexSource).not.toContain("path: 'platform/challenges'")
  })
})
