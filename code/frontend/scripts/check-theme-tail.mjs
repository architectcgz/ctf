import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

const roots = ['src/components', 'src/views', 'src/composables', 'src/utils']
const allowedRelativeFiles = new Set([
  'src/views/UILab.vue',
  'src/views/platform/ThemePreview.vue',
])
const allowedPathParts = new Set(['__tests__', 'refs'])
const sourceExtensions = new Set(['.vue', '.ts'])

const forbiddenThemePattern =
  /#(?:[0-9A-Fa-f]{8}|[0-9A-Fa-f]{6}|[0-9A-Fa-f]{3})(?![0-9A-Za-z_-])|rgba\(|text-slate-|text-emerald-|text-red-|text-orange-|text-blue-|text-cyan-|bg-white|border-white|text-white|color:\s*white|background:\s*white|%, black|bg-black\//g

function shouldSkip(path) {
  const normalized = path.split('/').join('/')
  if (allowedRelativeFiles.has(normalized)) return true
  return normalized.split('/').some((part) => allowedPathParts.has(part))
}

function extensionOf(path) {
  const index = path.lastIndexOf('.')
  return index === -1 ? '' : path.slice(index)
}

function collectFiles(root, files = []) {
  for (const entry of readdirSync(root)) {
    const path = join(root, entry)
    const normalized = relative(process.cwd(), path).split('/').join('/')
    if (shouldSkip(normalized)) continue

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

const violations = []

for (const root of roots) {
  for (const file of collectFiles(root)) {
    const source = readFileSync(file, 'utf8')
    const lines = source.split(/\r?\n/)

    lines.forEach((line, index) => {
      forbiddenThemePattern.lastIndex = 0
      const matches = [...line.matchAll(forbiddenThemePattern)]
      matches.forEach((match) => {
        violations.push({
          file: relative(process.cwd(), file),
          line: index + 1,
          token: match[0],
        })
      })
    })
  }
}

if (violations.length > 0) {
  console.error('Found hardcoded theme tail tokens:')
  violations.forEach(({ file, line, token }) => {
    console.error(`${file}:${line}: ${token}`)
  })
  process.exit(1)
}
