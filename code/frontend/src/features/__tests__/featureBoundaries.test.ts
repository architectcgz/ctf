import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

const sourceRoot = join(process.cwd(), 'src')

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

function isTestFile(filePath: string): boolean {
  return filePath.endsWith('.test.ts') || filePath.endsWith('.spec.ts')
}

function extractImportSpecifiers(source: string): string[] {
  return Array.from(source.matchAll(/from\s+['"]([^'"]+)['"]/g)).map((match) => match[1])
}

describe('feature boundaries', () => {
  it('feature runtime sources should not import components layer', () => {
    const featuresRoot = join(sourceRoot, 'features')
    const violations = collectSourceFiles(featuresRoot)
      .filter((filePath) => !isTestFile(filePath))
      .filter((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        return /from\s+['"]@\/components\//.test(source)
      })
      .map((filePath) => relative(sourceRoot, filePath))

    expect(violations).toEqual([])
  })

  it('runtime sources should not deep import another feature model via alias path', () => {
    const violations = collectSourceFiles(sourceRoot)
      .filter((filePath) => !isTestFile(filePath))
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const imports = extractImportSpecifiers(source)
        const importerRelative = relative(sourceRoot, filePath)
        const importerFeature = importerRelative.startsWith('features/')
          ? importerRelative.split('/')[1]
          : null

        return imports
          .filter((importPath) => importPath.startsWith('@/features/'))
          .filter((importPath) => /@\/features\/[^/]+\/model\//.test(importPath))
          .filter((importPath) => {
            const importedFeature = importPath.replace('@/features/', '').split('/')[0]
            // Same feature internal imports are allowed; cross-feature deep imports are not.
            return !(importerFeature && importedFeature === importerFeature)
          })
          .map((importPath) => `${importerRelative} -> ${importPath}`)
      })

    expect(violations).toEqual([])
  })
})
