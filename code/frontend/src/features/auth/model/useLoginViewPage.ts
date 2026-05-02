import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { sanitizeRedirectPath } from '@/router/guards'

export function useLoginViewPage() {
  const route = useRoute()
  const redirectTo = computed(() => sanitizeRedirectPath(route.query.redirect))

  return {
    redirectTo,
  }
}
