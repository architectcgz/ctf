import { computed, reactive, ref, type Ref } from 'vue'

import { runContestAWDCheckerPreview } from '@/api/admin/contests'
import type {
  AdminContestAWDServiceData,
  AWDCheckerPreviewData,
  AWDCheckerType,
} from '@/api/contracts'

interface RunAwdCheckerPreviewOptions {
  contestId: string
  serviceId?: number
  awdChallengeId: number
  checkerType: AWDCheckerType
  checkerConfig: Record<string, unknown>
  accessURL?: string
  previewFlag?: string
  previewRequestId?: string
}

export async function runAwdCheckerPreview(
  options: RunAwdCheckerPreviewOptions
): Promise<AWDCheckerPreviewData> {
  return runContestAWDCheckerPreview(options.contestId, {
    ...(options.serviceId && options.serviceId > 0 ? { service_id: options.serviceId } : {}),
    awd_challenge_id: options.awdChallengeId,
    checker_type: options.checkerType,
    checker_config: options.checkerConfig,
    ...(options.accessURL ? { access_url: options.accessURL } : {}),
    preview_flag: options.previewFlag?.trim() || undefined,
    ...(options.previewRequestId ? { preview_request_id: options.previewRequestId } : {}),
  })
}

interface UseAwdCheckerPreviewFlowOptions {
  contestId: Readonly<Ref<string>>
  selectedService: Readonly<Ref<AdminContestAWDServiceData | null>>
  selectedCheckerType: Readonly<Ref<AWDCheckerType | undefined>>
  currentSignature: Readonly<Ref<string>>
  syncingDraft: Readonly<Ref<boolean>>
  validateConfig: () => boolean
  buildCurrentCheckerConfig: () => Record<string, unknown>
}

export function useAwdCheckerPreviewFlow(options: UseAwdCheckerPreviewFlowOptions) {
  const {
    contestId,
    selectedService,
    selectedCheckerType,
    currentSignature,
    syncingDraft,
    validateConfig,
    buildCurrentCheckerConfig,
  } = options

  const previewing = ref(false)
  const previewResult = ref<AWDCheckerPreviewData | null>(null)
  const previewError = ref('')
  const previewToken = ref('')
  const previewSignature = ref('')
  const previewForm = reactive({
    access_url: '',
    preview_flag: 'flag{preview}',
  })

  const canAttachPreviewToken = computed(
    () => Boolean(previewToken.value) && previewSignature.value === currentSignature.value
  )

  function clearPreviewState() {
    previewResult.value = null
    previewError.value = ''
    previewToken.value = ''
    previewSignature.value = ''
    previewForm.access_url = ''
    previewForm.preview_flag = 'flag{preview}'
  }

  function handleSignatureChange(next: string, previous: string) {
    if (!syncingDraft.value && previous && next !== previewSignature.value) {
      previewToken.value = ''
    }
  }

  async function handlePreview() {
    if (
      previewing.value ||
      !selectedService.value ||
      !selectedCheckerType.value ||
      !validateConfig()
    ) {
      return
    }
    previewing.value = true
    previewError.value = ''
    previewResult.value = null
    previewToken.value = ''
    previewSignature.value = ''
    try {
      const result = await runAwdCheckerPreview({
        contestId: contestId.value,
        serviceId: readNumericID(selectedService.value.id),
        awdChallengeId: Number(selectedService.value.awd_challenge_id),
        checkerType: selectedCheckerType.value,
        checkerConfig: buildCurrentCheckerConfig(),
        accessURL: previewForm.access_url.trim() || undefined,
        previewFlag: previewForm.preview_flag.trim() || undefined,
      })
      previewResult.value = result
      previewToken.value = result.preview_token || ''
      previewSignature.value = currentSignature.value
    } catch (error) {
      previewError.value = error instanceof Error && error.message.trim() ? error.message : '试跑失败'
    } finally {
      previewing.value = false
    }
  }

  return {
    canAttachPreviewToken,
    clearPreviewState,
    handlePreview,
    handleSignatureChange,
    previewError,
    previewForm,
    previewResult,
    previewToken,
    previewing,
  }
}

function readNumericID(value: string): number | undefined {
  const next = Number(value)
  return Number.isFinite(next) && next > 0 ? next : undefined
}
