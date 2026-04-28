export function difficultyClass(difficulty: string): string {
  const map: Record<string, string> = {
    beginner: 'difficulty-chip difficulty-chip--beginner',
    easy: 'difficulty-chip difficulty-chip--easy',
    medium: 'difficulty-chip difficulty-chip--medium',
    hard: 'difficulty-chip difficulty-chip--hard',
    insane: 'difficulty-chip difficulty-chip--insane',
  }
  return map[difficulty] || 'difficulty-chip difficulty-chip--unknown'
}

export function difficultyLabel(difficulty: string): string {
  const map: Record<string, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    insane: '地狱',
  }
  return map[difficulty] || difficulty
}
