export function progressRate(total: number, solved: number): number {
  if (!total) return 0
  return Math.round((solved / total) * 100)
}

interface ProgressStat {
  total: number
  solved: number
}

interface CategoryProgressStat extends ProgressStat {
  category: string
}

interface DifficultyProgressStat extends ProgressStat {
  difficulty: string
}

export const canonicalDifficultyOrder = ['beginner', 'easy', 'medium', 'hard', 'insane'] as const

export function compareProgressPriority(left: ProgressStat, right: ProgressStat): number {
  const rateDiff = progressRate(left.total, left.solved) - progressRate(right.total, right.solved)
  if (rateDiff !== 0) return rateDiff

  const totalDiff = right.total - left.total
  if (totalDiff !== 0) return totalDiff

  return left.solved - right.solved
}

export function rankCategoryActionItems<T extends CategoryProgressStat>(
  categoryStats: T[]
): Array<T & { rate: number; remaining: number }> {
  return [...categoryStats]
    .map((item) => ({
      ...item,
      rate: progressRate(item.total, item.solved),
      remaining: Math.max(item.total - item.solved, 0),
    }))
    .sort(
      (left, right) =>
        compareProgressPriority(left, right) || left.category.localeCompare(right.category)
    )
}

function difficultyOrderIndex(difficulty: string): number {
  const index = canonicalDifficultyOrder.indexOf(
    difficulty as (typeof canonicalDifficultyOrder)[number]
  )
  return index === -1 ? canonicalDifficultyOrder.length : index
}

export function orderDifficultyActionItems<T extends DifficultyProgressStat>(
  difficultyStats: T[]
): Array<T & { rate: number; remaining: number; order: number }> {
  return canonicalDifficultyOrder
    .map((difficulty, order) => {
      const item = difficultyStats.find((stat) => stat.difficulty === difficulty)
      if (!item) return null

      return {
        ...item,
        rate: progressRate(item.total, item.solved),
        remaining: Math.max(item.total - item.solved, 0),
        order,
      }
    })
    .filter((item): item is T & { rate: number; remaining: number; order: number } => Boolean(item))
}

export function selectDifficultyPriority<T extends DifficultyProgressStat>(
  difficultyStats: T[]
): (T & { rate: number; remaining: number; order: number }) | null {
  const orderedItems = orderDifficultyActionItems(difficultyStats).filter(
    (item) => item.remaining > 0
  )
  if (orderedItems.length === 0) return null

  return [...orderedItems].sort((left, right) => {
    const rateDiff = left.rate - right.rate
    if (rateDiff !== 0) return rateDiff
    return difficultyOrderIndex(left.difficulty) - difficultyOrderIndex(right.difficulty)
  })[0]
}
