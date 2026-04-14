import { request } from './request'

import type { PracticeRankingItemData } from './contracts'

interface RawPracticeRankingItemData extends Omit<PracticeRankingItemData, 'user_id'> {
  user_id: string | number
}

function normalizePracticeRankingItem(item: RawPracticeRankingItemData): PracticeRankingItemData {
  return {
    ...item,
    user_id: String(item.user_id),
  }
}

export async function getPracticeRanking(params?: {
  limit?: number
}): Promise<PracticeRankingItemData[]> {
  const response = await request<RawPracticeRankingItemData[]>({
    method: 'GET',
    url: '/scoreboard/ranking',
    params: {
      limit: params?.limit,
    },
  })

  return response.map(normalizePracticeRankingItem)
}
