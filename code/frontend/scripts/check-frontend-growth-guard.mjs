import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const rootDir = process.cwd()
const policyPath = join(rootDir, 'scripts', 'frontend-architecture-policy.json')
const policy = JSON.parse(readFileSync(policyPath, 'utf8'))
const baselinePath = join(rootDir, policy.growth_baseline_file)

function readLineCount(path) {
  return readFileSync(join(rootDir, path), 'utf8').split(/\r?\n/).length
}

const baseline = JSON.parse(readFileSync(baselinePath, 'utf8'))
const violations = []

for (const [path, config] of Object.entries(baseline)) {
  const lineCount = readLineCount(path)
  const growthLimit = config.baseline_lines + config.max_growth

  if (lineCount > growthLimit) {
    violations.push(
      `${path}: ${lineCount} lines exceeds growth budget ${growthLimit} (baseline ${config.baseline_lines} + growth ${config.max_growth})`
    )
  }

  if (lineCount > config.max_lines) {
    violations.push(
      `${path}: ${lineCount} lines exceeds hard cap ${config.max_lines}`
    )
  }
}

if (violations.length > 0) {
  console.error('[frontend-growth] guard failed')
  for (const violation of violations) {
    console.error(`- ${violation}`)
  }
  process.exit(1)
}

console.log('[frontend-growth] guard passed')
