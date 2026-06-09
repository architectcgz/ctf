import type { Ref } from 'vue'

import type { AdminContestAWDServiceData } from '@/api/contracts'
import {
  createHTTPStandardDraft,
  createLegacyProbeDraft,
  createScriptCheckerDraft,
  createTCPStandardDraft,
  getHTTPStandardPresetDraft,
  type AWDCheckerFieldErrorKey,
  type AWDHTTPStandardDraft,
  type AWDLegacyProbeDraft,
  type AWDScriptCheckerDraft,
  type AWDTCPStandardDraft,
} from './awdCheckerConfigSupport'

interface CheckerScoreDraft {
  sla_score: number
  defense_score: number
}

interface UseAwdCheckerDraftHydrationOptions {
  form: CheckerScoreDraft
  legacyProbeDraft: AWDLegacyProbeDraft
  httpStandardDraft: AWDHTTPStandardDraft
  tcpStandardDraft: AWDTCPStandardDraft
  scriptCheckerDraft: AWDScriptCheckerDraft
  fieldErrors: Record<AWDCheckerFieldErrorKey | 'checker_type' | 'sla_score' | 'defense_score', string>
  expandedTCPCheckerStepIndex: Ref<number | null>
  syncingDraft: Ref<boolean>
  clearErrors: () => void
}

export function useAwdCheckerDraftHydration({
  form,
  legacyProbeDraft,
  httpStandardDraft,
  tcpStandardDraft,
  scriptCheckerDraft,
  fieldErrors,
  expandedTCPCheckerStepIndex,
  syncingDraft,
  clearErrors,
}: UseAwdCheckerDraftHydrationOptions) {
  function assignHTTPDraft(next: AWDHTTPStandardDraft) {
    httpStandardDraft.put_flag = { ...next.put_flag }
    httpStandardDraft.get_flag = { ...next.get_flag }
    httpStandardDraft.havoc = { ...next.havoc }
  }

  function assignTCPDraft(next: AWDTCPStandardDraft) {
    tcpStandardDraft.timeout_ms = next.timeout_ms
    tcpStandardDraft.steps.splice(
      0,
      tcpStandardDraft.steps.length,
      ...next.steps.map((step) => ({ ...step }))
    )
    expandedTCPCheckerStepIndex.value = null
  }

  function assignScriptDraft(next: AWDScriptCheckerDraft) {
    scriptCheckerDraft.runtime = next.runtime
    scriptCheckerDraft.entry = next.entry
    scriptCheckerDraft.timeout_sec = next.timeout_sec
    scriptCheckerDraft.args_text = next.args_text
    scriptCheckerDraft.env_text = next.env_text
    scriptCheckerDraft.output = next.output
  }

  function hydrateServiceDraft(service: AdminContestAWDServiceData | null) {
    syncingDraft.value = true
    clearErrors()
    form.sla_score = service?.sla_score ?? 1
    form.defense_score = service?.defense_score ?? 2
    legacyProbeDraft.health_path = createLegacyProbeDraft(service?.checker_config).health_path
    assignHTTPDraft(createHTTPStandardDraft(service?.checker_config))
    assignTCPDraft(createTCPStandardDraft(service?.checker_config))
    assignScriptDraft(createScriptCheckerDraft(service?.checker_config))
    syncingDraft.value = false
  }

  function applyHTTPPreset(presetId: string) {
    assignHTTPDraft(getHTTPStandardPresetDraft(presetId))
  }

  function applyFieldErrors(
    errors: Partial<Record<AWDCheckerFieldErrorKey, string>>
  ) {
    for (const [key, value] of Object.entries(errors)) {
      fieldErrors[key as AWDCheckerFieldErrorKey] = value || ''
    }
  }

  return {
    hydrateServiceDraft,
    applyHTTPPreset,
    applyFieldErrors,
  }
}
