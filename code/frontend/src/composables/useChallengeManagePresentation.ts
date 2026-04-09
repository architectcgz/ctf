import { ref } from 'vue'
import type { Router } from 'vue-router'

import type {
  AdminChallengeImportPreview,
  AdminChallengePublishRequestData,
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
} from '@/api/contracts'
import type { AdminChallengeListRow } from '@/composables/useAdminChallenges'

interface UseChallengeManagePresentationOptions {
  router: Router
  publish: (row: AdminChallengeListRow) => Promise<void>
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
      web: '#2563eb',
      pwn: '#dc2626',
      reverse: '#7c3aed',
      crypto: '#d97706',
      misc: '#0f766e',
      forensics: '#0891b2',
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
      beginner: '#16a34a',
      easy: '#2563eb',
      medium: '#d97706',
      hard: '#dc2626',
      insane: '#6d28d9',
    }[difficulty]
  }

  function getStatusLabel(status: ChallengeStatus): string {
    return { draft: '草稿', published: '已发布', archived: '已归档' }[status]
  }

  function getStatusColor(status: ChallengeStatus): string {
    return { draft: '#64748b', published: '#059669', archived: '#6b7280' }[status]
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
    if (!request) return '#64748b'

    return {
      queued: '#d97706',
      running: '#2563eb',
      succeeded: '#059669',
      failed: '#dc2626',
    }[request.status]
  }

  function formatDateTime(value: string): string {
    return new Date(value).toLocaleString('zh-CN')
  }

  async function inspectImportTask(item: AdminChallengeImportPreview): Promise<void> {
    await router.push({
      name: 'AdminChallengeImportPreview',
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
    void router.push(`/platform/challenges/${challengeId}/writeup`)
  }

  async function submitPublishCheck(row: AdminChallengeListRow): Promise<void> {
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
