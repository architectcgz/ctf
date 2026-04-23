import { computed, ref, watch, type Ref } from 'vue'
import { marked } from 'marked'

import type {
  ChallengeCategory,
  ChallengeDetailData,
  ChallengeDifficulty,
  CommunityChallengeSolutionData,
  RecommendedChallengeSolutionData,
  SubmissionWriteupData,
  SubmissionWriteupStatus,
  SubmissionWriteupVisibilityStatus,
} from '@/api/contracts'

export type ChallengeSolutionTab = 'recommended' | 'community'
export type ChallengeSubmissionRecordStatus = 'correct' | 'incorrect' | 'pending_review' | 'error'

interface SubmitResultState {
  variant: 'success' | 'error' | 'pending'
  className: string
  message: string
}

export interface ChallengeSolutionCard {
  id: string
  title: string
  content: string
  preview: string
  authorName: string
  sourceLabel: string
  badge: string
  badgeClass: string
  updatedAt?: string
}

interface UseChallengeDetailPresentationOptions {
  challenge: Ref<ChallengeDetailData | null>
  recommendedSolutions: Ref<RecommendedChallengeSolutionData[]>
  communitySolutions: Ref<CommunityChallengeSolutionData[]>
  myWriteup: Ref<SubmissionWriteupData | null>
  selectedSolutionId: Ref<string | null>
  submitResult: Ref<SubmitResultState | null>
  sanitizeHtml: (source: string) => string
}

