import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

const sourceRoot = join(process.cwd(), 'src')

function collectSourceFiles(directory: string): string[] {
  return readdirSync(directory).flatMap((entry) => {
    const path = join(directory, entry)
    const stats = statSync(path)
    if (stats.isDirectory()) {
      if (entry === '__tests__') {
        return []
      }
      return collectSourceFiles(path)
    }
    if (/\.(ts|vue)$/.test(entry)) {
      return [path]
    }
    return []
  })
}

function isTestFile(filePath: string): boolean {
  return (
    filePath.endsWith('.test.ts') ||
    filePath.endsWith('.spec.ts') ||
    filePath.endsWith('.test-harness.ts')
  )
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

  it('feature index.ts should not re-export sibling features as a compatibility barrel', () => {
    const featuresRoot = join(sourceRoot, 'features')

    const violations = collectSourceFiles(featuresRoot)
      .filter((filePath) => filePath.endsWith('/index.ts'))
      .filter((filePath) => {
        const dir = filePath.substring(0, filePath.lastIndexOf('/'))
        const source = readFileSync(filePath, 'utf-8')
        // Match `export * from '../<sibling>'` — a barrel re-exporting from a sibling
        // feature directory, which creates an unwanted cross-slice compatibility layer.
        const siblingExports = Array.from(
          source.matchAll(/export\s+\*\s+from\s+['"]\.\.\/([^'"]+)['"]/g)
        )
        return siblingExports.some((match) => {
          const targetDir = match[1].split('/')[0]
          // Allow re-export from self, disallow re-export from sibling features
          const featureParent = dir.substring(0, dir.lastIndexOf('/'))
          return join(featureParent, targetDir) !== dir
        })
      })
      .map((filePath) => relative(sourceRoot, filePath))

    expect(violations).toEqual([])
  })
})
