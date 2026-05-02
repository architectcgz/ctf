import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

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
          .filter((importPath) => !importPath.startsWith('contracts'))
          .map((importPath) => `${relative(sourceRoot, filePath)} -> @/api/${importPath}`)
      })

    expect(violations).toEqual([])
  })

  it('views should not own route navigation and query-tab hooks directly', () => {
    const violations = collectSourceFiles(viewsRoot)
      .filter(isViewRuntimeFile)
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const hits = [
          /\buseRoute\s*\(/.test(source) ? 'useRoute' : '',
          /\buseRouter\s*\(/.test(source) ? 'useRouter' : '',
          /\brouter\.push\s*\(/.test(source) ? 'router.push' : '',
          /\brouter\.replace\s*\(/.test(source) ? 'router.replace' : '',
          /\buseRouteQueryTabs\s*\(/.test(source) ? 'useRouteQueryTabs' : '',
        ].filter(Boolean)
        return hits.map((hit) => `${relative(sourceRoot, filePath)} -> ${hit}`)
      })

    expect(violations).toEqual([])
  })
})
