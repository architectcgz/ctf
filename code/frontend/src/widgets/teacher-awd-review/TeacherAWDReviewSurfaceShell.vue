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
      class="teacher-hero teacher-surface-hero flex min-h-full flex-1 flex-col border px-6 py-6 md:px-8"
      :class="sectionClass"
    >
      <slot />
    </section>
  </div>
</template>
