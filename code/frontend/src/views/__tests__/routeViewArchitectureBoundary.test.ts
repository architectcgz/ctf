import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

import { frontendArchitecturePolicy } from '@/__tests__/architectureAllowlist'

const sourceRoot = join(process.cwd(), 'src')
const viewsRoot = join(sourceRoot, 'views')

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

function isViewRuntimeFile(filePath: string): boolean {
  if (filePath.includes(`${join('views', '__tests__')}`)) {
    return false
  }
  return !filePath.endsWith('.test.ts') && !filePath.endsWith('.spec.ts')
}

describe('route view architecture boundaries', () => {
  it('views should not import business APIs directly except contracts types', () => {
    const violations = collectSourceFiles(viewsRoot)
      .filter(isViewRuntimeFile)
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const matches = Array.from(source.matchAll(/from\s+['"]@\/api\/([^'"]+)['"]/g))
        return matches
          .map((match) => match[1])
          .filter(
            (importPath) =>
              !frontendArchitecturePolicy.route_view.allow_api_contract_imports_only ||
              !importPath.startsWith('contracts')
          )
          .map((importPath) => `${relative(sourceRoot, filePath)} -> @/api/${importPath}`)
      })

    expect(violations).toEqual([])
  })

  it('views should not own route navigation and query-tab hooks directly', () => {
    const violations = collectSourceFiles(viewsRoot)
      .filter(isViewRuntimeFile)
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const hits = frontendArchitecturePolicy.route_view.forbidden_runtime_hooks.filter((hook) => {
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
