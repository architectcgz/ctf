import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

import { frontendArchitecturePolicy } from '@/__tests__/frontendArchitecturePolicy'

const sourceRoot = join(process.cwd(), 'src')
const pagesRoot = join(sourceRoot, 'pages')

function collectSourceFiles(directory: string): string[] {
  return readdirSync(directory).flatMap((entry) => {
    const path = join(directory, entry)
    const stats = statSync(path)
    if (stats.isDirectory()) {
      return collectSourceFiles(path)
    }
    if (/\.(ts|vue)$/.test(entry)) {
      return [path]
    }
    return []
  })
}

function isRoutePageRuntimeFile(filePath: string): boolean {
  return !filePath.endsWith('.test.ts') && !filePath.endsWith('.spec.ts')
}

describe('route page architecture boundaries', () => {
  it('route pages should not import business APIs directly except contracts types', () => {
    const violations = collectSourceFiles(pagesRoot)
      .filter(isRoutePageRuntimeFile)
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const matches = Array.from(source.matchAll(/from\s+['"]@\/api\/([^'"]+)['"]/g))
        return matches
          .map((match) => match[1])
          .filter(
            (importPath) =>
              !frontendArchitecturePolicy.route_page.allow_api_contract_imports_only ||
              !importPath.startsWith('contracts')
          )
          .map((importPath) => `${relative(sourceRoot, filePath)} -> @/api/${importPath}`)
      })

    expect(violations).toEqual([])
  })

  it('route pages should consume features through public APIs instead of internal modules', () => {
    const violations = collectSourceFiles(pagesRoot)
      .filter(isRoutePageRuntimeFile)
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const matches = Array.from(source.matchAll(/from\s+['"](@\/features\/[^'"]+)['"]/g))
        return matches
          .map((match) => match[1])
          .filter((importPath) =>
            /@\/features\/.+\/(model|ui|lib|api|types)(\/|$)/.test(importPath)
          )
          .map((importPath) => `${relative(sourceRoot, filePath)} -> ${importPath}`)
      })

    expect(violations).toEqual([])
  })

  // AppShellRoutePage 是 app shell 的组装根，需要 useRouter 来为 features/layout bridge 注入导航回调。
  // 它不是普通的业务路由页，不适用 route page 对 router hooks 的禁令。
  const routePageHooksExemptFiles = new Set([
    'pages/AppShellRoutePage.vue',
  ])

  it('route pages should not own route navigation and query-tab hooks directly', () => {
    const violations = collectSourceFiles(pagesRoot)
      .filter(isRoutePageRuntimeFile)
      .filter((filePath) => !routePageHooksExemptFiles.has(relative(sourceRoot, filePath)))
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const hits = frontendArchitecturePolicy.route_page.forbidden_runtime_hooks.filter((hook) => {
          switch (hook) {
            case 'useRoute':
              return /\buseRoute\s*\(/.test(source)
            case 'useRouter':
              return /\buseRouter\s*\(/.test(source)
            case 'router.push':
              return /\brouter\.push\s*\(/.test(source)
            case 'router.replace':
              return /\brouter\.replace\s*\(/.test(source)
            case 'useRouteQueryTabs':
              return /\buseRouteQueryTabs\s*\(/.test(source)
            default:
              return false
          }
        })
        return hits.map((hit) => `${relative(sourceRoot, filePath)} -> ${hit}`)
      })

    expect(violations).toEqual([])
  })
})
