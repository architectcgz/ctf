import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

const sourceRoots = ['src/components', 'src/views', 'src/widgets']

function collectVueFiles(dir: string): string[] {
  const entries = readdirSync(dir)
  const files: string[] = []

  for (const entry of entries) {
    const path = join(dir, entry)
    const stat = statSync(path)

    if (stat.isDirectory()) {
      if (entry === '__tests__') continue
      files.push(...collectVueFiles(path))
      continue
    }

    if (entry.endsWith('.vue')) {
      files.push(path)
    }
  }

  return files
}

describe('header button primitive boundary', () => {
  it('feature components should not override shared header-btn visual variables', () => {
    const offenders = sourceRoots.flatMap((root) =>
      collectVueFiles(join(process.cwd(), root))
        .map((file) => {
          const source = readFileSync(file, 'utf-8')
          return source.includes('--header-btn-') ? relative(process.cwd(), file) : null
        })
        .filter((file): file is string => Boolean(file))
    )

    expect(offenders).toEqual([])
  })
})
