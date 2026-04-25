import { ref, type Ref } from 'vue'

import {
  downloadAttachment as downloadChallengeAttachment,
  getMyChallengeSubmissionRecords,
  getMyChallengeWriteupSubmission,
  submitFlag,
  upsertChallengeWriteupSubmission,
} from '@/api/challenge'
import type { ChallengeDetailData, SubmissionWriteupData, SubmitFlagData } from '@/api/contracts'
import type { ChallengeSubmissionRecordStatus } from '@/composables/useChallengeDetailPresentation'
import { useToast } from '@/composables/useToast'

type EditableWriteupStatus = 'draft' | 'published'

interface SubmissionRecordItem {
  id: string
  answer?: string
  status: ChallengeSubmissionRecordStatus
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
    message: string
  } | null>(null)
  const submissionRecords = ref<SubmissionRecordItem[]>([])
  let latestWriteupRequestId = 0
  let latestSubmissionRecordsRequestId = 0

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
    const currentChallengeId = challengeId.value
    if (!currentChallengeId) return

    const requestId = ++latestWriteupRequestId
    submissionLoading.value = true
    try {
      const nextWriteup = await getMyChallengeWriteupSubmission(currentChallengeId)
      if (requestId !== latestWriteupRequestId || currentChallengeId !== challengeId.value) {
        return
      }
      myWriteup.value = nextWriteup
      hydrateSubmissionForm(myWriteup.value)
    } catch {
      if (requestId !== latestWriteupRequestId || currentChallengeId !== challengeId.value) {
        return
      }
      toast.error('加载个人题解失败')
    } finally {
      if (requestId === latestWriteupRequestId && currentChallengeId === challengeId.value) {
        submissionLoading.value = false
      }
    }
  }

  async function loadSubmissionRecords(): Promise<void> {
    const currentChallengeId = challengeId.value
    if (!currentChallengeId) return

    const requestId = ++latestSubmissionRecordsRequestId
    try {
      const records = await getMyChallengeSubmissionRecords(currentChallengeId)
      if (
        requestId !== latestSubmissionRecordsRequestId ||
        currentChallengeId !== challengeId.value
      ) {
        return
      }
      submissionRecords.value = records.map((item) => ({
        id: item.id,
        answer: item.answer,
        status: item.status,
        submittedAt: item.submitted_at,
      }))
    } catch {
      if (
        requestId !== latestSubmissionRecordsRequestId ||
        currentChallengeId !== challengeId.value
      ) {
        return
      }
      toast.error('加载提交记录失败')
    }
  }

  function formatShutdownCountdown(result: SubmitFlagData): string | null {
    if (!result.instance_shutdown_at) return null

    const submittedAt = new Date(result.submitted_at).getTime()
    const shutdownAt = new Date(result.instance_shutdown_at).getTime()
    if (Number.isNaN(submittedAt) || Number.isNaN(shutdownAt) || shutdownAt <= submittedAt) {
      return null
    }

    const deltaMs = shutdownAt - submittedAt
    const totalMinutes = Math.round(deltaMs / 60000)
    if (totalMinutes >= 1) {
      return `${totalMinutes} 分钟`
    }

    const totalSeconds = Math.max(1, Math.round(deltaMs / 1000))
    return `${totalSeconds} 秒`
  }

  function buildSubmitResultMessage(result: SubmitFlagData): string {
    const repeatedCorrect =
      result.status === 'correct' && (result.points ?? 0) <= 0 && challenge.value?.is_solved

    if (repeatedCorrect) {
      return 'Flag 校验通过，本题已解出，不重复计分'
    }
    if (result.status === 'correct') {
      const countdown = formatShutdownCountdown(result)
      if (countdown) {
        return `恭喜你，Flag 正确！当前实例将在 ${countdown}后自动关闭`
      }
      return '恭喜你，Flag 正确！'
    }
    if (result.status === 'pending_review') {
      return '答案已提交，等待教师审核'
    }
    return 'Flag 错误，请重试'
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
    if (submitting.value || !currentChallenge || !flagInput.value.trim()) return

    const answer = flagInput.value.trim()
    const alreadySolved = currentChallenge.is_solved
    submitting.value = true
    submitResult.value = null
    try {
      const result = await submitFlag(currentChallenge.id, answer)
      const submitMessage = buildSubmitResultMessage(result)
      submissionRecords.value = [
        {
          id: `${result.submitted_at}-${submissionRecords.value.length}`,
          answer,
          status: result.status,
          submittedAt: result.submitted_at,
        },
        ...submissionRecords.value,
      ]
      switch (result.status) {
        case 'correct':
          submitResult.value = {
            variant: 'success',
            message: submitMessage,
          }
          toast.success(submitMessage)
          currentChallenge.is_solved = true
          if (!alreadySolved) {
            await loadSolutions(currentChallenge.id)
          }
          break
        case 'pending_review':
          submitResult.value = {
            variant: 'pending',
            message: submitMessage,
          }
          toast.info(submitMessage)
          break
        default:
          submitResult.value = {
            variant: 'error',
            message: submitMessage,
          }
          break
      }
    } catch {
      submissionRecords.value = [
        {
          id: `error-${Date.now()}`,
          answer,
          status: 'error',
          submittedAt: new Date().toISOString(),
        },
        ...submissionRecords.value,
      ]
      submitResult.value = {
        variant: 'error',
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
    loadSubmissionRecords,
    isHintExpanded,
    toggleHint,
    submitFlagHandler,
    downloadAttachment,
    saveWriteup,
  }
}
