import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join, normalize, relative, resolve, sep } from 'node:path'

import { describe, expect, it } from 'vitest'

const sourceRoot = join(process.cwd(), 'src')

type Layer = 'common' | 'entities' | 'features' | 'widgets' | 'views' | 'other'

interface SourceFile {
  absolutePath: string
  relativePath: string
  layer: Layer
}

const importPattern =
  /(?:import|export)\s+(?:type\s+)?(?:[^'"]*?\s+from\s+)?['"]([^'"]+)['"]|import\s*\(\s*['"]([^'"]+)['"]\s*\)/g

function collectSourceFiles(directory: string): SourceFile[] {
  return readdirSync(directory).flatMap((entry) => {
    const absolutePath = join(directory, entry)
    const stats = statSync(absolutePath)
    if (stats.isDirectory()) {
      if (entry === '__tests__') {
        return []
      }
      return collectSourceFiles(absolutePath)
    }
    if (!/\.(ts|vue)$/.test(entry) || /\.d\.ts$/.test(entry) || /\.test\.ts$/.test(entry)) {
      return []
    }

    const relativePath = normalize(relative(sourceRoot, absolutePath))
    return [{ absolutePath, relativePath, layer: classifyLayer(relativePath) }]
  })
}

function classifyLayer(relativePath: string): Layer {
  if (relativePath.startsWith(`components${sep}common${sep}`)) return 'common'
  if (relativePath.startsWith(`entities${sep}`)) return 'entities'
  if (relativePath.startsWith(`features${sep}`)) return 'features'
  if (relativePath.startsWith(`widgets${sep}`)) return 'widgets'
  if (relativePath.startsWith(`views${sep}`)) return 'views'
  return 'other'
}

function extractImports(source: string): string[] {
  return Array.from(source.matchAll(importPattern))
    .map((match) => match[1] ?? match[2])
    .filter(Boolean)
}

function resolveImportLayer(fromFile: SourceFile, importPath: string): Layer | null {
  if (importPath.startsWith('@/')) {
    return classifyLayer(normalize(importPath.slice(2)))
  }
  if (!importPath.startsWith('.')) {
    return null
  }

  const resolvedPath = normalize(
    relative(sourceRoot, resolve(dirname(fromFile.absolutePath), importPath))
  )
  if (resolvedPath.startsWith('..')) {
    return null
  }
  return classifyLayer(resolvedPath)
}

function collectLayerViolations(
  files: SourceFile[],
  forbiddenByLayer: Record<Layer, Layer[]>
): string[] {
  return files.flatMap((file) => {
    const forbiddenLayers = forbiddenByLayer[file.layer] ?? []
    if (forbiddenLayers.length === 0) {
      return []
    }

    const source = readFileSync(file.absolutePath, 'utf-8')
    return extractImports(source)
      .map((importPath) => ({ importPath, importedLayer: resolveImportLayer(file, importPath) }))
      .filter(
        ({ importedLayer }) => importedLayer !== null && forbiddenLayers.includes(importedLayer)
      )
      .map(
        ({ importPath, importedLayer }) =>
          `${file.relativePath} -> ${importPath} (${importedLayer})`
      )
  })
}

describe('frontend architecture boundaries', () => {
  const sourceFiles = collectSourceFiles(sourceRoot)

  it('lower frontend layers should not import higher product layers', () => {
    const violations = collectLayerViolations(sourceFiles, {
      common: ['features', 'widgets', 'views'],
      entities: ['features', 'widgets', 'views'],
      features: ['widgets', 'views'],
      widgets: ['views'],
      views: [],
      other: [],
    })

    expect(violations).toEqual([])
  })
})
