import { computed, onMounted, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import {
  buildChallengeImportManageRoute,
  buildChallengeManageRoute,
} from './challengeImportRoutes'
import { useChallengePackageImport } from './useChallengePackageImport'

export function useChallengeImportPreviewPage(importId: MaybeRefOrGetter<string>) {
  const currentImportId = computed(() => String(toValue(importId) ?? '').trim())
  const backToImportRoute = buildChallengeImportManageRoute()
  const backToQueueRoute = buildChallengeImportManageRoute('#challenge-queue-workspace')
  const commitSuccessRedirectRoute = ref<ReturnType<typeof buildChallengeManageRoute> | null>(null)

  const { preview, uploading, committing, hasPreview, loadPreview, resetPreview, commitPreview } =
    useChallengePackageImport({
      onCommitted: () => {
        commitSuccessRedirectRoute.value = buildChallengeManageRoute()
      },
    })

  async function syncPreviewByRoute(): Promise<void> {
    resetPreview()
    const id = currentImportId.value
    if (!id) {
      return
    }
    await loadPreview(id)
  }

  async function handleCommitPreview(): Promise<void> {
    await commitPreview()
  }

  onMounted(() => {
    void syncPreviewByRoute()
  })

  watch(currentImportId, () => {
    void syncPreviewByRoute()
  })

  return {
    preview,
    uploading,
    committing,
    hasPreview,
    backToImportRoute,
    backToQueueRoute,
    commitSuccessRedirectRoute,
    handleCommitPreview,
  }
}
