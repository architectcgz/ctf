import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join, normalize, relative, resolve, sep } from 'node:path'

import { describe, expect, it } from 'vitest'

import {
  commonForbiddenImportAllowlist,
  componentFeatureImportAllowlist,
  componentNonContractApiAllowlist,
  composableMultiBoundaryAllowlist,
  featureRouterImportAllowlist,
  legacyComponentPageAllowlist,
  oversizedViewAllowlist,
  utilityBoundaryImportAllowlist,
  viewLineLimit,
  widgetLegacyComponentImportAllowlist,
  widgetNonContractApiAllowlist,
} from './architectureAllowlist'

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

function collectImportKeys(files: SourceFile[], importPathPrefix: string): string[] {
  return files.flatMap((file) => {
    const source = readFileSync(file.absolutePath, 'utf-8')
    return extractImports(source)
      .filter((importPath) => importPath.startsWith(importPathPrefix))
      .map((importPath) => `${file.relativePath} -> ${importPath}`)
  })
}

function collectImportsMatching(
  files: SourceFile[],
  predicate: (file: SourceFile, importPath: string) => boolean
): string[] {
  return files.flatMap((file) => {
    const source = readFileSync(file.absolutePath, 'utf-8')
    return extractImports(source)
      .filter((importPath) => predicate(file, importPath))
      .map((importPath) => `${file.relativePath} -> ${importPath}`)
  })
}

