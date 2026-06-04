import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

const rootDir = process.cwd()
const sourceRoot = join(rootDir, 'src')
const allowlistPath = join(rootDir, 'scripts', 'vue-deep-allowlist.json')
const allowlist = JSON.parse(readFileSync(allowlistPath, 'utf8'))

const sourceExtensions = new Set(['.vue', '.css', '.scss', '.sass', '.less'])
const skippedPathParts = new Set(['__tests__', 'refs'])
const deepPattern = /:deep\((?:[^)(]+|\([^)(]*\))*\)/g
const legacyDeepPattern = /::v-deep|\/deep\//g

function normalizePath(path) {
  return path.split('\\').join('/')
}

function extensionOf(path) {
  const index = path.lastIndexOf('.')
  return index === -1 ? '' : path.slice(index)
}

function shouldSkip(path) {
  return normalizePath(path)
    .split('/')
    .some((part) => skippedPathParts.has(part))
}

function collectFiles(root, files = []) {
  for (const entry of readdirSync(root)) {
    const path = join(root, entry)
    const normalized = normalizePath(relative(rootDir, path))

    if (shouldSkip(normalized)) {
      continue
    }

    const stat = statSync(path)
    if (stat.isDirectory()) {
      collectFiles(path, files)
      continue
    }

    if (sourceExtensions.has(extensionOf(path))) {
      files.push(path)
    }
  }

  return files
}

function collectMatchCounts(source, pattern) {
  const matches = source.match(pattern) ?? []
  const counts = {}

  for (const match of matches) {
    counts[match] = (counts[match] ?? 0) + 1
  }

  return counts
}

const violations = []

for (const file of collectFiles(sourceRoot)) {
  const source = readFileSync(file, 'utf8')
  const relativePath = normalizePath(relative(rootDir, file))

  const legacyMatches = source.match(legacyDeepPattern) ?? []
  if (legacyMatches.length > 0) {
    violations.push({
      type: 'legacy-syntax',
      file: relativePath,
      matches: [...new Set(legacyMatches)].sort(),
    })
  }

  const actualCounts = collectMatchCounts(source, deepPattern)
  const actualEntries = Object.entries(actualCounts).sort(([left], [right]) => left.localeCompare(right))

  if (actualEntries.length === 0) {
    continue
  }

  const allowedCounts = allowlist[relativePath]
  if (!allowedCounts) {
    violations.push({
      type: 'unexpected-file',
      file: relativePath,
      entries: actualEntries,
    })
    continue
  }

  for (const [selector, count] of actualEntries) {
    const allowedCount = allowedCounts[selector] ?? 0
    if (allowedCount === 0) {
      violations.push({
        type: 'unexpected-selector',
        file: relativePath,
        selector,
        count,
      })
      continue
    }

    if (count > allowedCount) {
      violations.push({
        type: 'selector-count',
        file: relativePath,
        selector,
        count,
        allowedCount,
      })
    }
  }
}

if (violations.length > 0) {
  console.error('[vue-deep] guard failed')

  for (const violation of violations) {
    if (violation.type === 'legacy-syntax') {
      console.error(`- ${violation.file}: forbidden legacy deep syntax ${violation.matches.join(', ')}`)
      continue
    }

    if (violation.type === 'unexpected-file') {
      console.error(`- ${violation.file}: new file introduces :deep selectors`)
      for (const [selector, count] of violation.entries) {
        console.error(`  - ${selector} x${count}`)
      }
      continue
    }

    if (violation.type === 'unexpected-selector') {
      console.error(
        `- ${violation.file}: selector ${violation.selector} is not in the allowlist (found ${violation.count})`
      )
      continue
    }

    if (violation.type === 'selector-count') {
      console.error(
        `- ${violation.file}: selector ${violation.selector} exceeds allowlist count ${violation.allowedCount} (found ${violation.count})`
      )
    }
  }

  console.error(
    'Update existing components to remove :deep when possible. If a new unavoidable exception appears, review the owner boundary first and only then update scripts/vue-deep-allowlist.json.'
  )
  process.exit(1)
}

const trackedFiles = Object.keys(allowlist).length
const trackedSelectors = Object.values(allowlist).reduce(
  (sum, selectors) => sum + Object.values(selectors).reduce((innerSum, count) => innerSum + count, 0),
  0
)

console.log(`[vue-deep] guard passed (${trackedFiles} files, ${trackedSelectors} tracked deep selectors)`)
