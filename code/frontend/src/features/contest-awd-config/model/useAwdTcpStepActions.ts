import type { Ref } from 'vue'

import type { AWDTCPStandardDraft } from './awdCheckerConfigSupport'

interface UseAwdTcpStepActionsOptions {
  tcpStandardDraft: AWDTCPStandardDraft
  expandedTCPCheckerStepIndex: Ref<number | null>
}

export function useAwdTcpStepActions({
  tcpStandardDraft,
  expandedTCPCheckerStepIndex,
}: UseAwdTcpStepActionsOptions) {
  function addTCPCheckerStep() {
    tcpStandardDraft.steps.push({
      send: '',
      send_template: '',
      send_hex: '',
      expect_contains: '',
      expect_regex: '',
      timeout_ms: 3000,
    })
    expandedTCPCheckerStepIndex.value = tcpStandardDraft.steps.length - 1
  }

  function removeTCPCheckerStep(index: number) {
    if (tcpStandardDraft.steps.length <= 1) return
    tcpStandardDraft.steps.splice(index, 1)
    if (expandedTCPCheckerStepIndex.value === null) return
    expandedTCPCheckerStepIndex.value = Math.min(
      expandedTCPCheckerStepIndex.value,
      tcpStandardDraft.steps.length - 1
    )
  }

  function toggleTCPCheckerStep(index: number) {
    expandedTCPCheckerStepIndex.value =
      expandedTCPCheckerStepIndex.value === index ? null : index
  }

  function summarizeTCPCheckerStep(step: AWDTCPStandardDraft['steps'][number]): string {
    const send = step.send_template || step.send || step.send_hex
    const expect = step.expect_contains || step.expect_regex
    const parts = [
      send ? `发送 ${send}` : '',
      expect ? `期望 ${expect}` : '',
      Number.isInteger(step.timeout_ms) && step.timeout_ms > 0 ? `${step.timeout_ms}ms` : '',
    ].filter(Boolean)
    return parts.length > 0 ? parts.join(' · ') : '未配置收发规则'
  }

  return {
    addTCPCheckerStep,
    removeTCPCheckerStep,
    toggleTCPCheckerStep,
    summarizeTCPCheckerStep,
  }
}