function expectBaseline(
  actualEntries: string[],
  allowlist: Set<string>,
  violationLabel: string
): void {
  const violations = actualEntries.filter((key) => !allowlist.has(key))
  const staleAllowlistEntries = Array.from(allowlist).filter((key) => !actualEntries.includes(key))

  expect(violations, violationLabel).toEqual([])
  expect(staleAllowlistEntries, `${violationLabel} stale allowlist`).toEqual([])
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

  it('legacy business components should not add new direct feature imports', () => {
    const componentFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`components${sep}`)
    )
    const featureImports = collectImportKeys(componentFiles, '@/features')
    expectBaseline(featureImports, componentFeatureImportAllowlist, 'component feature imports')
  })

  it('widgets should not add new dependencies on legacy business component directories', () => {
    const widgetFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`widgets${sep}`))
    const legacyComponentImports = widgetFiles.flatMap((file) => {
      const source = readFileSync(file.absolutePath, 'utf-8')
      return extractImports(source)
        .filter((importPath) =>
          /^@\/components\/(platform|teacher|notifications|challenge|dashboard|contests|scoreboard)/.test(
            importPath
          )
        )
        .map((importPath) => `${file.relativePath} -> ${importPath}`)
    })
    expectBaseline(
      legacyComponentImports,
      widgetLegacyComponentImportAllowlist,
      'widget legacy component imports'
    )
  })

  it('new route views should stay below the page-size threshold', () => {
    const viewFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`views${sep}`))
    const oversizedViews = viewFiles
      .map((file) => ({
        file: file.relativePath,
        lines: readFileSync(file.absolutePath, 'utf-8').split(/\r?\n/).length,
      }))
      .filter(({ lines }) => lines > viewLineLimit)

    const violations = oversizedViews
      .filter(({ file }) => !oversizedViewAllowlist.has(file))
      .map(({ file, lines }) => `${file} has ${lines} lines`)
    const staleAllowlistEntries = Array.from(oversizedViewAllowlist).filter(
      (file) => !oversizedViews.some((view) => view.file === file)
    )

    expect(violations).toEqual([])
    expect(staleAllowlistEntries).toEqual([])
  })

  it('components and widgets should not add new non-contract API imports', () => {
    const componentFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`components${sep}`)
    )
    const widgetFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`widgets${sep}`))

    const componentApiImports = collectImportKeys(componentFiles, '@/api/').filter(
      (key) => !key.includes(' -> @/api/contracts')
    )
    const widgetApiImports = collectImportKeys(widgetFiles, '@/api/').filter(
      (key) => !key.includes(' -> @/api/contracts')
    )

    expectBaseline(componentApiImports, componentNonContractApiAllowlist, 'component API imports')
    expectBaseline(widgetApiImports, widgetNonContractApiAllowlist, 'widget API imports')
  })

  it('common and entity layers should stay free of app services, router, and stores', () => {
    const lowLevelFiles = sourceFiles.filter(
      (file) =>
        file.relativePath.startsWith(`components${sep}common${sep}`) ||
        file.relativePath.startsWith(`entities${sep}`)
    )
    const forbiddenImports = lowLevelFiles.flatMap((file) => {
      const source = readFileSync(file.absolutePath, 'utf-8')
      return extractImports(source)
        .filter(
          (importPath) =>
            /^@\/(api|features|widgets|views|stores|router)/.test(importPath) ||
            importPath === 'vue-router' ||
            importPath === 'pinia'
        )
        .map((importPath) => `${file.relativePath} -> ${importPath}`)
    })

    expectBaseline(forbiddenImports, commonForbiddenImportAllowlist, 'low-level forbidden imports')
  })

  it('feature UI files should not import non-contract API modules directly', () => {
    const featureUiFiles = sourceFiles.filter(
      (file) =>
        file.relativePath.startsWith(`features${sep}`) &&
        file.relativePath.includes(`${sep}ui${sep}`)
    )
    const apiImports = collectImportKeys(featureUiFiles, '@/api/').filter(
      (key) => !key.includes(' -> @/api/contracts')
    )

    expect(apiImports).toEqual([])
  })

  it('new page components should be route views or widgets instead of legacy component pages', () => {
    const componentPageFiles = sourceFiles
      .map((file) => file.relativePath)
      .filter((relativePath) => relativePath.startsWith(`components${sep}`))
      .filter((relativePath) => /Page\.vue$/.test(relativePath))

    expectBaseline(componentPageFiles, legacyComponentPageAllowlist, 'legacy component page files')
  })

  it('stores and utilities should not depend on UI or app orchestration layers', () => {
    const storeFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`stores${sep}`))
    const storeForbiddenImports = collectImportsMatching(storeFiles, (_file, importPath) =>
      /^@\/(views|components)/.test(importPath)
    )

    const utilityFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`utils${sep}`))
    const utilityForbiddenImports = collectImportsMatching(
      utilityFiles,
      (_file, importPath) =>
        /^@\/(api|features|widgets|views|stores|router|components)/.test(importPath) ||
        importPath === 'vue' ||
        importPath === 'vue-router' ||
        importPath === 'pinia'
    )

    expect(storeForbiddenImports).toEqual([])
    expectBaseline(utilityForbiddenImports, utilityBoundaryImportAllowlist, 'utility imports')
  })

  it('feature router access should stay in reviewed route-aware composables', () => {
    const featureFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`features${sep}`)
    )
    const routerImports = collectImportsMatching(
      featureFiles,
      (_file, importPath) => importPath === 'vue-router' || importPath.startsWith('@/router')
    )

    expectBaseline(routerImports, featureRouterImportAllowlist, 'feature router imports')
  })

  it('shared composables should not mix API, router, store, and multiple feature owners', () => {
    const composableFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`composables${sep}`)
    )
    const mixedComposables = composableFiles
      .map((file) => {
        const imports = extractImports(readFileSync(file.absolutePath, 'utf-8'))
        const flags: string[] = []
        if (imports.some((importPath) => importPath.startsWith('@/api/'))) flags.push('api')
        if (
          imports.some(
            (importPath) => importPath === 'vue-router' || importPath.startsWith('@/router')
          )
        ) {
          flags.push('router')
        }
        if (
          imports.some((importPath) => importPath === 'pinia' || importPath.startsWith('@/stores'))
        ) {
          flags.push('store')
        }
        const featureOwners = new Set(
          imports
            .filter((importPath) => importPath.startsWith('@/features/'))
            .map((importPath) => importPath.split('/').slice(0, 3).join('/'))
        )
        if (featureOwners.size > 1) flags.push(`features:${featureOwners.size}`)
        return flags.length > 1 ? `${file.relativePath} -> ${flags.join('+')}` : ''
      })
      .filter(Boolean)

    expectBaseline(mixedComposables, composableMultiBoundaryAllowlist, 'mixed shared composables')
  })
})
