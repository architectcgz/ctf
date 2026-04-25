import { ref } from 'vue'
import type { Router } from 'vue-router'

import type {
  AdminChallengeImportPreview,
  AdminChallengePublishRequestData,
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
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

  function getCategoryColor(category: ChallengeCategory): string {
    return {
      web: 'var(--color-cat-web)',
      pwn: 'var(--color-cat-pwn)',
      reverse: 'var(--color-cat-reverse)',
      crypto: 'var(--color-cat-crypto)',
      misc: 'var(--color-cat-misc)',
      forensics: 'var(--color-cat-forensics)',
    }[category]
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

  function getDifficultyColor(difficulty: ChallengeDifficulty): string {
    return {
      beginner: 'var(--color-diff-beginner)',
      easy: 'var(--color-diff-easy)',
      medium: 'var(--color-diff-medium)',
      hard: 'var(--color-diff-hard)',
      insane: 'var(--color-diff-insane)',
    }[difficulty]
  }

  function getStatusLabel(status: ChallengeStatus): string {
    return { draft: '草稿', published: '已发布', archived: '已归档' }[status]
  }

  function getStatusColor(status: ChallengeStatus): string {
    return {
      draft: 'var(--color-text-muted)',
      published: 'var(--color-success)',
      archived: 'var(--color-text-secondary)',
    }[status]
  }

  function getPublishRequestLabel(request: AdminChallengePublishRequestData | null): string {
    if (!request) return '未提交检查'

    return {
      queued: '等待检查',
      running: '检查中',
      succeeded: '检查通过',
      failed: '检查失败',
    }[request.status]
  }

  function getPublishRequestColor(request: AdminChallengePublishRequestData | null): string {
    if (!request) return 'var(--color-text-muted)'

    return {
      queued: 'var(--color-warning)',
      running: 'var(--color-primary)',
      succeeded: 'var(--color-success)',
      failed: 'var(--color-danger)',
    }[request.status]
  }

  function formatDateTime(value: string): string {
    return new Date(value).toLocaleString('zh-CN')
  }

  async function inspectImportTask(item: AdminChallengeImportPreview): Promise<void> {
    await router.push({
      name: 'PlatformChallengeImportPreview',
      params: { importId: item.id },
    })
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
    getCategoryColor,
    getDifficultyLabel,
    getDifficultyColor,
    getStatusLabel,
    getStatusColor,
    getPublishRequestLabel,
    getPublishRequestColor,
    formatDateTime,
    inspectImportTask,
    toggleActionMenu,
    closeActionMenu,
    openChallengeDetail,
    openChallengeTopology,
    openChallengeWriteup,
    submitPublishCheck,
    removeChallenge,
  }
}
