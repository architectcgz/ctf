<script setup lang="ts">
import type { Component } from 'vue'

interface SummaryItem {
  label: string
  value: string | number
  hint: string
  valueClass?: string
  icon?: Component
}

withDefaults(
  defineProps<{
    title: string
    items: SummaryItem[]
    summaryClass?: string
  }>(),
  {
    summaryClass: '',
  }
)
</script>

<template>
  <section
    class="teacher-summary teacher-summary--flat metric-panel-default-surface"
    :class="summaryClass"
  >
    <div class="teacher-summary-title">
      <slot name="title-prefix" />
      <span>{{ title }}</span>
      <slot name="title-suffix" />
    </div>

    <div class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
      <article
        v-for="item in items"
        :key="item.label"
        class="progress-card metric-panel-card"
      >
        <div class="progress-card-label metric-panel-label">
          <span>{{ item.label }}</span>
          <component :is="item.icon" v-if="item.icon" class="h-4 w-4" />
        </div>
        <div
          class="progress-card-value metric-panel-value"
          :class="item.valueClass"
        >
          {{ item.value }}
        </div>
        <div class="progress-card-hint metric-panel-helper">
          {{ item.hint }}
        </div>
      </article>
    </div>
  </section>
</template>
