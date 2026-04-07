<script setup lang="ts">
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'

const props = withDefaults(
  defineProps<{
    page: number
    totalPages: number
    total: number
    totalLabel?: string
    disabled?: boolean
    prevButtonId?: string
    nextButtonId?: string
  }>(),
  {
    totalLabel: '',
    disabled: false,
    prevButtonId: '',
    nextButtonId: '',
  }
)

const emit = defineEmits<{
  changePage: [page: number]
}>()
</script>

<template>
  <PagePaginationControls
    :page="page"
    :total-pages="totalPages"
    :total="total"
    :total-label="totalLabel"
    :disabled="disabled"
    :prev-button-id="prevButtonId"
    :next-button-id="nextButtonId"
    :show-jump="true"
    @change-page="emit('changePage', $event)"
  />
</template>
