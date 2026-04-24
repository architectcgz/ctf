import { ref } from 'vue'
import type { Router } from 'vue-router'

import type {
  ChallengeCategory,
  ChallengeDifficulty,
} from '@/api/contracts'
import type { PlatformChallengeListRow } from '@/composables/usePlatformChallenges'

interface UseChallengeManagePresentationOptions {
  router: Router
  publish: (row: PlatformChallengeListRow) => Promise<void>
  remove: (challengeId: string) => Promise<void>
}

export function useChallengeManagePresentation({
  router,
  publish,
  remove,
}: UseChallengeManagePresentationOptions) {
  const openActionMenuId = ref<string | null>(null)

  function getCategoryLabel(category: ChallengeCategory): string {
    const labels: Record<ChallengeCategory, string> = {
      web: 'Web',
      pwn: 'Pwn',
      reverse: '逆向',
      crypto: '密码',
      misc: '杂项',
      forensics: '取证',
    }
    return labels[category]
  }

  function getDifficultyLabel(difficulty: ChallengeDifficulty): string {
    const labels: Record<ChallengeDifficulty, string> = {
      beginner: '入门',
      easy: '简单',
      medium: '中等',
      hard: '困难',
      insane: '地狱',
    }
    return labels[difficulty]
  }

  function toggleActionMenu(challengeId: string): void {
    openActionMenuId.value = openActionMenuId.value === challengeId ? null : challengeId
  }

  function closeActionMenu(): void {
    openActionMenuId.value = null
  }

  function openChallengeDetail(challengeId: string): void {
    closeActionMenu()
    void router.push(`/platform/challenges/${challengeId}`)
  }

  function openChallengeTopology(challengeId: string): void {
    closeActionMenu()
    void router.push(`/platform/challenges/${challengeId}/topology`)
  }

  function openChallengeWriteup(challengeId: string): void {
    closeActionMenu()
    void router.push({
      name: 'PlatformChallengeDetail',
      params: { id: challengeId },
      query: { panel: 'writeup' },
    })
  }

  async function submitPublishCheck(row: PlatformChallengeListRow): Promise<void> {
    closeActionMenu()
    await publish(row)
  }

  async function removeChallenge(challengeId: string): Promise<void> {
    closeActionMenu()
    await remove(challengeId)
  }

  return {
    openActionMenuId,
    getCategoryLabel,
    getDifficultyLabel,
    toggleActionMenu,
    closeActionMenu,
    openChallengeDetail,
    openChallengeTopology,
    openChallengeWriteup,
    submitPublishCheck,
    removeChallenge,
  }
}
