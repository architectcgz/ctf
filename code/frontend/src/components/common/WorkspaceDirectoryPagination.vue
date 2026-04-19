<script setup lang="ts">
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'

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
</script>

<template>
  <div v-if="total > 0" class="workspace-directory-pagination workspace-directory-pagination-shell">
    <PlatformPaginationControls
      :page="page"
      :total-pages="totalPages"
      :total="total"
      :total-label="totalLabel"
      :disabled="disabled"
      @change-page="emit('changePage', $event)"
    />
  </div>
</template>

<style scoped>
.workspace-directory-pagination-shell {
  margin-top: 1.5rem;
}
</style>
