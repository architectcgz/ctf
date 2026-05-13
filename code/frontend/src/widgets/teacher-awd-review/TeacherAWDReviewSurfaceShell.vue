<script setup lang="ts">
withDefaults(
  defineProps<{
    sectionClass?: string
  }>(),
  {
    sectionClass: '',
  }
)
</script>

<template>
  <div class="teacher-management-shell teacher-surface workspace-shell flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero awd-review-surface-shell__section flex min-h-full flex-1 flex-col border"
      :class="sectionClass"
    >
      <slot />
    </section>
  </div>
</template>

<style scoped>
.awd-review-surface-shell__section {
  padding:
    var(--workspace-content-start-padding-top, var(--space-workspace-content-start-padding-top))
    var(--workspace-side-padding, var(--space-workspace-side-padding))
    var(--workspace-content-padding, var(--space-workspace-content-padding));
}
</style>
