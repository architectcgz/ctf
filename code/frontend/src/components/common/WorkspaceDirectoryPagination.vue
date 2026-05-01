<script setup lang="ts">
import { computed } from 'vue'

import PagePaginationControls from '@/components/common/PagePaginationControls.vue'

const props = withDefaults(
  defineProps<{
    page: number
    totalPages: number
    total: number
    totalLabel?: string
    disabled?: boolean
  }>(),
  {
    totalLabel: '',
    disabled: false,
  }
)

const emit = defineEmits<{
  changePage: [page: number]
}>()

const summaryLabel = computed(() => {
  const label = props.totalLabel.trim()
  if (!label) return ''
  return label.startsWith('共') ? label : `共 ${props.total} ${label}`
})
</script>

<template>
  <div
    v-if="total > 0"
    class="workspace-directory-pagination workspace-directory-pagination-shell"
  >
    <PagePaginationControls
      :page="page"
      :total-pages="totalPages"
      :total="total"
      :total-label="summaryLabel"
      :disabled="disabled"
      show-jump
      @change-page="emit('changePage', $event)"
    />
  </div>
</template>

<style scoped>
.workspace-directory-pagination-shell {
  margin-top: 1.5rem;
}
</style>
