import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAwdCheckResultPresentation } from '@/features/awd-inspector'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useAwdChallengeSelection } from './useAwdChallengeSelection'
import { useAwdCheckerConfigDraft } from './useAwdCheckerConfigDraft'
import {
  formatAwdCheckDateTime,
  getAwdCheckerTypeLabel,
  getAwdProtocolLabel,
  getAwdValidationLabel,
} from './awdCheckerLabels'
import { useAwdCheckerPreviewFlow } from './useAwdCheckerPreview'
import { useAwdCheckerSaveFlow } from './useAwdCheckerSaveFlow'
import { useContestAwdConfigDataLoader } from './useContestAwdConfigDataLoader'

export function useContestAwdConfigPage() {
  const route = useRoute()
  const router = useRouter()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const contestId = computed(() => String(route.params.id ?? ''))
  const {
    clearBreadcrumbDetailTitle,
    contest,
    loadError,
    loading,
    loadPage,
    refreshing,
    services,
    setAfterLoadHandler,
  } = useContestAwdConfigDataLoader({
    contestId,
    setBreadcrumbDetailTitle,
  })

  const {
    selectedServiceId,
    selectedService,
    selectedCheckerType,
    sortedServices,
    readServiceQuery,
    reconcileSelectedServiceId,
    selectService,
  } = useAwdChallengeSelection({
    contestId,
    route,
    router,
    services,
  })
  setAfterLoadHandler(reconcileSelectedServiceId)
  const {
    AWD_HTTP_METHOD_OPTIONS,
    AWD_HTTP_STANDARD_PRESETS,
    checkerConfigJSON,
    currentSignature,
    expandedTCPCheckerStepIndex,
    fieldErrors,
    form,
    httpActionSections,
    httpStandardDraft,
    legacyProbeDraft,
    scriptCheckerDraft,
    syncingDraft,
    tcpStandardDraft,
    addTCPCheckerStep,
    applyHTTPPreset,
    buildCurrentCheckerConfig,
    hydrateServiceDraft,
    removeTCPCheckerStep,
    summarizeTCPCheckerStep,
    toggleTCPCheckerStep,
    validateConfig,
  } = useAwdCheckerConfigDraft({
    selectedService,
    selectedCheckerType,
  })

  const {
    canAttachPreviewToken,
    clearPreviewState,
    handlePreview,
    handleSignatureChange,
    previewError,
    previewForm,
    previewResult,
    previewToken,
    previewing,
  } = useAwdCheckerPreviewFlow({
    contestId,
    selectedService,
    selectedCheckerType,
    currentSignature,
    syncingDraft,
    validateConfig,
    buildCurrentCheckerConfig,
  })
  const { handleSave, saving } = useAwdCheckerSaveFlow({
    contestId,
    selectedService,
    selectedCheckerType,
    canAttachPreviewToken,
    previewToken,
    form,
    validateConfig,
    buildCurrentCheckerConfig,
    reloadPage: () => loadPage(false),
  })

  const { summarizeCheckResult, getCheckStatusLabel, getPrimaryAccessURL } =
    useAwdCheckResultPresentation({
      formatDateTime: formatAwdCheckDateTime,
    })
  const previewSummary = computed(() =>
    previewResult.value
      ? summarizeCheckResult({
          ...previewResult.value.check_result,
          preview_context: previewResult.value.preview_context,
        })
      : ''
  )
  const previewAccessURL = computed(() =>
    previewResult.value
      ? getPrimaryAccessURL({
          ...previewResult.value.check_result,
          preview_context: previewResult.value.preview_context,
        })
      : ''
  )

  function goBackToStudio() {
    void router.push({
      name: 'ContestEdit',
      params: { id: contestId.value },
      query: { panel: 'awd-config' },
    })
  }

  watch(selectedService, (service) => {
    clearPreviewState()
    hydrateServiceDraft(service)
  })

  watch(currentSignature, (next, previous) => {
    handleSignatureChange(next, previous)
  })

  onMounted(() => {
    selectedServiceId.value = readServiceQuery()
    void loadPage(true)
  })

  onUnmounted(() => {
    clearBreadcrumbDetailTitle()
  })

  return {
    AWD_HTTP_METHOD_OPTIONS,
    AWD_HTTP_STANDARD_PRESETS,
    canAttachPreviewToken,
    checkerConfigJSON,
    contest,
    expandedTCPCheckerStepIndex,
    fieldErrors,
    form,
    getCheckStatusLabel,
    getCheckerTypeLabel: getAwdCheckerTypeLabel,
    getProtocolLabel: getAwdProtocolLabel,
    getValidationLabel: getAwdValidationLabel,
    goBackToStudio,
    handlePreview,
    handleSave,
    httpActionSections,
    httpStandardDraft,
    legacyProbeDraft,
    loadError,
    loading,
    loadPage,
    previewAccessURL,
    previewError,
    previewForm,
    previewResult,
    previewSummary,
    previewing,
    refreshing,
    saving,
    scriptCheckerDraft,
    selectService,
    selectedCheckerType,
    selectedService,
    selectedServiceId,
    sortedServices,
    tcpStandardDraft,
    addTCPCheckerStep,
    applyHTTPPreset,
    removeTCPCheckerStep,
    summarizeTCPCheckerStep,
    toggleTCPCheckerStep,
  }
}
