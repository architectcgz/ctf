export function difficultyClass(difficulty: string): string {
  const map: Record<string, string> = {
    beginner: 'bg-green-100 text-green-700',
    easy: 'bg-blue-100 text-blue-700',
    medium: 'bg-yellow-100 text-yellow-700',
    hard: 'bg-orange-100 text-orange-700',
    insane: 'bg-red-100 text-red-700',
  }
  return map[difficulty] || 'bg-gray-100 text-gray-700'
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
