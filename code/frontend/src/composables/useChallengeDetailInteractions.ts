import { ref, type Ref } from 'vue'

import {
  downloadAttachment as downloadChallengeAttachment,
  getMyChallengeWriteupSubmission,
  submitFlag,
  upsertChallengeWriteupSubmission,
} from '@/api/challenge'
import type { ChallengeDetailData, SubmissionWriteupData } from '@/api/contracts'
import type { ChallengeSubmissionRecordStatus } from '@/composables/useChallengeDetailPresentation'
import { useToast } from '@/composables/useToast'

type EditableWriteupStatus = 'draft' | 'published'

interface SubmissionRecordItem {
  id: string
  answer: string
  status: ChallengeSubmissionRecordStatus
  message: string
  submittedAt?: string
}

interface UseChallengeDetailInteractionsOptions {
  challengeId: Ref<string>
  challenge: Ref<ChallengeDetailData | null>
  loadSolutions: (challengeId: string) => Promise<void>
}

export function useChallengeDetailInteractions({
  challengeId,
  challenge,
  loadSolutions,
}: UseChallengeDetailInteractionsOptions) {
  const toast = useToast()

  const myWriteup = ref<SubmissionWriteupData | null>(null)
  const submitting = ref(false)
  const submissionLoading = ref(false)
  const submissionSaving = ref<EditableWriteupStatus | null>(null)
  const writeupTitle = ref('')
  const writeupContent = ref('')
  const flagInput = ref('')
  const expandedHintLevels = ref<number[]>([])
  const submitResult = ref<{
    variant: 'success' | 'error' | 'pending'
    className: string
    message: string
  } | null>(null)
  const submissionRecords = ref<SubmissionRecordItem[]>([])

  function hydrateSubmissionForm(item: SubmissionWriteupData | null): void {
    writeupTitle.value = item?.title ?? ''
    writeupContent.value = item?.content ?? ''
  }

  function resetChallengeInteractions(): void {
    myWriteup.value = null
    submitting.value = false
    submissionLoading.value = false
    submissionSaving.value = null
    writeupTitle.value = ''
    writeupContent.value = ''
    flagInput.value = ''
    expandedHintLevels.value = []
    submitResult.value = null
    submissionRecords.value = []
  }

  async function loadMyWriteupSubmission(): Promise<void> {
    if (!challengeId.value) return

    submissionLoading.value = true
    try {
      myWriteup.value = await getMyChallengeWriteupSubmission(challengeId.value)
      hydrateSubmissionForm(myWriteup.value)
    } catch {
      toast.error('加载个人题解失败')
    } finally {
      submissionLoading.value = false
    }
  }

  function isHintExpanded(level: number): boolean {
    return expandedHintLevels.value.includes(level)
  }

  function toggleHint(level: number): void {
    if (isHintExpanded(level)) {
      expandedHintLevels.value = expandedHintLevels.value.filter((item) => item !== level)
      return
    }
    expandedHintLevels.value = [...expandedHintLevels.value, level]
  }

  async function submitFlagHandler(): Promise<void> {
    const currentChallenge = challenge.value
    if (!currentChallenge || !flagInput.value.trim()) return

    const answer = flagInput.value.trim()
    submitting.value = true
    submitResult.value = null
    try {
      const result = await submitFlag(currentChallenge.id, answer)
      submissionRecords.value = [
        {
          id: `${result.submitted_at}-${submissionRecords.value.length}`,
          answer,
          status: result.status,
          message: result.message,
          submittedAt: result.submitted_at,
        },
        ...submissionRecords.value,
      ]
      switch (result.status) {
        case 'correct':
          submitResult.value = {
            variant: 'success',
            className: 'text-[var(--color-success)]',
            message: result.message,
          }
          toast.success(result.message)
          currentChallenge.is_solved = true
          await loadSolutions(currentChallenge.id)
          break
        case 'pending_review':
          submitResult.value = {
            variant: 'pending',
            className: 'text-[var(--color-warning)]',
            message: result.message,
          }
          toast.info('答案已提交，等待教师审核')
          break
        default:
          submitResult.value = {
            variant: 'error',
            className: 'text-[var(--color-danger)]',
            message: result.message,
          }
          break
      }
    } catch {
      submissionRecords.value = [
        {
          id: `error-${Date.now()}`,
          answer,
          status: 'error',
          message: '提交失败，请重试',
          submittedAt: new Date().toISOString(),
        },
        ...submissionRecords.value,
      ]
      submitResult.value = {
        variant: 'error',
        className: 'text-[var(--color-danger)]',
        message: '提交失败，请重试',
      }
    } finally {
      submitting.value = false
    }
  }

  async function downloadAttachment(): Promise<void> {
    if (!challenge.value?.attachment_url) return

    const attachmentURL = challenge.value.attachment_url
    try {
      const parsed = new URL(attachmentURL, window.location.origin)
      if (parsed.origin !== window.location.origin) {
        window.open(attachmentURL, '_blank', 'noopener')
        return
      }
    } catch {
      // keep axios fallback for relative urls
    }

    try {
      const { blob, filename } = await downloadChallengeAttachment(attachmentURL)
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
    } catch {
      toast.error('下载附件失败')
    }
  }

  async function saveWriteup(status: EditableWriteupStatus): Promise<void> {
    if (!challenge.value) return
    if (!writeupTitle.value.trim() || !writeupContent.value.trim()) {
      toast.error('请先补全题解标题和正文')
      return
    }
    if (status === 'published' && !challenge.value.is_solved) {
      toast.error('解题后才能发布到社区')
      return
    }

    submissionSaving.value = status
    try {
      const saved = await upsertChallengeWriteupSubmission(challenge.value.id, {
        title: writeupTitle.value.trim(),
        content: writeupContent.value.trim(),
        submission_status: status,
      })
      myWriteup.value = saved
      hydrateSubmissionForm(saved)
      toast.success(status === 'published' ? '题解已发布到社区' : '草稿已保存')
    } catch {
      toast.error(status === 'published' ? '发布题解失败' : '保存草稿失败')
    } finally {
      submissionSaving.value = null
    }
  }

  return {
    myWriteup,
    submitting,
    submissionLoading,
    submissionSaving,
    writeupTitle,
    writeupContent,
    flagInput,
    expandedHintLevels,
    submitResult,
    submissionRecords,
    resetChallengeInteractions,
    loadMyWriteupSubmission,
    isHintExpanded,
    toggleHint,
    submitFlagHandler,
    downloadAttachment,
    saveWriteup,
  }
}
