import { computed } from 'vue'
import { useRoute } from 'vue-router'

export function useContestAwdDefenseWorkbenchPage() {
  const route = useRoute()
  const contestId = computed(() => String(route.params.id ?? ''))
  const backLink = computed(() => ({
    name: 'ContestDetail',
    params: { id: contestId.value },
    query: { panel: 'challenges' },
  }))

  return {
    contestId,
    backLink,
  }
}
