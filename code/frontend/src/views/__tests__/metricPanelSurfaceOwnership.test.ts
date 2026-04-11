import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

function collectVueFiles(root: string): string[] {
  const entries = readdirSync(root, { withFileTypes: true })
  const files: string[] = []

  for (const entry of entries) {
    const fullPath = join(root, entry.name)

    if (entry.isDirectory()) {
      files.push(...collectVueFiles(fullPath))
      continue
    }

    if (entry.isFile() && entry.name.endsWith('.vue')) {
      files.push(fullPath)
    }
  }

  return files
}

describe('metric panel surface ownership', () => {
  it('metric-panel-default-surface 不应在页面局部被随意覆写背景变量', () => {
    const srcRoot = join(process.cwd(), 'src')
    const matches = collectVueFiles(srcRoot)
      .map((filePath) => ({
        filePath,
        source: readFileSync(filePath, 'utf8'),
      }))
      .filter(
        ({ source }) =>
          source.includes('metric-panel-default-surface') &&
          source.includes('--metric-panel-background')
      )
      .map(({ filePath }) => relative(srcRoot, filePath))

    expect(matches).toEqual(['components/dashboard/student/StudentTimelinePage.vue'])
  })

  it('student timeline 的局部桥接应保留共享卡片的高光加底面双层背景语义', () => {
    const timelineSource = readFileSync(
      join(process.cwd(), 'src/components/dashboard/student/StudentTimelinePage.vue'),
      'utf8'
    )

    expect(timelineSource).toMatch(
      /\.timeline-metric-grid\.metric-panel-default-surface\s*\{[\s\S]*--metric-panel-background:\s*radial-gradient\([\s\S]*linear-gradient\(/s
    )
  })
})
