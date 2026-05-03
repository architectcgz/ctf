import { type ComputedRef, type Ref } from 'vue'

import {
  createContestAWDService,
  createAdminContestChallenge,
  deleteContestAWDService,
  deleteAdminContestChallenge,
  updateContestAWDService,
  updateAdminContestChallenge,
} from '@/api/admin/contests'
import type {
  AdminAwdChallengeData,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

interface ContestOrchestrationSavePayload {
  challenge_id?: number
  awd_challenge_id?: number
  awd_challenge_ids?: number[]
  points: number
  order: number
  is_visible: boolean
}

interface UseContestChallengeMutationsOptions {
  contestId: Readonly<Ref<string>>
  contestMode: Readonly<Ref<ContestDetailData['mode']>>
  usingExternalChallengeLinks: ComputedRef<boolean>
  isAwdContest: ComputedRef<boolean>
  dialogMode: Ref<'create' | 'edit'>
  editingChallenge: Ref<AdminContestChallengeViewData | null>
  awdChallengeCatalog: Ref<AdminAwdChallengeData[]>
  saving: Ref<boolean>
  removingChallengeId: Ref<string | null>
  onUpdated: () => void
  refresh: () => Promise<void>
  closeDialog: () => void
  humanizeRequestError: (error: unknown, fallback: string) => string
  getChallengeTitle: (item: AdminContestChallengeViewData) => string
}

export function useContestChallengeMutations(options: UseContestChallengeMutationsOptions) {
  const {
    contestId,
    contestMode,
    usingExternalChallengeLinks,
    isAwdContest,
    dialogMode,
    editingChallenge,
    awdChallengeCatalog,
    saving,
    removingChallengeId,
    onUpdated,
    refresh,
    closeDialog,
    humanizeRequestError,
    getChallengeTitle,
  } = options
  const toast = useToast()

  function summarizeAwdChallengeFailures(awdChallengeIds: number[]): string {
    const failedNames = awdChallengeIds.map(
      (awdChallengeId) =>
        awdChallengeCatalog.value.find((item) => Number(item.id) === awdChallengeId)?.name ||
        `AWD #${awdChallengeId}`
    )
    return `部分 AWD 题目关联失败：${failedNames.join('、')}`
  }

  function buildAwdServiceCreatePayload(
    awdChallengeId: number,
    payload: ContestOrchestrationSavePayload,
    order: number
  ) {
    const awdChallenge = awdChallengeCatalog.value.find(
      (item) => Number(item.id) === awdChallengeId
    )
    const checkerConfig =
      awdChallenge?.checker_config && typeof awdChallenge.checker_config === 'object'
        ? awdChallenge.checker_config
        : undefined

    return {
      awd_challenge_id: awdChallengeId,
      points: payload.points,
      order,
      is_visible: payload.is_visible,
      ...(awdChallenge?.checker_type ? { checker_type: awdChallenge.checker_type } : {}),
      ...(checkerConfig ? { checker_config: checkerConfig } : {}),
    }
  }

  async function handleSave(payload: ContestOrchestrationSavePayload) {
    saving.value = true
    try {
      if (isAwdContest.value) {
        const awdChallengeIds =
          dialogMode.value === 'create' && payload.awd_challenge_ids?.length
            ? payload.awd_challenge_ids
            : payload.awd_challenge_id
              ? [payload.awd_challenge_id]
              : []

        if (awdChallengeIds.length === 0) {
          toast.error('请选择 AWD 题目')
          return
        }

        if (dialogMode.value === 'create') {
          const results = await Promise.allSettled(
            awdChallengeIds.map((awdChallengeId, index) =>
              createContestAWDService(
                contestId.value,
                buildAwdServiceCreatePayload(awdChallengeId, payload, payload.order + index)
              )
            )
          )
          const failedResults = results.flatMap((result, index) =>
            result.status === 'rejected'
              ? [{ awdChallengeId: awdChallengeIds[index], error: result.reason }]
              : []
          )

          if (failedResults.length > 0) {
            const failedIds = failedResults.map(({ awdChallengeId }) => awdChallengeId)
            const failureMessage = summarizeAwdChallengeFailures(failedIds)

            if (failedResults.length === awdChallengeIds.length) {
              toast.error(failureMessage)
              return
            }

            toast.warning(failureMessage)
            onUpdated()
            if (!usingExternalChallengeLinks.value) {
              await refresh()
            }
            return
          }
        } else if (editingChallenge.value) {
          await updateContestAWDService(contestId.value, editingChallenge.value.awd_service_id!, {
            awd_challenge_id: awdChallengeIds[0],
            points: payload.points,
            order: payload.order,
            is_visible: payload.is_visible,
          })
        }
      } else if (dialogMode.value === 'create') {
        await createAdminContestChallenge(contestId.value, {
          challenge_id: payload.challenge_id!,
          points: payload.points,
          order: payload.order,
          is_visible: payload.is_visible,
        })
      } else if (editingChallenge.value) {
        await updateAdminContestChallenge(contestId.value, editingChallenge.value.challenge_id, {
          points: payload.points,
          order: payload.order,
          is_visible: payload.is_visible,
        })
      }

      toast.success('题目已保存')
      closeDialog()
      onUpdated()
      if (!usingExternalChallengeLinks.value) await refresh()
    } catch (error) {
      toast.error(humanizeRequestError(error, '保存失败'))
    } finally {
      saving.value = false
    }
  }

  async function handleRemove(challenge: AdminContestChallengeViewData) {
    const confirmed = await confirmDestructiveAction({
      title: '移除题目',
      message: `确认将“${getChallengeTitle(challenge)}”从竞赛中移除吗？`,
    })
    if (!confirmed) return

    removingChallengeId.value = challenge.id
    try {
      if (contestMode.value === 'awd') {
        await deleteContestAWDService(contestId.value, challenge.awd_service_id!)
      } else {
        await deleteAdminContestChallenge(contestId.value, challenge.challenge_id)
      }
      toast.success('题目已移除')
      onUpdated()
      if (!usingExternalChallengeLinks.value) await refresh()
    } catch (error) {
      toast.error(humanizeRequestError(error, '移除失败'))
    } finally {
      removingChallengeId.value = null
    }
  }

  return {
    handleSave,
    handleRemove,
  }
}
