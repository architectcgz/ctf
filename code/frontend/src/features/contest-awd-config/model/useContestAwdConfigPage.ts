import { computed, onMounted, onUnmounted, watch } from 'vue'

import { useAwdCheckResultPresentation } from '@/features/awd-inspector'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useRouteNavigationTransport } from '@/composables/routeNavigationTransport'
import { useRouteQueryTransport } from '@/composables/routeQueryTransport'
import { contestAwdConfigBackToStudioRoute } from './contestAwdConfigRoutes'
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
  const { params, query, replaceQuery } = useRouteQueryTransport()
  const { push } = useRouteNavigationTransport()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const contestId = computed(() => String(params.value.id ?? ''))
  function readServiceQuery(): string {
    const value = query.value.service
    if (Array.isArray(value)) {
      return String(value[0] ?? '')
    }
    return typeof value === 'string' ? value : ''
  }

  function replaceServiceQuery(serviceId: string) {
    void replaceQuery({ ...query.value, service: serviceId })
  }
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
    reconcileSelectedServiceId,
    selectService,
  } = useAwdChallengeSelection({
    services,
    readServiceQuery,
    replaceServiceQuery,
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
    void push(contestAwdConfigBackToStudioRoute(contestId.value))
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
