import { computed } from 'vue'
import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router'

export function useRouteQueryTransport() {
  const route = useRoute()
  const router = useRouter()

  const name = computed(() => route.name)
  const params = computed<Record<string, unknown>>(() => route.params)
  const query = computed<Record<string, unknown>>(() => route.query)

  async function replaceQuery(nextQuery: Record<string, unknown>): Promise<void> {
    await router.replace({
      query: nextQuery as LocationQueryRaw,
    })
  }

  return {
    name,
    params,
    query,
    replaceQuery,
  }
}
