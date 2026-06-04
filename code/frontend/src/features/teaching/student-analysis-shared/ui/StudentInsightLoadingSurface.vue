<template>
  <div class="student-insight-loading-surface student-insight-glass-surface" v-bind="$attrs">
    <slot />
  </div>
</template>

<style src="./studentInsightSurface.css"></style>

<script setup lang="ts">
defineOptions({
  inheritAttrs: false,
})
</script>
