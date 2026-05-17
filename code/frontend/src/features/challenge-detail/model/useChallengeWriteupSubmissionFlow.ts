import { ref, type Ref } from 'vue'

import {
  getMyChallengeWriteupSubmission,
  upsertChallengeWriteupSubmission,
} from '@/api/challenge'
import type { ChallengeDetailData, SubmissionWriteupData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

type EditableWriteupStatus = 'draft' | 'published'

interface UseChallengeWriteupSubmissionFlowOptions {
  challengeId: Ref<string>
  challenge: Ref<ChallengeDetailData | null>
}

export function useChallengeWriteupSubmissionFlow(options: UseChallengeWriteupSubmissionFlowOptions) {
  const { challengeId, challenge } = options
  const toast = useToast()

  const myWriteup = ref<SubmissionWriteupData | null>(null)
  const submissionLoading = ref(false)
  const submissionSaving = ref<EditableWriteupStatus | null>(null)
  const writeupTitle = ref('')
  const writeupContent = ref('')
  let latestWriteupRequestId = 0

  function hydrateSubmissionForm(item: SubmissionWriteupData | null): void {
    writeupTitle.value = item?.title ?? ''
    writeupContent.value = item?.content ?? ''
  }

  function resetWriteupSubmissionState(): void {
    myWriteup.value = null
    submissionLoading.value = false
    submissionSaving.value = null
    writeupTitle.value = ''
    writeupContent.value = ''
  }

  async function loadMyWriteupSubmission(): Promise<boolean> {
    const currentChallengeId = challengeId.value
    if (!currentChallengeId) return false

    const requestId = ++latestWriteupRequestId
    submissionLoading.value = true
    try {
      const nextWriteup = await getMyChallengeWriteupSubmission(currentChallengeId)
      if (requestId !== latestWriteupRequestId || currentChallengeId !== challengeId.value) {
        return false
      }
      myWriteup.value = nextWriteup
      hydrateSubmissionForm(myWriteup.value)
      return true
    } catch {
      if (requestId !== latestWriteupRequestId || currentChallengeId !== challengeId.value) {
        return false
      }
      toast.error('加载个人题解失败')
      return false
    } finally {
      if (requestId === latestWriteupRequestId && currentChallengeId === challengeId.value) {
        submissionLoading.value = false
      }
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
    submissionLoading,
    submissionSaving,
    writeupTitle,
    writeupContent,
    hydrateSubmissionForm,
    resetWriteupSubmissionState,
    loadMyWriteupSubmission,
    saveWriteup,
  }
}
