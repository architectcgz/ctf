import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join, normalize, relative, resolve, sep } from 'node:path'

import { describe, expect, it } from 'vitest'

import {
  frontendArchitecturePolicy,
  type FrontendArchitecturePolicy,
  type FrontendArchitectureLayer,
  routePageLineLimit,
} from './frontendArchitecturePolicy'

const sourceRoot = join(process.cwd(), 'src')

type Layer = FrontendArchitectureLayer | 'other'

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
  for (const [layer, config] of Object.entries(frontendArchitecturePolicy.layers) as Array<
    [FrontendArchitectureLayer, FrontendArchitecturePolicy['layers'][FrontendArchitectureLayer]]
  >) {
    if (
      config.relative_prefixes.some((prefix) =>
        relativePath.startsWith(normalize(prefix).replaceAll('/', sep))
      )
    ) {
      return layer
    }
  }
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

function collectSourceMatches(
  files: SourceFile[],
  predicate: (file: SourceFile, source: string) => boolean
): string[] {
  return files
    .filter((file) => predicate(file, readFileSync(file.absolutePath, 'utf-8')))
    .map((file) => file.relativePath)
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
      common: frontendArchitecturePolicy.layers.common.forbidden_import_layers,
      shared: frontendArchitecturePolicy.layers.shared.forbidden_import_layers,
      entities: frontendArchitecturePolicy.layers.entities.forbidden_import_layers,
      pages: frontendArchitecturePolicy.layers.pages.forbidden_import_layers,
      features: frontendArchitecturePolicy.layers.features.forbidden_import_layers,
      widgets: frontendArchitecturePolicy.layers.widgets.forbidden_import_layers,
      other: [],
    })

    expect(violations).toEqual([])
  })

  it('legacy business components should not add new direct feature imports', () => {
    const componentFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`components${sep}`)
    )
    const featureImports = collectImportKeys(componentFiles, '@/features')
    expect(featureImports).toEqual([])
  })

  it('widgets should not add new dependencies on legacy business component directories', () => {
    const widgetFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`widgets${sep}`))
    const legacyComponentImports = widgetFiles.flatMap((file) => {
      const source = readFileSync(file.absolutePath, 'utf-8')
      const legacyDirectoryPattern = frontendArchitecturePolicy.legacy_business_component_directories.join(
        '|'
      )
      return extractImports(source)
        .filter((importPath) =>
          new RegExp(`^@/components/(${legacyDirectoryPattern})`).test(
            importPath
          )
        )
        .map((importPath) => `${file.relativePath} -> ${importPath}`)
    })
    expect(legacyComponentImports).toEqual([])
  })

  it('route pages should live only in the dedicated pages layer', () => {
    const routePageSuffix = frontendArchitecturePolicy.route_page_suffix
    const routePageLayer = frontendArchitecturePolicy.route_page_layer
    const routePageFiles = sourceFiles.filter((file) => file.relativePath.endsWith(routePageSuffix))

    const violations = routePageFiles
      .filter((file) => file.layer !== routePageLayer)
      .map((file) => `${file.relativePath} -> ${file.layer}`)

    expect(violations).toEqual([])
  })

  it('router runtime entries should point to the dedicated pages layer only', () => {
    const routerFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`router${sep}`))
    const routeComponentImportPattern = /component:\s*\(\)\s*=>\s*import\(\s*['"](@\/[^'"]+)['"]\s*\)/g
    const routerComponentAllowlist = new Set(['@/shared/ui/layout/AppLayout.vue'])

    const violations = routerFiles.flatMap((file) => {
      const source = readFileSync(file.absolutePath, 'utf-8')
      return Array.from(source.matchAll(routeComponentImportPattern))
        .map((match) => match[1])
        .filter(
          (importPath) =>
            !importPath.startsWith('@/pages/') && !routerComponentAllowlist.has(importPath)
        )
        .map((importPath) => `${file.relativePath} -> ${importPath}`)
    })

    expect(violations).toEqual([])
  })

  it('new route pages should stay below the page-size threshold', () => {
    const routePageFiles = sourceFiles.filter(
      (file) =>
        file.relativePath.startsWith(`pages${sep}`) &&
        file.relativePath.endsWith(frontendArchitecturePolicy.route_page_suffix)
    )
    const oversizedRoutePages = routePageFiles
      .map((file) => ({
        file: file.relativePath,
        lines: readFileSync(file.absolutePath, 'utf-8').split(/\r?\n/).length,
      }))
      .filter(({ lines }) => lines > routePageLineLimit)

    const violations = oversizedRoutePages
      .map(({ file, lines }) => `${file} has ${lines} lines`)

    expect(violations).toEqual([])
  })

  it('components and widgets should not add new non-contract API imports', () => {
    const componentFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`components${sep}`)
    )
    const widgetFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`widgets${sep}`))

    const componentApiImports = collectImportKeys(componentFiles, '@/api/').filter(
      (key) =>
        !key.includes(' -> @/api/contracts')
    )
    const widgetApiImports = collectImportKeys(widgetFiles, '@/api/').filter(
      (key) =>
        !key.includes(' -> @/api/contracts')
    )

    expect(componentApiImports).toEqual([])
    expect(widgetApiImports).toEqual([])
  })

  it('common, shared lib, and entity layers should stay free of app services, router, and stores', () => {
    const lowLevelFiles = sourceFiles.filter(
      (file) =>
        file.relativePath.startsWith(`shared${sep}ui${sep}common${sep}`) ||
        file.relativePath.startsWith(`shared${sep}model${sep}common${sep}`) ||
        file.relativePath.startsWith(`shared${sep}lib${sep}`) ||
        file.relativePath.startsWith(`entities${sep}`)
    )
    const allowedBareImportsByFile = new Map<string, string[]>([
      [`shared${sep}lib${sep}navigation${sep}routeTarget.ts`, ['vue-router']],
    ])
    const forbiddenImports = lowLevelFiles.flatMap((file) => {
      const source = readFileSync(file.absolutePath, 'utf-8')
      return extractImports(source)
        .filter(
          (importPath) =>
            !allowedBareImportsByFile.get(file.relativePath)?.includes(importPath) &&
            frontendArchitecturePolicy.low_level_forbidden_import_prefixes.some((prefix) =>
              importPath.startsWith(prefix)
            ) ||
            (!allowedBareImportsByFile.get(file.relativePath)?.includes(importPath) &&
              frontendArchitecturePolicy.low_level_forbidden_bare_imports.includes(importPath))
        )
        .map((importPath) => `${file.relativePath} -> ${importPath}`)
    })

    expect(forbiddenImports).toEqual([])
  })

  it('feature UI files should not import non-contract API modules directly', () => {
    const featureUiFiles = sourceFiles.filter(
      (file) =>
        file.relativePath.startsWith(`features${sep}`) &&
        file.relativePath.includes(`${sep}ui${sep}`)
    )
    const apiImports = collectImportKeys(featureUiFiles, '@/api/').filter(
      (key) =>
        !key.includes(' -> @/api/contracts')
    )

    expect(apiImports).toEqual([])
  })

  it('new page components should be route views or widgets instead of legacy component pages', () => {
    const componentPageFiles = sourceFiles
      .map((file) => file.relativePath)
      .filter((relativePath) => relativePath.startsWith(`components${sep}`))
      .filter((relativePath) => /Page\.vue$/.test(relativePath))

    expect(componentPageFiles).toEqual([])
  })

  it('stores and utilities should not depend on UI or app orchestration layers', () => {
    const storeFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`stores${sep}`))
    const storeForbiddenImports = collectImportsMatching(storeFiles, (_file, importPath) =>
      frontendArchitecturePolicy.store_forbidden_import_prefixes.some((prefix) =>
        importPath.startsWith(prefix)
      )
    )

    const utilityFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`utils${sep}`))
    const utilityForbiddenImports = collectImportsMatching(
      utilityFiles,
      (_file, importPath) =>
        frontendArchitecturePolicy.utility_forbidden_import_prefixes.some((prefix) =>
          importPath.startsWith(prefix)
        ) || frontendArchitecturePolicy.utility_forbidden_bare_imports.includes(importPath)
    )

    expect(storeForbiddenImports).toEqual([])
    expect(utilityForbiddenImports).toEqual([])
  })

  it('legacy teacher frontend route paths should not remain in active frontend source', () => {
    const legacyTeacherFrontendRoutePattern =
      /['"`]\/?teacher\/(?:dashboard|classes|students|awd-reviews|instances)(?:[/'"`:?]|$)/

    const scannedFiles = sourceFiles.filter(
      (file) => !file.relativePath.startsWith(`api${sep}`)
    )

    const violations = collectSourceMatches(
      scannedFiles,
      (_file, source) => legacyTeacherFrontendRoutePattern.test(source)
    )

    expect(violations).toEqual([])
  })

  it('feature router access should stay in reviewed route-aware composables', () => {
    const featureFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`features${sep}`)
    )
    const routerImports = collectImportsMatching(
      featureFiles,
      (_file, importPath) =>
        frontendArchitecturePolicy.feature_router_forbidden_imports.some((forbiddenImport) =>
          forbiddenImport.endsWith('/') ? importPath.startsWith(forbiddenImport) : importPath === forbiddenImport
        ) || importPath.startsWith('@/router/')
    )

    expect(routerImports).toEqual([])
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

    expect(mixedComposables).toEqual([])
  })
})
