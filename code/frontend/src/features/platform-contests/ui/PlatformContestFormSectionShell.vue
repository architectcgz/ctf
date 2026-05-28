<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  title: string
  description: string
  icon: Component
}>()
</script>

<template>
  <section class="contest-form-section">
    <div class="contest-form-section__info">
      <div class="contest-form-section__icon">
        <component :is="icon" class="contest-form-section__icon-glyph" />
      </div>
      <h3 class="list-heading__title">
        {{ title }}
      </h3>
      <p class="contest-form-section__description">
        {{ description }}
      </p>
    </div>

    <div class="contest-form-section__content">
      <slot />
    </div>
  </section>
</template>
