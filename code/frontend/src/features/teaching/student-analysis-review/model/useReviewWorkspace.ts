import { computed, ref } from 'vue'

import {
  getStudentAttackSessions,
  getStudentEvidence,
} from '@/api/teaching'
import type {
  AttackSessionQuery,
  AttackSessionResponseData,
  StudentEvidenceData,
  StudentEvidenceQuery,
} from '@/api/contracts'

export function useReviewWorkspace() {
  const evidence = ref<StudentEvidenceData | null>(null)
  const attackSessions = ref<AttackSessionResponseData | null>(null)
  const reviewChallengeOptions = ref<Array<{ value: string; label: string }>>([])
  const reviewWorkspaceLoading = ref(false)

  const sessionQuery = ref<AttackSessionQuery>({
    with_events: true,
    limit: 20,
    offset: 0,
  })

  const evidenceChallengeId = computed(() => sessionQuery.value.challenge_id)

  function reset(): void {
    evidence.value = null
    attackSessions.value = null
    reviewChallengeOptions.value = []
  }

  function syncChallengeOptions(): void {
    const nextOptions = new Map(reviewChallengeOptions.value.map((item) => [item.value, item.label]))

    evidence.value?.events.forEach((event) => {
      if (!event.challenge_id) return
      nextOptions.set(event.challenge_id, event.title || `题目 ${event.challenge_id}`)
    })

    attackSessions.value?.sessions.forEach((session) => {
      if (!session.challenge_id) return
      nextOptions.set(session.challenge_id, session.title || `题目 ${session.challenge_id}`)
    })

    reviewChallengeOptions.value = Array.from(nextOptions.entries())
      .sort((left, right) => left[0].localeCompare(right[0], 'zh-CN', { numeric: true }))
      .map(([value, label]) => ({ value, label }))
  }

  async function loadAttackSessions(studentId: string): Promise<void> {
    if (!studentId) {
      attackSessions.value = null
      return
    }

    reviewWorkspaceLoading.value = true
    try {
      attackSessions.value = await getStudentAttackSessions(studentId, sessionQuery.value)
      syncChallengeOptions()
    } finally {
      reviewWorkspaceLoading.value = false
    }
  }

  async function load(studentId: string): Promise<void> {
    if (!studentId) {
      reset()
      return
    }

    reviewWorkspaceLoading.value = true
    try {
      const evidenceQuery: StudentEvidenceQuery = evidenceChallengeId.value
        ? {
            challenge_id: evidenceChallengeId.value,
          }
        : {}
      evidence.value = await getStudentEvidence(studentId, evidenceQuery)
      syncChallengeOptions()
    } finally {
      reviewWorkspaceLoading.value = false
    }

    await loadAttackSessions(studentId)
  }

  function setSessionQuery(nextQuery: Partial<AttackSessionQuery>): void {
    sessionQuery.value = {
      ...sessionQuery.value,
      ...nextQuery,
    }
  }

  return {
    evidence,
    attackSessions,
    reviewChallengeOptions,
    reviewWorkspaceLoading,
    sessionQuery,
    loadReviewWorkspace: load,
    reloadAttackSessions: loadAttackSessions,
    resetReviewWorkspace: reset,
    setSessionQuery,
  }
}