export function useChallengeDetailPresentation({
  challenge,
  recommendedSolutions,
  communitySolutions,
  myWriteup,
  selectedSolutionId,
  submitResult,
  sanitizeHtml,
}: UseChallengeDetailPresentationOptions) {
  const activeSolutionTab = ref<ChallengeSolutionTab>('recommended')

  function renderRichContent(source?: string): string {
    if (!source) return ''
    const html = marked.parse(source, {
      gfm: true,
      breaks: true,
    })
    return sanitizeHtml(typeof html === 'string' ? html : source)
  }

  function buildPreview(source?: string): string {
    if (!source) return ''
    return source
      .replace(/<[^>]+>/g, ' ')
      .replace(/\s+/g, ' ')
      .trim()
      .slice(0, 120)
  }

  const sanitizedDescription = computed(() => renderRichContent(challenge.value?.description))

  const recommendedSolutionCards = computed<ChallengeSolutionCard[]>(() =>
    recommendedSolutions.value.map((item) => ({
      id: item.id,
      title: item.title,
      content: item.content,
      preview: buildPreview(item.content),
      authorName: item.author_name,
      sourceLabel: item.source_type === 'official' ? '官方题解' : '社区推荐',
      badge: '推荐题解',
      badgeClass: 'writeup-status-pill--primary',
      updatedAt: item.updated_at,
    }))
  )

  const communitySolutionCards = computed<ChallengeSolutionCard[]>(() =>
    communitySolutions.value.map((item) => ({
      id: item.id,
      title: item.title,
      content: item.content,
      preview: item.content_preview || buildPreview(item.content),
      authorName: item.author_name,
      sourceLabel: '社区题解',
      badge: item.is_recommended ? '推荐' : '',
      badgeClass: item.is_recommended
        ? 'writeup-status-pill--primary'
        : 'writeup-status-pill--muted',
      updatedAt: item.updated_at,
    }))
  )

  const displayedSolutionCards = computed(() =>
    activeSolutionTab.value === 'recommended'
      ? recommendedSolutionCards.value
      : communitySolutionCards.value
  )

  const activeSolution = computed(() => {
    if (displayedSolutionCards.value.length === 0) return null
    return (
      displayedSolutionCards.value.find((item) => item.id === selectedSolutionId.value) ??
      displayedSolutionCards.value[0]
    )
  })

  const sanitizedActiveSolutionContent = computed(() =>
    renderRichContent(activeSolution.value?.content)
  )

  const submitPlaceholder = computed(() => {
    if (challenge.value?.is_solved) return '本题已通过，可继续输入 Flag 做校验'

    switch (submitResult.value?.variant) {
      case 'success':
        return '答案已通过'
      case 'pending':
        return '已提交，等待教师审核'
      case 'error':
        return '答案不正确，请继续尝试'
      default:
        return 'flag{...}'
    }
  })

  const submitPanelTitle = computed(() => 'Flag 提交')

  const submitPanelCopy = computed(() =>
    challenge.value?.is_solved
      ? '本题已解出，仍可继续提交 Flag 做校验；系统不会重复计分。'
      : '输入当前题目的 Flag 并提交验证。'
  )

  const submitFieldLabel = computed(() => 'Flag')

  const submitInputClass = computed(() => {
    switch (submitResult.value?.variant) {
      case 'success':
        return 'border-[var(--color-success)] bg-[var(--color-success)]/5'
      case 'pending':
        return 'border-[var(--color-warning)] bg-[var(--color-warning)]/8'
      case 'error':
        return 'border-[var(--color-danger)] bg-[var(--color-danger)]/5'
      default:
        return 'border-[var(--journal-accent)]'
    }
  })

  function clearSolutions(): void {
    recommendedSolutions.value = []
    communitySolutions.value = []
    activeSolutionTab.value = 'recommended'
    selectedSolutionId.value = null
  }

  function buildMetaPillStyle(color: string): Record<string, string> {
    return {
      borderColor: `color-mix(in srgb, ${color} 18%, transparent)`,
      backgroundColor: `color-mix(in srgb, ${color} 12%, transparent)`,
      color,
    }
  }

  function submissionStatusLabel(status?: SubmissionWriteupStatus): string {
    if (status === 'draft') return '草稿'
    if (status === 'published' || status === 'submitted') return '已发布'
    return '未开始'
  }

  function submissionStatusText(status: ChallengeSubmissionRecordStatus): string {
    if (status === 'correct') return '正确'
    if (status === 'incorrect') return '错误答案'
    if (status === 'pending_review') return '待审核'
    if (status === 'error') return '提交失败'
    return '未知'
  }

  function submissionRecordMessage(status: ChallengeSubmissionRecordStatus): string {
    if (status === 'correct') return '恭喜你，Flag 正确！'
    if (status === 'incorrect') return 'Flag 错误，请重试'
    if (status === 'pending_review') return '答案已提交，等待教师审核'
    if (status === 'error') return '提交失败，请重试'
    return '提交状态未知'
  }

  function visibilityStatusLabel(status?: SubmissionWriteupVisibilityStatus): string {
    if (status === 'hidden') return '已隐藏'
    if (
      myWriteup.value?.submission_status === 'published' ||
      myWriteup.value?.submission_status === 'submitted'
    ) {
      return '已公开'
    }
    return '未发布'
  }

  function formatWriteupTime(value?: string): string {
    if (!value) return '-'
    return new Date(value).toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  function formatSubmissionTime(value?: string): string {
    if (!value) return '-'
    return new Date(value).toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    })
  }

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
    const colors: Record<ChallengeCategory, string> = {
      web: 'var(--challenge-tone-web)',
      pwn: 'var(--challenge-tone-pwn)',
      reverse: 'var(--challenge-tone-reverse)',
      crypto: 'var(--challenge-tone-crypto)',
      misc: 'var(--challenge-tone-misc)',
      forensics: 'var(--challenge-tone-forensics)',
    }
    return colors[category]
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
    const colors: Record<ChallengeDifficulty, string> = {
      beginner: 'var(--challenge-tone-beginner)',
      easy: 'var(--challenge-tone-easy)',
      medium: 'var(--challenge-tone-medium)',
      hard: 'var(--challenge-tone-hard)',
      insane: 'var(--challenge-tone-insane)',
    }
    return colors[difficulty]
  }

  watch(
    displayedSolutionCards,
    (items) => {
      if (!items.some((item) => item.id === selectedSolutionId.value)) {
        selectedSolutionId.value = items[0]?.id ?? null
      }
    },
    { immediate: true }
  )

  watch(
    [recommendedSolutionCards, communitySolutionCards],
    ([recommended, community]) => {
      if (
        activeSolutionTab.value === 'recommended' &&
        recommended.length === 0 &&
        community.length > 0
      ) {
        activeSolutionTab.value = 'community'
      } else if (
        activeSolutionTab.value === 'community' &&
        community.length === 0 &&
        recommended.length > 0
      ) {
        activeSolutionTab.value = 'recommended'
      }
    },
    { immediate: true }
  )

  return {
    activeSolutionTab,
    sanitizedDescription,
    displayedSolutionCards,
    activeSolution,
    sanitizedActiveSolutionContent,
    submitPlaceholder,
    submitPanelTitle,
    submitPanelCopy,
    submitFieldLabel,
    submitInputClass,
    clearSolutions,
    buildMetaPillStyle,
    submissionStatusLabel,
    submissionStatusText,
    submissionRecordMessage,
    visibilityStatusLabel,
    formatWriteupTime,
    formatSubmissionTime,
    getCategoryLabel,
    getCategoryColor,
    getDifficultyLabel,
    getDifficultyColor,
  }
}
