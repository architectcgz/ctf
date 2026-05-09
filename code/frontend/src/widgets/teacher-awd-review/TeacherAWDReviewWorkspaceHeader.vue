<script setup lang="ts">
withDefaults(
  defineProps<{
    overline: string
    title: string
    headerClass?: string
    overlineClass?: string
  }>(),
  {
    headerClass: '',
    overlineClass: '',
  }
)
</script>

<template>
  <header
    class="teacher-topbar workspace-tab-heading"
    :class="headerClass"
  >
    <div class="teacher-heading workspace-tab-heading__main">
      <div
        class="workspace-overline"
        :class="overlineClass"
      >
        {{ overline }}
      </div>
      <h1 class="teacher-title workspace-page-title">
        {{ title }}
      </h1>
      <p class="teacher-copy workspace-page-copy">
        <slot name="description" />
      </p>
    </div>

    <div class="header-actions">
      <slot name="actions" />
    </div>
  </header>
</template>
