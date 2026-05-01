import { ref } from 'vue'
import type { Router } from 'vue-router'

import type {
  AdminChallengeImportPreview,
  AdminChallengePublishRequestData,
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
} from '@/api/contracts'
import {
  getChallengeCategoryColor,
  getChallengeCategoryLabel,
  getChallengeDifficultyColor,
  getChallengeDifficultyLabel,
} from '@/entities/challenge'
import type { PlatformChallengeListRow } from './usePlatformChallenges'

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

  const getCategoryLabel = getChallengeCategoryLabel

  function getCategoryColor(category: ChallengeCategory): string {
    return getChallengeCategoryColor(category, {
      web: 'var(--color-cat-web)',
      pwn: 'var(--color-cat-pwn)',
      reverse: 'var(--color-cat-reverse)',
      crypto: 'var(--color-cat-crypto)',
      misc: 'var(--color-cat-misc)',
      forensics: 'var(--color-cat-forensics)',
    })
  }

  const getDifficultyLabel = getChallengeDifficultyLabel

  function getDifficultyColor(difficulty: ChallengeDifficulty): string {
    return getChallengeDifficultyColor(difficulty, {
      beginner: 'var(--color-diff-beginner)',
      easy: 'var(--color-diff-easy)',
      medium: 'var(--color-diff-medium)',
      hard: 'var(--color-diff-hard)',
      insane: 'var(--color-diff-insane)',
    })
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
